package intergationtesting

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"git.ecadlabs.com/ecad/auth/migrations"
	"git.ecadlabs.com/ecad/auth/notification"
	"git.ecadlabs.com/ecad/auth/service"
	"git.ecadlabs.com/ecad/auth/storage"
	"github.com/dgrijalva/jwt-go"
	"github.com/golang-migrate/migrate"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/satori/go.uuid"
)

const (
	testJWTSecret  = "09f67dc4-8fea-4d97-9cc0-bf674f5ec418"
	testPassword   = "admin"
	superUserEmail = "admin@admin"
	usersNum       = 10
)

var (
	dbURL = flag.String("db", "postgres://localhost/userstest?connect_timeout=10&sslmode=disable", "PostgreSQL server URL")
)

func init() {
	flag.Parse()
}

type testNotifier chan string

func (t testNotifier) InviteUser(ctx context.Context, d *notification.NotificationData) error {
	(chan string)(t) <- d.Token
	return nil
}

func (t testNotifier) PasswordReset(ctx context.Context, d *notification.NotificationData) error {
	return t.InviteUser(ctx, d)
}

func genTestUser(n int) *storage.User {
	return &storage.User{
		Email: fmt.Sprintf("user%d@example.com", n),
		Name:  fmt.Sprintf("Test User %d", n),
	}
}

func createUser(srv *httptest.Server, u *storage.User, token string, tokenCh chan string) (int, *storage.User, error) {
	buf, err := json.Marshal(u)
	if err != nil {
		return 0, nil, err
	}

	req, err := http.NewRequest("POST", srv.URL+"/users/", bytes.NewReader(buf))
	if err != nil {
		return 0, nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := srv.Client().Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		return resp.StatusCode, nil, nil
	}

	var res storage.User
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&res); err != nil {
		return 0, nil, err
	}

	resetToken := <-tokenCh

	resetRequest := struct {
		Token    string `json:"token"`
		Password string `json:"password"`
	}{
		Token:    resetToken,
		Password: testPassword,
	}

	buf, err = json.Marshal(&resetRequest)
	if err != nil {
		return 0, nil, err
	}

	req, err = http.NewRequest("POST", srv.URL+"/password_reset", bytes.NewReader(buf))
	req.Header.Set("Content-Type", "application/json")

	resp, err = srv.Client().Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		return resp.StatusCode, nil, nil
	}

	return resp.StatusCode, &res, nil
}

type tokenResponse struct {
	Token      string `json:"token"`
	RefreshURL string `json:"refresh"`
}

func doLogin(srv *httptest.Server, email, password string) (code int, token string, refresh string, err error) {
	req := struct {
		Name     string `json:"name"`
		Password string `json:"password"`
	}{
		Name:     email,
		Password: password,
	}

	buf, err := json.Marshal(&req)
	if err != nil {
		return 0, "", "", err
	}

	resp, err := srv.Client().Post(srv.URL+"/login", "application/json", bytes.NewReader(buf))
	if err != nil {
		return 0, "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		return resp.StatusCode, "", "", nil
	}

	var res tokenResponse

	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&res); err != nil {
		return 0, "", "", err
	}

	return resp.StatusCode, res.Token, res.RefreshURL, nil
}

func getUser(srv *httptest.Server, token string, uid uuid.UUID) (int, *storage.User, error) {
	req, err := http.NewRequest("GET", srv.URL+"/users/"+uid.String(), nil)
	if err != nil {
		return 0, nil, err
	}

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := srv.Client().Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return resp.StatusCode, nil, nil
	}

	var u storage.User

	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&u); err != nil {
		return 0, nil, err
	}

	return resp.StatusCode, &u, nil
}

func getList(srv *httptest.Server, token string, query url.Values) (int, []*storage.User, error) {
	tmpUrl, err := url.Parse(srv.URL)
	if err != nil {
		return 0, nil, err
	}

	tmpUrl.Path = "/users/"
	tmpUrl.RawQuery = query.Encode()
	reqUrl := tmpUrl.String()

	result := make([]*storage.User, 0)

	for {
		req, err := http.NewRequest("GET", reqUrl, nil)
		if err != nil {
			return 0, nil, err
		}

		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}

		resp, err := srv.Client().Do(req)
		if err != nil {
			return 0, nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusNoContent {
			break
		}

		if resp.StatusCode != http.StatusOK {
			return resp.StatusCode, nil, nil
		}

		var res struct {
			Value      []*storage.User `json:"value"`
			TotalCount int             `json:"total_count"`
			Next       string          `json:"next"`
		}

		dec := json.NewDecoder(resp.Body)
		if err := dec.Decode(&res); err != nil {
			return 0, nil, err
		}

		resp.Body.Close()

		if len(res.Value) == 0 {
			break
		}

		reqUrl = res.Next
		result = append(result, res.Value...)
	}

	return http.StatusOK, result, nil
}

