package service

import (
	"net/http"
	"net/url"
	"time"

	"git.ecadlabs.com/ecad/auth/handlers"
	"git.ecadlabs.com/ecad/auth/middleware"
	"git.ecadlabs.com/ecad/auth/users"
	"github.com/auth0/go-jwt-middleware"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"strconv"
)

const (
	version               = "0.0.1"
	defaultConnectTimeout = 10
)

var jwtSigningMethod = jwt.SigningMethodHS256

type Service struct {
	config  Config
	storage *users.Storage
}

func (c *Config) New() (*Service, error) {
	url, err := url.Parse(c.PostgresURL)
	if err != nil {
		return nil, err
	}

	// Set connection timeout
	q := url.Query()
	if _, ok := q["connect_timeout"]; !ok {
		q["connect_timeout"] = []string{strconv.FormatInt(defaultConnectTimeout, 10)}
	}
	url.RawQuery = q.Encode()

	db, err := sqlx.Open("postgres", url.String())
	if err != nil {
		return nil, err
	}

	return &Service{
		config:  *c,
		storage: &users.Storage{DB: db},
	}, nil
}

func (s *Service) APIHandler() http.Handler {
	tokenHandler := handlers.TokenHandler{
		Storage:       s.storage,
		Timeout:       time.Duration(s.config.DBTimeout) * time.Second,
		SessionMaxAge: time.Duration(s.config.SessionMaxAge) * time.Second,
		JWTSecretGetter: func() ([]byte, error) {
			return []byte(s.config.JWTSecret), nil
		},
		JWTSigningMethod: jwtSigningMethod,
		Namespace:        s.config.BaseURL,
	}

	if s.config.BaseURL != "" {
		tokenHandler.RefreshURL = s.config.BaseURL + "/refresh"
	}

	usersHandler := handlers.Users{
		Storage:   s.storage,
		BaseURL:   s.config.BaseURL + "/users/",
		Namespace: s.config.BaseURL,
		Timeout:   time.Duration(s.config.DBTimeout) * time.Second,
	}

	jwtOptions := jwtmiddleware.Options{
		ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) { return []byte(s.config.JWTSecret), nil },
		SigningMethod:       jwtSigningMethod,
		UserProperty:        handlers.TokenContextKey,
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err string) {
			handlers.JSONError(w, err, http.StatusUnauthorized)
		},
	}
	jwtMiddleware := jwtmiddleware.New(jwtOptions)

	jwtOptions.CredentialsOptional = true
	jwtMiddlewareOptionalAuth := jwtmiddleware.New(jwtOptions)

	m := mux.NewRouter()

	m.Use((&middleware.Logging{}).Handler)
	m.Use(middleware.NewPrometheus().Handler)
	m.Use((&middleware.Recover{}).Handler)

	// Login API
	m.Methods("GET", "POST").Path("/login").HandlerFunc(tokenHandler.Login)
	m.Methods("GET").Path("/refresh").Handler(jwtMiddleware.Handler(http.HandlerFunc(tokenHandler.Refresh)))

	// Users API
	m.Methods("POST").Path("/users/").Handler(jwtMiddlewareOptionalAuth.Handler(http.HandlerFunc(usersHandler.NewUser)))

	umux := m.PathPrefix("/users").Subrouter()
	umux.Use(jwtMiddleware.Handler)

	umux.Methods("GET").Path("/").HandlerFunc(usersHandler.GetUsers)
	umux.Methods("GET").Path("/{id}").HandlerFunc(usersHandler.GetUser)
	umux.Methods("PATCH").Path("/{id}").HandlerFunc(usersHandler.PatchUser)
	umux.Methods("DELETE").Path("/{id}").HandlerFunc(usersHandler.DeleteUser)

	// Miscellaneous
	m.Methods("GET").Path("/version").Handler(handlers.VersionHandler(version))
	m.Methods("GET").Path("/metrics").Handler(promhttp.Handler())

	m.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlers.JSONError(w, "Resource not found", http.StatusNotFound)
	})

	return m
}

func (s *Service) HealthHandler() http.Handler {
	// Health handler
	healthMon := &handlers.HealthMonitor{
		Pinger:  s.storage,
		Timeout: time.Duration(s.config.DBTimeout) * time.Second,
	}

	// Health router
	m := mux.NewRouter()
	m.Use((&middleware.Logging{}).Handler)
	m.Use((&middleware.Recover{}).Handler)

	m.Methods("GET").Path("/healthz").Handler(healthMon)

	return m
}
