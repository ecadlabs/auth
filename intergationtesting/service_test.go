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

	"github.com/dgrijalva/jwt-go"
	"github.com/ecadlabs/auth/migrations"
	"github.com/ecadlabs/auth/notification"
	"github.com/ecadlabs/auth/rbac"
	"github.com/ecadlabs/auth/service"
	"github.com/ecadlabs/auth/storage"
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
	dbURL = flag.String("db", "postgres://auth:auth@localhost/userstest?connect_timeout=10&sslmode=disable", "PostgreSQL server URL")
)

func init() {
	flag.Parse()
}

type testNotifier chan string

func (t testNotifier) Notify(ctx context.Context, tpl string, d *notification.NotificationData) error {
	(chan string)(t) <- d.Token
	return nil
}

func genTestEmail(n int) string {
	return fmt.Sprintf("test+χρήστης%d@екзампл.ком", n)
}

func genTestName(n int) string {
	return fmt.Sprintf("Test Тест 日本語 ☺☻☹ %d", n)
}

func genTestUser(n int) *storage.CreateUser {
	return &storage.CreateUser{
		Email: genTestEmail(n),
		Name:  genTestName(n),
		Roles: storage.Roles{
			"regular": struct{}{},
		},
	}
}

func createUser(srv *httptest.Server, u *storage.CreateUser, token string, tokenCh chan string) (int, *storage.User, error) {
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

func doLogin(srv *httptest.Server, email, password string, tenantId *uuid.UUID) (code int, token string, refresh string, err error) {
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

	var url = ""
	if tenantId == nil {
		url = srv.URL + "/login"
	} else {
		url = fmt.Sprintf(srv.URL+"/login/%s", *tenantId)
	}
	resp, err := srv.Client().Post(url, "application/json", bytes.NewReader(buf))
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

func deleteUser(srv *httptest.Server, token string, uid uuid.UUID) (int, error) {
	req, err := http.NewRequest("DELETE", srv.URL+"/users/"+uid.String(), nil)
	if err != nil {
		return 0, err
	}

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := srv.Client().Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	return resp.StatusCode, nil
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

type tenantsAndUsers struct {
	Tenants []*storage.TenantModel
	Users   []*storage.User
}

func (t *tenantsAndUsers) GetUser(email string) *storage.User {
	for _, val := range t.Users {
		if val.Email == email {
			return val
		}
	}
	return nil
}

func (t *tenantsAndUsers) GetTenantbyName(email string) *storage.TenantModel {
	for _, val := range t.Tenants {
		if val.Name == email {
			return val
		}
	}
	return nil
}

var res *tenantsAndUsers

func fetchTenantAndUsers(srv *httptest.Server, refresh bool) (*tenantsAndUsers, error) {
	if res == nil || refresh {
		code, token, _, err := doLogin(srv, superUserEmail, testPassword, nil)
		if err != nil {
			return nil, err
		}

		if code != http.StatusOK {
			return nil, err
		}

		_, list, err := getTenantList(srv, token, url.Values{})
		if err != nil {
			return nil, err
		}

		_, userList, err := getList(srv, token, url.Values{})
		if err != nil {
			return nil, err
		}

		res = &tenantsAndUsers{
			Users:   userList,
			Tenants: list,
		}
	}
	return res, nil
}

var testRBAC = rbac.StaticRBAC{
	Roles: map[string]*rbac.StaticRole{
		"admin": &rbac.StaticRole{
			RoleName:    "admin",
			Description: "A super user that has all access",
			Permissions: map[string]struct{}{
				"com.ecadlabs.users.delegate:admin":   struct{}{},
				"com.ecadlabs.users.delegate:noc":     struct{}{},
				"com.ecadlabs.users.delegate:owner":   struct{}{},
				"com.ecadlabs.users.delegate:regular": struct{}{},
				"com.ecadlabs.users.full_control":     struct{}{},
				"com.ecadlabs.tenants.full_control":   struct{}{},
			},
		},
		"owner": &rbac.StaticRole{
			RoleName:    "owner",
			Description: "Tenant owner",
			Permissions: map[string]struct{}{
				"com.ecadlabs.users.delegate:owner":   struct{}{},
				"com.ecadlabs.users.delegate:regular": struct{}{},
				"com.ecadlabs.tenants.read_owned":     struct{}{},
				"com.ecadlabs.tenants.write_owned":    struct{}{},
				"com.ecadlabs.users.read_self":        struct{}{},
				"com.ecadlabs.users.write_self":       struct{}{},
			},
		},
		"regular": &rbac.StaticRole{
			RoleName:    "regular",
			Description: "Regular member",
			Permissions: map[string]struct{}{
				"com.ecadlabs.users.read_self":  struct{}{},
				"com.ecadlabs.users.write_self": struct{}{},
			},
		},
	},
	Permissions: map[string]string{
		"com.ecadlabs.users.delegate:admin": "Assign `admin' role",
		"com.ecadlabs.users.delegate:noc":   "Assign `noc' role",
		"com.ecadlabs.users.delegate:owner": "Assign `owner' role",
		"com.ecadlabs.users.delegate:ops":   "Assign `ops' role",
		"com.ecadlabs.users.full_control":   "Allows user to manage all accounts",
		"com.ecadlabs.tenants.full_control": "Allows user to manage all tenants",
		"com.ecadlabs.users.read":           "Allows user to view users",
		"com.ecadlabs.users.read_logs":      "Allows user to access logs",
		"com.ecadlabs.users.read_self":      "Allows user to view their own user resource record",
		"com.ecadlabs.users.write":          "Allows user to create new users",
		"com.ecadlabs.users.write_self":     "Allows user to edit their own user resource record",
		"com.ecadlabs.tenants.read_self":    "Allows user to read their own tenant resource record",
		"com.ecadlabs.tenants.write_self":   "Allows user to write their own tenant resource record",
	},
}

func TestService(t *testing.T) {
	// Clear everything
	db, err := sqlx.Open("postgres", *dbURL)
	if err != nil {
		t.Error(err)
		return
	}
	defer db.Close()

	_, err = db.Exec(`DROP TABLE IF EXISTS schema_migrations, users, membership, tenants, roles, log, bootstrap`)
	_, err = db.Exec(`DROP TYPE IF EXISTS membership_type, membership_status`)
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
		BaseURLFunc:            func() string { return srv.URL },
		JWTSecret:              testJWTSecret,
		SessionMaxAge:          259200,
		ResetTokenMaxAge:       259200,
		TenantInviteMaxAge:     259200,
		EmailUpdateTokenMaxAge: 259200,
		PostgresURL:            *dbURL,
		DBTimeout:              10 * 60 * 60,
		Notifier:               testNotifier(tokenCh),
	}

	svc, err := service.New(&config, &testRBAC)
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
	code, token, _, err := doLogin(srv, superUserEmail, testPassword, nil)
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
		code, token, refresh, err := doLogin(srv, genTestEmail(0), testPassword, nil)
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

		t.Run("TestCreateTenant", func(t *testing.T) {
			model := createTenantModel{Name: "test"}
			code, _, err := createTenant(srv, &model, token)
			if err != nil {
				t.Error(err)
				return
			}

			if code != http.StatusForbidden {
				t.Error(code)
				return
			}
		})

		t.Run("TestListTenant", func(t *testing.T) {
			code, list, err := getTenantList(srv, token, url.Values{})
			if err != nil {
				t.Error(err)
				return
			}

			if len(list) != 1 {
				t.Error("Len is not one")
			}

			if code != http.StatusOK {
				t.Error(code)
				return
			}
		})
	})

	t.Run("TestSupesUser", func(t *testing.T) {
		code, token, _, err := doLogin(srv, superUserEmail, testPassword, nil)
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
				code, ret, err := getUser(srv, token, u.ID)
				if err != nil {
					t.Error(err)
					return
				}

				if code != http.StatusOK {
					t.Error(code)
					return
				}

				if ret.ID != u.ID {
					t.Errorf("%v != %v", ret.ID, u.ID)
				}

				if ret.Email != u.Email {
					t.Errorf("%v != %v", ret.Email, u.Email)
				}

				if ret.Name != u.Name {
					t.Errorf("%v != %v", ret.Name, u.Name)
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

		t.Run("TestCreateTenant", func(t *testing.T) {
			model := createTenantModel{Name: "test"}
			code, _, err := createTenant(srv, &model, token)
			if err != nil {
				t.Error(err)
				return
			}

			if code != http.StatusOK {
				t.Error(code)
				return
			}
		})

		t.Run("TestListTenant", func(t *testing.T) {
			code, list, err := getTenantList(srv, token, url.Values{})
			if err != nil {
				t.Error(err)
				return
			}

			if len(list) != 13 {
				t.Error("Len is not 13", len(list))
			}

			if code != http.StatusOK {
				t.Error(code)
				return
			}
		})
	})

	t.Run("InviteToTenant", func(t *testing.T) {
		results, err := fetchTenantAndUsers(srv, false)
		if err != nil {
			t.Error(err)
			return
		}

		firstTenant := results.GetTenantbyName("test")

		code, err = inviteTenant(srv, token, fmt.Sprintf("%s", firstTenant.ID), genTestEmail(0))

		if code != http.StatusNoContent {
			t.Error(code)
			return
		}

		if err != nil {
			t.Error(err)
			return
		}

		inviteToken := <-tokenCh
		fmt.Printf("%s", inviteToken)
		code, err = acceptInvite(srv, inviteToken)

		if code != http.StatusNoContent {
			t.Error(code)
			return
		}

		if err != nil {
			t.Error(err)
			return
		}
	})

	t.Run("New Member Should Not Be Able To Invite Other", func(t *testing.T) {
		results, err := fetchTenantAndUsers(srv, false)
		if err != nil {
			t.Error(err)
			return
		}

		firstTenant := results.GetTenantbyName("test")

		code, token, _, err = doLogin(srv, genTestEmail(0), testPassword, &firstTenant.ID)
		if err != nil {
			t.Error(err)
			return
		}

		if code != http.StatusOK {
			t.Error(code)
			return
		}

		code, err = inviteTenant(srv, token, fmt.Sprintf("%s", firstTenant.ID), genTestEmail(1))

		if code != http.StatusForbidden {
			t.Error(code)
			return
		}

		if err != nil {
			t.Error(err)
			return
		}
	})

	t.Run("Admin can patch any membership", func(t *testing.T) {
		results, err := fetchTenantAndUsers(srv, false)
		if err != nil {
			t.Error(err)
			return
		}

		firstTenant := results.GetTenantbyName("test")
		firstUser := results.GetUser(genTestEmail(0))

		if firstTenant == nil {
			t.Error("Tenant do not exists")
			return
		}

		if firstUser == nil {
			t.Error("User do not exists")
			return
		}

		code, token, _, err := doLogin(srv, superUserEmail, testPassword, nil)

		if code != http.StatusOK {
			t.Error(code)
			return
		}

		if err != nil {
			t.Error(code)
			return
		}

		code, err = patchMembership(srv, token, firstUser.ID, firstTenant.ID)

		if code != http.StatusOK {
			t.Error(code)
			return
		}

		if err != nil {
			t.Error(code)
			return
		}

		fmt.Printf("%v, %v", firstTenant.ID, firstUser.ID)
	})

	t.Run("OwnerShouldBeAbleToInviteInHisOwnTenant", func(t *testing.T) {
		results, err := fetchTenantAndUsers(srv, false)
		if err != nil {
			t.Error(err)
			return
		}

		firstTenant := results.GetTenantbyName(genTestEmail(0))

		if firstTenant == nil {
			t.Error("Tenant do not exists")
			return
		}

		code, token, _, err = doLogin(srv, genTestEmail(0), testPassword, &firstTenant.ID)
		if err != nil {
			t.Error(err)
			return
		}

		if code != http.StatusOK {
			t.Error(code)
			return
		}

		code, err = inviteTenant(srv, token, fmt.Sprintf("%s", firstTenant.ID), genTestEmail(1))

		if code != http.StatusNoContent {
			t.Error(code)
			return
		}

		if err != nil {
			t.Error(err)
			return
		}

		inviteToken := <-tokenCh
		fmt.Printf("%s", inviteToken)
		code, err = acceptInvite(srv, inviteToken)

		if code != http.StatusNoContent {
			t.Error(code)
			return
		}

		if err != nil {
			t.Error(err)
			return
		}
	})

	t.Run("Owner Should Not Be Able To Invite In Other Tenant", func(t *testing.T) {
		results, err := fetchTenantAndUsers(srv, false)
		if err != nil {
			t.Error(err)
			return
		}

		rootTenant := results.GetTenantbyName("admin@admin")
		tenantWithOwner := results.GetTenantbyName(genTestEmail(0))

		if tenantWithOwner == nil {
			t.Error("Tenant do not exists")
			return
		}

		code, token, _, err = doLogin(srv, genTestEmail(0), testPassword, &tenantWithOwner.ID)
		if err != nil {
			t.Error(err)
			return
		}

		if code != http.StatusOK {
			t.Error(code)
			return
		}
		code, err = inviteTenant(srv, token, fmt.Sprintf("%s", rootTenant.ID), genTestEmail(1))

		if code != http.StatusForbidden {
			t.Error(code)
			return
		}

		if err != nil {
			t.Error(err)
			return
		}
	})

	t.Run("Owner can't delegate role in other tenant", func(t *testing.T) {
		results, err := fetchTenantAndUsers(srv, false)

		if err != nil {
			t.Error(err)
			return
		}

		rootTenant := results.GetTenantbyName("admin@admin")
		tenantWithOwner := results.GetTenantbyName(genTestEmail(0))

		if tenantWithOwner == nil {
			t.Error("Tenant do not exists")
			return
		}
		code, token, _, err = doLogin(srv, genTestEmail(0), testPassword, &tenantWithOwner.ID)

		if code != http.StatusOK {
			t.Error(code)
			return
		}

		if err != nil {
			t.Error(err)
			return
		}

		code, err = patchMembership(srv, token, results.Users[0].ID, rootTenant.ID)

		if code != http.StatusForbidden {
			t.Error(code)
			return
		}

		if err != nil {
			t.Error(code)
			return
		}
	})

	t.Run("Regular user can't delete membership", func(t *testing.T) {
		results, err := fetchTenantAndUsers(srv, false)

		if err != nil {
			t.Error(err)
			return
		}

		tenantWithOwner := results.GetTenantbyName(genTestEmail(0))

		if tenantWithOwner == nil {
			t.Error("Tenant do not exists")
			return
		}
		code, token, _, err = doLogin(srv, genTestEmail(1), testPassword, &tenantWithOwner.ID)

		if code != http.StatusOK {
			t.Error(code)
			return
		}

		if err != nil {
			t.Error(err)
			return
		}

		user := results.GetUser(genTestEmail(0))
		code, err = DeleteMembership(srv, token, tenantWithOwner.ID, user.ID)

		if code != http.StatusForbidden {
			t.Error(code)
			return
		}

		if err != nil {
			t.Error(code)
			return
		}
	})

	t.Run("Owner user can delete membership", func(t *testing.T) {
		results, err := fetchTenantAndUsers(srv, false)

		if err != nil {
			t.Error(err)
			return
		}

		tenantWithOwner := results.GetTenantbyName(genTestEmail(0))

		if tenantWithOwner == nil {
			t.Error("Tenant do not exists")
			return
		}
		code, token, _, err = doLogin(srv, genTestEmail(0), testPassword, &tenantWithOwner.ID)

		if code != http.StatusOK {
			t.Error(code)
			return
		}

		if err != nil {
			t.Error(err)
			return
		}

		user := results.GetUser(genTestEmail(1))
		code, err = DeleteMembership(srv, token, tenantWithOwner.ID, user.ID)

		if code != http.StatusNoContent {
			t.Error(code)
			return
		}

		if err != nil {
			t.Error(code)
			return
		}
	})

	t.Run("Owner user can see all membership in tenant", func(t *testing.T) {
		results, err := fetchTenantAndUsers(srv, false)
		if err != nil {
			t.Error(err)
			return
		}
		tenantWithOwner := results.GetTenantbyName(genTestEmail(0))
		if tenantWithOwner == nil {
			t.Error("Tenant do not exists")
			return
		}
		code, token, _, err = doLogin(srv, genTestEmail(0), testPassword, &tenantWithOwner.ID)

		if code != http.StatusOK {
			t.Error(code)
			return
		}

		if err != nil {
			t.Error(err)
			return
		}

		_, list, err := getTenantMembershipsList(srv, token, tenantWithOwner.ID, url.Values{})
		if err != nil {
			t.Error(err)
			return
		}

		if len(list) != 1 {
			t.Error("Should return 1 membership", len(list))
		}
	})

	t.Run("Regular user can't see all member", func(t *testing.T) {
		results, err := fetchTenantAndUsers(srv, false)
		if err != nil {
			t.Error(err)
			return
		}

		firstTenant := results.GetTenantbyName(genTestEmail(0))
		if firstTenant == nil {
			t.Error("Tenant do not exists")
			return
		}

		code, token, _, err = doLogin(srv, genTestEmail(0), testPassword, &firstTenant.ID)
		if err != nil {
			t.Error(err)
			return
		}

		if code != http.StatusOK {
			t.Error(code)
			return
		}

		code, err = inviteTenant(srv, token, fmt.Sprintf("%s", firstTenant.ID), genTestEmail(4))

		if code != http.StatusNoContent {
			t.Error(code)
			return
		}

		if err != nil {
			t.Error(err)
			return
		}

		inviteToken := <-tokenCh
		fmt.Printf("%s", inviteToken)
		code, err = acceptInvite(srv, inviteToken)

		if code != http.StatusNoContent {
			t.Error(code)
			return
		}

		if err != nil {
			t.Error(err)
			return
		}

		code, token, _, err = doLogin(srv, genTestEmail(4), testPassword, &firstTenant.ID)

		if code != http.StatusOK {
			t.Error(code)
			return
		}

		if err != nil {
			t.Error(err)
			return
		}

		code, _, err = getTenantMembershipsList(srv, token, firstTenant.ID, url.Values{})
		if code != http.StatusForbidden {
			t.Error(code)
			return
		}

		if err != nil {
			t.Error(err)
			return
		}
	})

	t.Run("User should be able to see all his membership", func(t *testing.T) {
		results, err := fetchTenantAndUsers(srv, false)
		if err != nil {
			t.Error(err)
			return
		}
		tenantWithOwner := results.GetTenantbyName(genTestEmail(0))
		if tenantWithOwner == nil {
			t.Error("Tenant do not exists")
			return
		}
		code, token, _, err = doLogin(srv, genTestEmail(0), testPassword, &tenantWithOwner.ID)

		if code != http.StatusOK {
			t.Error(code)
			return
		}

		if err != nil {
			t.Error(err)
			return
		}

		code, list, err := getUserMembershipsList(srv, token, results.GetUser(genTestEmail(0)).ID, url.Values{})
		if err != nil {
			t.Error(err)
			return
		}

		if len(list) != 2 {
			t.Error("Should return 2 membership", len(list), code)
		}
	})

	t.Run("Regular user should not be able to see other user memberships", func(t *testing.T) {
		results, err := fetchTenantAndUsers(srv, false)
		if err != nil {
			t.Error(err)
			return
		}
		tenantWithOwner := results.GetTenantbyName(genTestEmail(0))
		if tenantWithOwner == nil {
			t.Error("Tenant do not exists")
			return
		}
		code, token, _, err = doLogin(srv, genTestEmail(0), testPassword, &tenantWithOwner.ID)

		if code != http.StatusOK {
			t.Error(code)
			return
		}

		if err != nil {
			t.Error(err)
			return
		}

		code, _, err = getUserMembershipsList(srv, token, results.GetUser(genTestEmail(1)).ID, url.Values{})
		if code != http.StatusForbidden {
			t.Error(code)
			return
		}

		if err != nil {
			t.Error(err)
			return
		}
	})

	t.Run("Invited user should not be able to login", func(t *testing.T) {
		results, err := fetchTenantAndUsers(srv, false)
		if err != nil {
			t.Error(err)
			return
		}
		tenantWithOwner := results.GetTenantbyName(genTestEmail(0))
		if tenantWithOwner == nil {
			t.Error("Tenant do not exists")
			return
		}
		code, err = inviteTenant(srv, token, fmt.Sprintf("%s", tenantWithOwner.ID), genTestEmail(3))

		if code != http.StatusNoContent {
			t.Error(code)
			return
		}

		if err != nil {
			t.Error(err)
			return
		}

		code, token, _, err = doLogin(srv, genTestEmail(3), testPassword, &tenantWithOwner.ID)

		if code != http.StatusForbidden {
			t.Error(code)
			return
		}

		if err != nil {
			t.Error(err)
			return
		}
	})

	t.Run("Delete user should archive orphan tenants", func(t *testing.T) {
		results, err := fetchTenantAndUsers(srv, false)
		if err != nil {
			t.Error(err)
			return
		}

		user := results.GetUser(genTestEmail(5))

		if user == nil {
			t.Error("User does not exists")
			return
		}

		code, token, _, err := doLogin(srv, superUserEmail, testPassword, nil)

		if code != http.StatusOK {
			t.Error(code)
			return
		}

		if err != nil {
			t.Error(err)
			return
		}

		code, err = deleteUser(srv, token, user.ID)

		if code != http.StatusNoContent {
			t.Error(code)
			return
		}

		if err != nil {
			t.Error(err)
			return
		}

		results, err = fetchTenantAndUsers(srv, true)
		if err != nil {
			t.Error(err)
			return
		}

		deletedTenant := results.GetTenantbyName(genTestEmail(5))

		if err != nil {
			t.Error(err)
			return
		}

		if deletedTenant != nil {
			t.Error("Tenant should have been archived", deletedTenant.Name, deletedTenant.Archived)
		}

	})

	t.Run("TestWrongUserNameLogin", func(t *testing.T) {
		code, _, _, err := doLogin(srv, "_dummy_@domain.com", "_dummy_", nil)
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
		code, _, _, err := doLogin(srv, "user0@example.com", "_dummy_", nil)
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
