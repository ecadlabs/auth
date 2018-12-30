package service

import (
	"database/sql"
	"net/http"
	"net/mail"
	"net/url"
	"strconv"
	"time"

	"github.com/auth0/go-jwt-middleware"
	"github.com/dgrijalva/jwt-go"
	"github.com/ecadlabs/auth/errors"
	"github.com/ecadlabs/auth/handlers"
	"github.com/ecadlabs/auth/logger"
	"github.com/ecadlabs/auth/middleware"
	"github.com/ecadlabs/auth/notification"
	"github.com/ecadlabs/auth/rbac"
	"github.com/ecadlabs/auth/storage"
	"github.com/ecadlabs/auth/utils"
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
	config            Config
	storage           *storage.Storage
	tenantStorage     *storage.TenantStorage
	membershipStorage *storage.MembershipStorage
	notifier          notification.Notifier
	DB                *sql.DB
	ac                rbac.RBAC
}

func New(c *Config, ac rbac.RBAC) (*Service, error) {
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

	var dbCon = sqlx.NewDb(db, "postgres")

	return &Service{
		config:            *c,
		storage:           &storage.Storage{DB: dbCon},
		tenantStorage:     &storage.TenantStorage{DB: dbCon},
		membershipStorage: &storage.MembershipStorage{DB: dbCon},
		DB:                db,
		notifier:          notifier,
		ac:                ac,
	}, nil
}