func TestService(t *testing.T) {
	// Clear everything
	db, err := sqlx.Open("postgres", *dbURL)
	if err != nil {
		t.Error(err)
		return
	}
	defer db.Close()

	_, err = db.Exec("DROP TABLE IF EXISTS schema_migrations, users, roles, log, bootstrap")
	if err != nil {
		t.Error(err)
		return
	}

	// Migrate
	m, err := migrations.New(*dbURL)
	if err != nil {
		t.Error(err)
		return
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		t.Error(err)
		return
	}

	tokenCh := make(chan string, 10)

	// Create test server
	var srv *httptest.Server

	config := service.Config{
		BaseURLFunc:      func() string { return srv.URL },
		JWTSecret:        testJWTSecret,
		SessionMaxAge:    259200,
		ResetTokenMaxAge: 259200,
		PostgresURL:      *dbURL,
		DBTimeout:        10,
		Notifier:         testNotifier(tokenCh),
	}

	svc, err := config.New()
	if err != nil {
		t.Error(err)
		return
	}

	// Bootstrap
	superuser, err := svc.Bootstrap()
	if err != nil {
		t.Error(err)
		return
	}

	srv = httptest.NewServer(svc.APIHandler())
	defer srv.Close()

	// Login as superuser
	code, token, _, err := doLogin(srv, superUserEmail, testPassword)
	if err != nil {
		t.Error(err)
		return
	}

	//fmt.Println(token)

	if code != http.StatusOK {
		t.Error(code)
		return
	}

	usersList := []*storage.User{superuser}

	// Create other users
	for i := 0; i < usersNum; i++ {
		u := genTestUser(i)

		code, res, err := createUser(srv, u, token, tokenCh)
		if err != nil {
			t.Error(err)
			return
		}

		if code/100 != 2 {
			t.Error(code)
			return
		}

		usersList = append(usersList, res)
	}

	// Run all other tests in parallel
	t.Run("TestRegularUser", func(t *testing.T) {
		code, token, refresh, err := doLogin(srv, "user0@example.com", testPassword)
		if err != nil {
			t.Error(err)
			return
		}

		if code != http.StatusOK {
			t.Error(code)
			return
		}

		// Test refresh
		req, err := http.NewRequest("GET", refresh, nil)
		if err != nil {
			t.Error(err)
			return
		}
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := srv.Client().Do(req)
		if err != nil {
			t.Error(err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Error(resp.StatusCode)
			return
		}

		var res tokenResponse

		dec := json.NewDecoder(resp.Body)
		if err := dec.Decode(&res); err != nil {
			t.Error(err)
			return
		}

		tok, err := jwt.Parse(res.Token, func(t *jwt.Token) (interface{}, error) {
			return []byte([]byte(testJWTSecret)), nil
		})
		if err != nil {
			t.Error(err)
			return
		}

		if !tok.Valid || tok.Header["alg"] != service.JWTSigningMethod.Alg() {
			t.Errorf("alg: %s != %s\n", tok.Header["alg"], service.JWTSigningMethod.Alg())
			return
		}

		sub, ok := tok.Claims.(jwt.MapClaims)["sub"].(string)
		if !ok || sub == "" {
			t.Errorf("Empty sub")
			return
		}

		uid, err := uuid.FromString(sub)
		if err != nil {
			t.Error(err)
			return
		}

		t.Run("TestGetSelf", func(t *testing.T) {
			code, _, err := getUser(srv, res.Token, uid)
			if err != nil {
				t.Error(err)
				return
			}

			if code != http.StatusOK {
				t.Error(code)
				return
			}
		})

		t.Run("TestGetAllRegular", func(t *testing.T) {
			for _, u := range usersList {
				code, _, err := getUser(srv, token, u.ID)
				if err != nil {
					t.Error(err)
					return
				}

				if (u.ID == uid && code != http.StatusOK) || (u.ID != uid && code != http.StatusForbidden) {
					t.Error(code)
					return
				}
			}
		})

		t.Run("TestGetListRegular", func(t *testing.T) {
			code, _, err := getList(srv, token, url.Values{})
			if err != nil {
				t.Error(err)
				return
			}

			if code != http.StatusForbidden {
				t.Error(code)
				return
			}
		})
	})

	t.Run("TestSupesUser", func(t *testing.T) {
		code, token, _, err := doLogin(srv, superUserEmail, testPassword)
		if err != nil {
			t.Error(err)
			return
		}

		if code != http.StatusOK {
			t.Error(code)
			return
		}

		t.Run("TestGetAllSuper", func(t *testing.T) {
			for _, u := range usersList {
				code, _, err := getUser(srv, token, u.ID)
				if err != nil {
					t.Error(err)
					return
				}

				if code != http.StatusOK {
					t.Error(code)
					return
				}
			}
		})

		t.Run("TestGetListSuper", func(t *testing.T) {
			// Sort ASC
			code, listAsc, err := getList(srv, token, url.Values{
				"limit": []string{"2"},
				"order": []string{"asc"},
			})
			if err != nil {
				t.Error(err)
				return
			}

			if code != http.StatusOK {
				t.Error(code)
				return
			}

			code, listDesc, err := getList(srv, token, url.Values{
				"limit": []string{"2"},
				"order": []string{"desc"},
			})
			if err != nil {
				t.Error(err)
				return
			}

			if code != http.StatusOK {
				t.Error(code)
				return
			}

			if len(listAsc) != len(listDesc) {
				t.Errorf("%d != %d", len(listAsc), len(listDesc))
				return
			}

			for i, u := range listAsc {
				if u.ID != listDesc[len(listDesc)-1-i].ID {
					t.Errorf("Sort error")
					return
				}
			}
		})
	})

	t.Run("TestWrongUserNameLogin", func(t *testing.T) {
		code, _, _, err := doLogin(srv, "_dummy_@domain.com", "_dummy_")
		if err != nil {
			t.Error(err)
			return
		}

		if code != http.StatusUnauthorized {
			t.Error(code)
			return
		}
	})

	t.Run("TestWrongPasswordLogin", func(t *testing.T) {
		code, _, _, err := doLogin(srv, "user0@example.com", "_dummy_")
		if err != nil {
			t.Error(err)
			return
		}

		if code != http.StatusUnauthorized {
			t.Error(code)
			return
		}
	})
}
