package router

import (
	"net/http"
	"net/url"
	"time"

	pgauth "git.ecadlabs.com/ecad/auth/authenticator/postgresql"
	"git.ecadlabs.com/ecad/auth/handlers"
	"git.ecadlabs.com/ecad/auth/middleware"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"strconv"
)

const (
	version               = "0.0.1"
	claimsNamespace       = "http://git.ecadlabs.com/ecad/auth"
	defaultConnectTimeout = 10
	defaultAuthTimeout    = 10 * time.Second
	defaultSessionMaxAge  = 72 * time.Hour
)

// Simplify testing
type Config struct {
	PostgresURL      string
	JWTSecret        []byte
	AuthTimeout      time.Duration
	JWTSigningMethod jwt.SigningMethod
}

func (c *Config) authTimeout() time.Duration {
	if c.AuthTimeout != 0 {
		return c.AuthTimeout
	}
	return defaultAuthTimeout
}

type Handlers struct {
	Login  http.Handler
	Health http.Handler
}

func (c *Config) Handlers() (r *Handlers, err error) {
	if c.JWTSigningMethod == nil {
		panic("JWTSigningMethod == nil")
	}

	if len(c.JWTSecret) == 0 {
		panic("len(c.JWTSecret) == 0")
	}

	url, err := url.Parse(c.PostgresURL)
	if err != nil {
		return
	}

	// Set connection timeout
	q := url.Query()
	if _, ok := q["connect_timeout"]; !ok {
		q["connect_timeout"] = []string{strconv.FormatInt(defaultConnectTimeout, 10)}
	}
	url.RawQuery = q.Encode()

	// Radius client
	pgAuth, err := pgauth.New("postgres", url.String(), claimsNamespace)
	if err != nil {
		return
	}

	// Health handler
	healthMon := &handlers.HealthMonitor{
		Pinger:  pgAuth,
		Timeout: c.authTimeout(),
	}

	// Health router
	hmux := mux.NewRouter()
	hmux.Use((&middleware.Logging{}).Handler)
	hmux.Use((&middleware.Recover{}).Handler)
	hmux.Methods("GET").Path("/healthz").Handler(healthMon)

	r = &Handlers{
		Health: hmux,
	}

	// Login handler
	login := &handlers.Login{
		SessionMaxAge: defaultSessionMaxAge,
		Authenticator: pgAuth,
		JWTSecretGetter: func() ([]byte, error) {
			return c.JWTSecret, nil
		},
		JWTSigningMethod: c.JWTSigningMethod,
		Namespace:        claimsNamespace,
		AuthTimeout:      c.authTimeout(),
	}

	prometheusMiddleware := middleware.NewPrometheus()

	m := mux.NewRouter()
	m.Use((&middleware.Logging{}).Handler)
	m.Use(prometheusMiddleware.Handler)
	m.Use((&middleware.Recover{}).Handler)
	m.Methods("GET", "POST").Path("/login").Handler(login)
	m.Methods("GET").Path("/version").Handler(handlers.VersionHandler(version))
	m.Path("/metrics").Handler(promhttp.Handler())

	r.Login = m

	return
}