func (s *Service) APIHandler() http.Handler {
	baseURLFunc := s.config.GetBaseURLFunc()

	dbLogger := logrus.New()
	dbLogger.AddHook(&logger.Hook{
		DB: s.DB,
	})

	tokenFactory := &handlers.TokenFactory{
		Namespace: s.config.Namespace(),
		JWTSecretGetter: func() ([]byte, error) {
			return []byte(s.config.JWTSecret), nil
		},
		JWTSigningMethod: JWTSigningMethod,

		BaseURL:       baseURLFunc,
		SessionMaxAge: time.Duration(s.config.SessionMaxAge) * time.Second,
	}

	usersHandler := &handlers.Users{
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

		Enforcer: s.ac,

		SessionMaxAge:          time.Duration(s.config.SessionMaxAge) * time.Second,
		ResetTokenMaxAge:       time.Duration(s.config.ResetTokenMaxAge) * time.Second,
		EmailUpdateTokenMaxAge: time.Duration(s.config.EmailUpdateTokenMaxAge) * time.Second,

		AuxLogger: dbLogger,
		Notifier:  s.notifier,
	}

	tenantsHandler := &handlers.Tenants{
		UserStorage:       s.storage,
		Storage:           s.tenantStorage,
		MembershipStorage: s.membershipStorage,
		Timeout:           time.Duration(s.config.DBTimeout) * time.Second,
		Enforcer:          s.ac,

		BaseURL:            baseURLFunc,
		TenantsPath:        "/tenants/",
		InvitePath:         "/tenants/accept_invite",
		TokenFactory:       tokenFactory,
		AuxLogger:          dbLogger,
		Notifier:           s.notifier,
		TenantInviteMaxAge: time.Duration(s.config.TenantInviteMaxAge) * time.Second,
	}

	jwtOptions := jwtmiddleware.Options{
		ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) { return []byte(s.config.JWTSecret), nil },
		SigningMethod:       JWTSigningMethod,
		UserProperty:        handlers.TokenContextKey,
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err string) {
			utils.JSONError(w, err, errors.CodeUnauthorized)
		},
	}
	jwtMiddleware := jwtmiddleware.New(jwtOptions)

	// Check audience
	aud := &middleware.Audience{
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
	m.Methods("GET", "POST").Path("/login/{id}").HandlerFunc(usersHandler.Login)
	m.Methods("GET", "POST").Path("/login").HandlerFunc(usersHandler.Login)

	userdata := &middleware.UserData{
		TokenContextKey: handlers.TokenContextKey,
		UserContextKey:  handlers.UserContextKey,
		Storage:         s.storage,
	}

	membershipData := &middleware.MembershipData{
		MembershipContextKey: handlers.MembershipContextKey,
		TokenContextKey:      handlers.TokenContextKey,
		Storage:              s.membershipStorage,
		Namespace:            s.config.Namespace(),
	}

	m.Methods("GET").Path("/refresh/{id}").Handler(jwtMiddleware.Handler(aud.Handler(userdata.Handler(http.HandlerFunc(usersHandler.Refresh)))))

	// Users API
	m.Methods("POST").Path("/request_email_update").Handler(jwtMiddleware.Handler(aud.Handler(userdata.Handler(http.HandlerFunc(usersHandler.SendUpdateEmailRequest)))))
	m.Methods("POST").Path("/email_update").HandlerFunc(usersHandler.UpdateEmail)

	umux := m.PathPrefix("/users").Subrouter()
	umux.Use(jwtMiddleware.Handler)
	umux.Use(aud.Handler)
	umux.Use(userdata.Handler)
	umux.Use(membershipData.Handler)

	umux.Methods("POST").Path("/").HandlerFunc(usersHandler.NewUser)
	umux.Methods("GET").Path("/").HandlerFunc(usersHandler.GetUsers)
	umux.Methods("GET").Path("/{id}").HandlerFunc(usersHandler.GetUser)
	umux.Methods("PATCH").Path("/{id}").HandlerFunc(usersHandler.PatchUser)
	umux.Methods("DELETE").Path("/{id}").HandlerFunc(usersHandler.DeleteUser)

	// Tenants API
	tmux := m.PathPrefix("/tenants").Subrouter()
	tmux.Use(jwtMiddleware.Handler)
	tmux.Use(aud.Handler)
	tmux.Use(userdata.Handler)
	tmux.Use(membershipData.Handler)

	tmux.Methods("POST").Path("/").HandlerFunc(tenantsHandler.CreateTenant)
	tmux.Methods("GET").Path("/{id}").HandlerFunc(tenantsHandler.FindTenant)
	tmux.Methods("GET").Path("/").HandlerFunc(tenantsHandler.FindTenants)
	tmux.Methods("DELETE").Path("/{id}").HandlerFunc(tenantsHandler.DeleteTenant)
	tmux.Methods("PATCH").Path("/{id}").HandlerFunc(tenantsHandler.UpdateTenant)
	tmux.Methods("POST").Path("/{id}/invite").HandlerFunc(tenantsHandler.InviteExistingUser)
	tmux.Methods("POST").Path("/accept_invite").HandlerFunc(tenantsHandler.AcceptInvite)

	// Log API
	lmux := m.PathPrefix("/logs").Subrouter()
	lmux.Use(jwtMiddleware.Handler)
	lmux.Use(aud.Handler)
	lmux.Use(userdata.Handler)

	lmux.Methods("GET").Path("/").HandlerFunc(usersHandler.GetLogs)

	// Roles API
	rbacHandler := &handlers.RolesHandler{
		DB:      s.ac,
		Timeout: time.Duration(s.config.DBTimeout) * time.Second,
	}

	rmux := m.PathPrefix("/rbac").Subrouter()
	rmux.Use(jwtMiddleware.Handler)

	rmux.Methods("GET").Path("/roles/").HandlerFunc(rbacHandler.GetRoles)
	rmux.Methods("GET").Path("/roles/{id}").HandlerFunc(rbacHandler.GetRole)
	rmux.Methods("GET").Path("/permissions/").HandlerFunc(rbacHandler.GetPermissions)
	rmux.Methods("GET").Path("/permissions/{id}").HandlerFunc(rbacHandler.GetPermission)

	// Miscellaneous
	m.Methods("GET").Path("/version").Handler(handlers.VersionHandler(version))
	m.Methods("GET").Path("/metrics").Handler(promhttp.Handler())

	m.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		utils.JSONErrorResponse(w, errors.ErrResourceNotFound)
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
