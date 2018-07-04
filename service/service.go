package service

import (
	"database/sql"
	"log"
	"net/http"
	"net/mail"
	"net/url"
	"os"
	"strconv"
	"time"

	"git.ecadlabs.com/ecad/auth/handlers"
	"git.ecadlabs.com/ecad/auth/logger"
	"git.ecadlabs.com/ecad/auth/middleware"
	"git.ecadlabs.com/ecad/auth/notification"
	"git.ecadlabs.com/ecad/auth/storage"
	"git.ecadlabs.com/ecad/auth/utils"
	"github.com/auth0/go-jwt-middleware"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

const (
	version               = "0.0.1"
	defaultConnectTimeout = 10
)

var JWTSigningMethod = jwt.SigningMethodHS256

type Service struct {
	config   Config
	storage  *storage.Storage
	notifier notification.Notifier
	DB       *sql.DB
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

	var notifier notification.Notifier
	if c.Notifier != nil {
		notifier = c.Notifier
	} else {
		notifier, err = notification.NewEmailNotifier(&mail.Address{
			Name:    c.Email.FromName,
			Address: c.Email.FromAddress,
		}, &c.Email.TemplateData, c.Email.Driver, c.Email.Config)
		if err != nil {
			return nil, err
		}
	}

	return &Service{
		config:   *c,
		storage:  &storage.Storage{DB: sqlx.NewDb(db, "postgres")},
		DB:       db,
		notifier: notifier,
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

		BaseURL:         baseURLFunc,
		UsersPath:       "/users/",
		RefreshPath:     "/refresh",
		ResetPath:       "/password_reset",
		LogPath:         "/logs/",
		EmailUpdatePath: "/email_update",
		Namespace:       s.config.Namespace(),

		SessionMaxAge:          time.Duration(s.config.SessionMaxAge) * time.Second,
		ResetTokenMaxAge:       time.Duration(s.config.ResetTokenMaxAge) * time.Second,
		EmailUpdateTokenMaxAge: time.Duration(s.config.EmailUpdateTokenMaxAge) * time.Second,

		AuxLogger: dbLogger,
		Notifier:  s.notifier,
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
	m.Methods("POST").Path("/password_reset").HandlerFunc(usersHandler.ResetPassword)
	m.Methods("GET", "POST").Path("/request_password_reset").HandlerFunc(usersHandler.SendResetRequest)
	m.Methods("GET", "POST").Path("/login").HandlerFunc(usersHandler.Login)

	//TODO replace this "feature" with "ServiceAccount" concept.
	//Visitors can "log in" using their source IP address alone, and will get a
	//JWT token in return
	//
	// See gitlab issue ecad/auth#29
	ipCheckMiddleware, err := middleware.NewIPAccessChecker([]string{
		"10.60.58.5/32",
		"208.92.18.70/32",    //Simon montreal
		"216.232.49.35/32",   //ECADLabs vancouver
		"217.194.176.242/32", //NOC
		"217.194.176.254/32", //NOC
		"217.194.177.209/32", //VPN from vancouver
		// "172.19.0.0/24",      //Docker default class C for dev
	})

	if err != nil {
		log.Printf("Error setting up IP Access Checker middleware %e", err)
		os.Exit(1)
	}

	// /checkip checks if the RequestIP is in our access list
	m.Methods("GET").
		Path("/checkip").
		Handler(ipCheckMiddleware.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			return
		})))

	userdata := middleware.UserData{
		TokenContextKey: handlers.TokenContextKey,
		UserContextKey:  handlers.UserContextKey,
		Storage:         s.storage,
	}

	m.Methods("GET").Path("/refresh").Handler(jwtMiddleware.Handler(aud.Handler(userdata.Handler(http.HandlerFunc(usersHandler.Refresh)))))

	// Users API
	m.Methods("POST").Path("/request_email_update").Handler(jwtMiddleware.Handler(aud.Handler(userdata.Handler(http.HandlerFunc(usersHandler.SendUpdateEmailRequest)))))
	m.Methods("POST").Path("/email_update").HandlerFunc(usersHandler.UpdateEmail)

	umux := m.PathPrefix("/users").Subrouter()
	umux.Use(jwtMiddleware.Handler)
	umux.Use(aud.Handler)
	umux.Use(userdata.Handler)

	umux.Methods("POST").Path("/").HandlerFunc(usersHandler.NewUser)
	umux.Methods("GET").Path("/").HandlerFunc(usersHandler.GetUsers)
	umux.Methods("GET").Path("/{id}").HandlerFunc(usersHandler.GetUser)
	umux.Methods("PATCH").Path("/{id}").HandlerFunc(usersHandler.PatchUser)
	umux.Methods("DELETE").Path("/{id}").HandlerFunc(usersHandler.DeleteUser)

	// Log API
	lmux := m.PathPrefix("/logs").Subrouter()
	lmux.Use(jwtMiddleware.Handler)
	lmux.Use(aud.Handler)
	lmux.Use(userdata.Handler)

	lmux.Methods("GET").Path("/").HandlerFunc(usersHandler.GetLogs)

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
