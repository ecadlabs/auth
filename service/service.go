package service

import (
	"net/http"
	"net/url"
	"time"

	"database/sql"
	"git.ecadlabs.com/ecad/auth/handlers"
	"git.ecadlabs.com/ecad/auth/logger"
	"git.ecadlabs.com/ecad/auth/middleware"
	"git.ecadlabs.com/ecad/auth/notification"
	"git.ecadlabs.com/ecad/auth/users"
	"git.ecadlabs.com/ecad/auth/utils"
	"github.com/auth0/go-jwt-middleware"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"strconv"
)

const (
	version               = "0.0.1"
	defaultConnectTimeout = 10
)

var JWTSigningMethod = jwt.SigningMethodHS256

type Service struct {
	config  Config
	storage *users.Storage
	DB      *sql.DB
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

	db, err := sql.Open("postgres", url.String())
	if err != nil {
		return nil, err
	}

	return &Service{
		config:  *c,
		storage: &users.Storage{DB: sqlx.NewDb(db, "postgres")},
		DB:      db,
	}, nil
}

func (s *Service) APIHandler() http.Handler {
	baseURLFunc := s.config.GetBaseURLFunc()

	dbLogger := logrus.New()
	dbLogger.AddHook(&logger.Hook{
		DB: s.DB,
	})

	usersHandler := handlers.Users{
		Storage: s.storage,
		Timeout: time.Duration(s.config.DBTimeout) * time.Second,

		JWTSecretGetter: func() ([]byte, error) {
			return []byte(s.config.JWTSecret), nil
		},
		JWTSigningMethod: JWTSigningMethod,

		BaseURL:     baseURLFunc,
		UsersPath:   "/users/",
		RefreshPath: "/refresh",
		ResetPath:   "/password_reset",
		Namespace:   s.config.Namespace(),

		SessionMaxAge:    time.Duration(s.config.SessionMaxAge) * time.Second,
		ResetTokenMaxAge: time.Duration(s.config.ResetTokenMaxAge) * time.Second,
		Notifier:         notification.Log{},
	}

	jwtOptions := jwtmiddleware.Options{
		ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) { return []byte(s.config.JWTSecret), nil },
		SigningMethod:       JWTSigningMethod,
		UserProperty:        handlers.TokenContextKey,
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err string) {
			utils.JSONError(w, err, http.StatusUnauthorized)
		},
	}
	jwtMiddleware := jwtmiddleware.New(jwtOptions)

	// Check audience
	aud := middleware.Audience{
		Value:           baseURLFunc,
		TokenContextKey: handlers.TokenContextKey,
	}

	m := mux.NewRouter()

	m.Use((&middleware.Logging{}).Handler)
	m.Use(middleware.NewPrometheus().Handler)
	m.Use((&middleware.Recover{}).Handler)

	// Login API
	m.Methods("GET", "POST").Path("/password_reset").HandlerFunc(usersHandler.ResetPassword)
	m.Methods("GET", "POST").Path("/login").HandlerFunc(usersHandler.Login)
	m.Methods("GET").Path("/refresh").Handler(jwtMiddleware.Handler(aud.Handler(http.HandlerFunc(usersHandler.Refresh))))

	// Users API
	ud := middleware.TokenUserData{
		Namespace:       s.config.Namespace(),
		TokenContextKey: handlers.TokenContextKey,
		UserContextKey:  handlers.UserContextKey,
		DefaultRole:     handlers.RoleAnonymous,
		RolePrefix:      handlers.RolePrefix,
	}

	umux := m.PathPrefix("/users").Subrouter()
	umux.Use(jwtMiddleware.Handler)
	umux.Use(aud.Handler)
	umux.Use(ud.Handler)

	umux.Methods("POST").Path("/").HandlerFunc(usersHandler.NewUser)
	umux.Methods("GET").Path("/").HandlerFunc(usersHandler.GetUsers)
	umux.Methods("GET").Path("/{id}").HandlerFunc(usersHandler.GetUser)
	umux.Methods("PATCH").Path("/{id}").HandlerFunc(usersHandler.PatchUser)
	umux.Methods("DELETE").Path("/{id}").HandlerFunc(usersHandler.DeleteUser)

	// Miscellaneous
	m.Methods("GET").Path("/version").Handler(handlers.VersionHandler(version))
	m.Methods("GET").Path("/metrics").Handler(promhttp.Handler())

	m.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		utils.JSONError(w, "Resource not found", http.StatusNotFound)
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
