package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"git.ecadlabs.com/ecad/auth/authenticator/postgresql"
	"git.ecadlabs.com/ecad/auth/authenticator/postgresql/migrations"
	"git.ecadlabs.com/ecad/auth/router"
	"github.com/dgrijalva/jwt-go"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

const (
	testJWTSecret    = "09f67dc4-8fea-4d97-9cc0-bf674f5ec418"
	testPassword     = "password"
	testEmail        = "john_doe@example.com"
	testUserIDString = "ce9285b3-78e7-498d-b9a8-3b2b2fcfea67"
)

var (
	dbURL = flag.String("db", "postgres://localhost/auth?connect_timeout=10&sslmode=disable", "PostgreSQL server URL")
)

func init() {
	flag.Parse()
}

func createTestUser(url string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(testPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	id, _ := uuid.FromString(testUserIDString)
	user := postgresql.User{
		Id:           id,
		Email:        testEmail,
		PasswordHash: hash,
		FirstName:    "John",
		LastName:     "Doe",
	}

	db, err := sqlx.Open("postgres", url)
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.NamedExec("INSERT INTO users (id, email, password_hash, first_name, last_name) VALUES (:id, :email, :password_hash, :first_name, :last_name)", &user)
	// Ignore existed
	if pe, ok := err.(*pq.Error); !ok || pe.Code != "23505" {
		return err
	}

	return nil
}

func testLoginRequest(t *testing.T, server *httptest.Server, req *http.Request) {
	resp, err := server.Client().Do(req)

	if err != nil {
		t.Error(err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Error(resp.Status)
		return
	}

	var response struct {
		Token string `json:"token"`
	}

	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&response); err != nil {
		t.Error(err)
		return
	}

	tok, err := jwt.Parse(response.Token, func(t *jwt.Token) (interface{}, error) {
		return []byte([]byte(testJWTSecret)), nil
	})
	if err != nil {
		t.Error(err)
		return
	}

	if !tok.Valid || tok.Header["alg"] != jwtSigningMethod.Alg() {
		t.Errorf("alg: %s != %s\n", tok.Header["alg"], jwtSigningMethod.Alg())
		return
	}

	if c := tok.Claims.(jwt.MapClaims); c["sub"] != testUserIDString {
		t.Errorf("sub: %s != %s\n", c["sub"], testUserIDString)
	}

	t.Run("TestRefresh", func(t *testing.T) {
		req, err := http.NewRequest("GET", server.URL+"/refresh", nil)
		if err != nil {
			t.Error(err)
			return
		}

		req.Header.Set("Authorization", "Bearer "+response.Token)

		resp, err := server.Client().Do(req)
		if err != nil {
			t.Error(err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Error(resp.Status)
			return
		}
	})
}

func loginBody(name, password string) ([]byte, error) {
	request := struct {
		Name     string `json:"name"`
		Password string `json:"password"`
	}{
		Name:     name,
		Password: password,
	}

	body, err := json.Marshal(&request)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func TestPostLogin(t *testing.T) {
	if err := migrations.Migrate(*dbURL); err != nil {
		t.Error(err)
		return
	}

	if err := createTestUser(*dbURL); err != nil {
		t.Error(err)
		return
	}

	conf := router.Config{
		PostgresURL:      *dbURL,
		JWTSecret:        []byte(testJWTSecret),
		JWTSigningMethod: jwtSigningMethod,
	}

	r, err := conf.Handlers()
	if err != nil {
		t.Error(err)
		return
	}

	server := httptest.NewServer(r.Login)
	defer server.Close()

	body, err := loginBody(testEmail, testPassword)
	if err != nil {
		t.Error(err)
		return
	}

	req, err := http.NewRequest("POST", server.URL+"/login", bytes.NewReader(body))
	if err != nil {
		t.Error(err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	testLoginRequest(t, server, req)
}

func TestGetLogin(t *testing.T) {
	if err := migrations.Migrate(*dbURL); err != nil {
		t.Error(err)
		return
	}

	if err := createTestUser(*dbURL); err != nil {
		t.Error(err)
		return
	}

	conf := router.Config{
		PostgresURL:      *dbURL,
		JWTSecret:        []byte(testJWTSecret),
		JWTSigningMethod: jwtSigningMethod,
	}

	r, err := conf.Handlers()
	if err != nil {
		t.Error(err)
		return
	}

	server := httptest.NewServer(r.Login)
	defer server.Close()

	req, err := http.NewRequest("GET", server.URL+"/login", nil)
	if err != nil {
		t.Error(err)
		return
	}
	req.SetBasicAuth(testEmail, testPassword)

	testLoginRequest(t, server, req)
}

func TestLoginInvalidUser(t *testing.T) {
	if err := migrations.Migrate(*dbURL); err != nil {
		t.Error(err)
		return
	}

	if err := createTestUser(*dbURL); err != nil {
		t.Error(err)
		return
	}

	conf := router.Config{
		PostgresURL:      *dbURL,
		JWTSecret:        []byte(testJWTSecret),
		JWTSigningMethod: jwtSigningMethod,
	}

	r, err := conf.Handlers()
	if err != nil {
		t.Error(err)
		return
	}

	server := httptest.NewServer(r.Login)
	defer server.Close()

	body, err := loginBody("__dummy__", testPassword)
	if err != nil {
		t.Error(err)
		return
	}

	req, err := http.NewRequest("POST", server.URL+"/login", bytes.NewReader(body))
	if err != nil {
		t.Error(err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := server.Client().Do(req)
	if err != nil {
		t.Error(err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusForbidden {
		t.Error(resp.Status)
	}
}

func TestLoginInvalidPassword(t *testing.T) {
	if err := migrations.Migrate(*dbURL); err != nil {
		t.Error(err)
		return
	}

	if err := createTestUser(*dbURL); err != nil {
		t.Error(err)
		return
	}

	conf := router.Config{
		PostgresURL:      *dbURL,
		JWTSecret:        []byte(testJWTSecret),
		JWTSigningMethod: jwtSigningMethod,
	}

	r, err := conf.Handlers()
	if err != nil {
		t.Error(err)
		return
	}

	server := httptest.NewServer(r.Login)
	defer server.Close()

	body, err := loginBody(testEmail, "__dummy__")
	if err != nil {
		t.Error(err)
		return
	}

	req, err := http.NewRequest("POST", server.URL+"/login", bytes.NewReader(body))
	if err != nil {
		t.Error(err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := server.Client().Do(req)
	if err != nil {
		t.Error(err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusForbidden {
		t.Error(resp.Status)
	}
}

func TestHealthOK(t *testing.T) {
	conf := router.Config{
		PostgresURL:      *dbURL,
		JWTSecret:        []byte(testJWTSecret),
		JWTSigningMethod: jwtSigningMethod,
	}

	r, err := conf.Handlers()
	if err != nil {
		t.Error(err)
		return
	}

	// Test health server
	server := httptest.NewServer(r.Health)
	defer server.Close()

	resp, err := server.Client().Get(server.URL + "/healthz")
	if err != nil {
		t.Error(err)
		return
	}
	defer resp.Body.Close()

	var response struct {
		IsAlive bool   `json:"is_alive"`
		Error   string `json:"error"`
	}

	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&response); err != nil {
		t.Error(err)
		return
	}

	if resp.StatusCode != http.StatusOK || !response.IsAlive {
		t.Error(response.Error)
	}
}

func TestHealthNotOK(t *testing.T) {
	conf := router.Config{
		PostgresURL:      "postgres://8.8.8.8/auth?connect_timeout=1",
		JWTSecret:        []byte(testJWTSecret),
		JWTSigningMethod: jwtSigningMethod,
		AuthTimeout:      1 * time.Second,
	}

	r, err := conf.Handlers()
	if err != nil {
		t.Error(err)
		return
	}

	// Test health server
	server := httptest.NewServer(r.Health)
	defer server.Close()

	resp, err := server.Client().Get(server.URL + "/healthz")
	if err != nil {
		t.Error(err)
		return
	}
	defer resp.Body.Close()

	var response struct {
		IsAlive bool   `json:"is_alive"`
		Error   string `json:"error"`
	}

	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&response); err != nil {
		t.Error(err)
		return
	}

	if resp.StatusCode != http.StatusServiceUnavailable || response.IsAlive {
		t.Error(resp.StatusCode)
	}
}

func TestLoginTimeout(t *testing.T) {
	conf := router.Config{
		PostgresURL:      "postgres://8.8.8.8/auth?connect_timeout=1",
		JWTSecret:        []byte(testJWTSecret),
		JWTSigningMethod: jwtSigningMethod,
		AuthTimeout:      1 * time.Second,
	}

	r, err := conf.Handlers()
	if err != nil {
		t.Error(err)
		return
	}

	server := httptest.NewServer(r.Login)
	defer server.Close()

	body, err := loginBody(testEmail, testPassword)
	if err != nil {
		t.Error(err)
		return
	}

	req, err := http.NewRequest("POST", server.URL+"/login", bytes.NewReader(body))
	if err != nil {
		t.Error(err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := server.Client().Do(req)
	if err != nil {
		t.Error(err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Error(resp.Status)
	}
}

func TestVersion(t *testing.T) {
	conf := router.Config{
		PostgresURL:      *dbURL,
		JWTSecret:        []byte(testJWTSecret),
		JWTSigningMethod: jwtSigningMethod,
	}

	r, err := conf.Handlers()
	if err != nil {
		t.Error(err)
		return
	}

	server := httptest.NewServer(r.Login)
	defer server.Close()

	resp, err := server.Client().Get(server.URL + "/version")
	if err != nil {
		t.Error(err)
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Error(resp.StatusCode)
	}

	var response struct {
		Version string `json:"version"`
	}

	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&response); err != nil {
		t.Error(err)
		return
	}

	if response.Version == "" {
		t.Error("Empty version")
	}
}
