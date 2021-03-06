package service

import (
	"database/sql"
	"io/ioutil"
	"net/http"
	"net/mail"
	"net/url"
	"strconv"
	"time"

	jwtmiddleware "github.com/auth0/go-jwt-middleware"
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
	log "github.com/sirupsen/logrus"
)

const (
	version               = "0.0.1"
	defaultConnectTimeout = 10
)

var JWTSigningMethod = jwt.SigningMethodHS256

type Service struct {
	config    Config
	storage   *storage.Storage
	notifier  notification.Notifier
	DB        *sql.DB
	ac        rbac.RBAC
	enableLog bool
}

func New(c *Config, ac rbac.RBAC, enableLog bool) (*Service, error) {
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
		}, c.Email.Driver, c.Email.Config)
		if err != nil {
			return nil, err
		}
	}

	var dbCon = sqlx.NewDb(db, "postgres")

	return &Service{
		config:    *c,
		storage:   &storage.Storage{DB: dbCon, DefaultRole: ac.GetDefaultRole()},
		DB:        db,
		notifier:  notifier,
		ac:        ac,
		enableLog: enableLog,
	}, nil
}

func (s *Service) APIHandler() http.Handler {
	dbLogger := log.New()
	if !s.enableLog {
		dbLogger.Out = ioutil.Discard
	}
	dbLogger.AddHook(&logger.Hook{
		DB: s.DB,
	})

	tokenFactory := &handlers.TokenFactory{
		Namespace: s.config.Namespace(),
		JWTSecretGetter: func() ([]byte, error) {
			return []byte(s.config.JWTSecret), nil
		},
		JWTSigningMethod: JWTSigningMethod,
	}

	usersHandler := &handlers.Users{
		Storage: s.storage,
		Timeout: time.Duration(s.config.DBTimeout) * time.Second,

		JWTSecretGetter: func() ([]byte, error) {
			return []byte(s.config.JWTSecret), nil
		},
		JWTSigningMethod: JWTSigningMethod,

		UsersPath:       "/users/",
		RefreshPath:     "/refresh",
		ResetPath:       "/password_reset",
		LogPath:         "/logs/",
		EmailUpdatePath: "/email_update",
		Namespace:       s.config.Namespace(),

		Enforcer: s.ac,

		AuxLogger: dbLogger,
		Notifier:  s.notifier,
	}

	tenantsHandler := &handlers.Tenants{
		Storage:  s.storage,
		Timeout:  time.Duration(s.config.DBTimeout) * time.Second,
		Enforcer: s.ac,

		TenantsPath:  "/tenants/",
		InvitePath:   "/tenants/accept_invite",
		TokenFactory: tokenFactory,
		AuxLogger:    dbLogger,
		Notifier:     s.notifier,
	}

	membershipsHandler := &handlers.Memberships{
		Storage:     s.storage,
		Timeout:     time.Duration(s.config.DBTimeout) * time.Second,
		Enforcer:    s.ac,
		TenantsPath: "/tenants/",
		UsersPath:   "/users/",
		AuxLogger:   dbLogger,
	}

	jwtOptions := jwtmiddleware.Options{
		ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) { return []byte(s.config.JWTSecret), nil },
		SigningMethod:       JWTSigningMethod,
		UserProperty:        middleware.TokenContextKey,
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err string) {
			utils.JSONError(w, err, errors.CodeUnauthorized)
		},
	}
	jwtMiddleware := jwtmiddleware.New(jwtOptions)

	// Check audience
	aud := &middleware.Audience{
		Value: func(r *http.Request) string {
			site := r.Context().Value(middleware.DomainConfigContextKey).(*middleware.DomainConfigData)
			return site.GetBaseURL()
		},
		Namespace: s.config.Namespace(),
	}

	domainData := &middleware.DomainConfig{
		Storage: &s.config,
	}

	m := mux.NewRouter()

	m.Use(middleware.NewPrometheusWithHandlerID().Handler)
	if s.enableLog {
		m.Use((&middleware.Logging{}).Handler)
	}
	m.Use((&middleware.Recover{}).Handler)
	m.Use(domainData.Handler)

	// Login API
	m.Methods("POST").Path("/password_reset").HandlerFunc(usersHandler.ResetPassword)
	m.Methods("GET", "POST").Path("/request_password_reset").HandlerFunc(usersHandler.SendResetRequest)
	m.Methods("GET", "POST").Path("/login/{id}").HandlerFunc(usersHandler.Login)
	m.Methods("GET", "POST").Path("/login").HandlerFunc(usersHandler.Login)

	userdata := &middleware.UserData{
		Storage: s.storage,
	}

	membershipData := &middleware.MembershipData{
		Storage:   s.storage,
		Namespace: s.config.Namespace(),
	}

	serviceAPI := &middleware.ServiceAPI{
		Storage:   s.storage,
		Namespace: s.config.Namespace(),
	}

	m.Methods("GET").Path("/refresh").Handler(jwtMiddleware.Handler(aud.Handler(userdata.Handler(membershipData.Handler(http.HandlerFunc(usersHandler.Refresh))))))

	// Users API
	m.Methods("POST").Path("/request_email_update").Handler(jwtMiddleware.Handler(aud.Handler(userdata.Handler(http.HandlerFunc(usersHandler.SendUpdateEmailRequest)))))
	m.Methods("POST").Path("/email_update").HandlerFunc(usersHandler.UpdateEmail)

	umux := m.PathPrefix("/users").Subrouter()
	umux.Use(jwtMiddleware.Handler)
	umux.Use(serviceAPI.Handler)
	umux.Use(aud.Handler)
	umux.Use(userdata.Handler)
	umux.Use(membershipData.Handler)

	umux.Methods("POST").Path("/").HandlerFunc(usersHandler.NewUser)
	umux.Methods("GET").Path("/").HandlerFunc(usersHandler.GetUsers)
	umux.Methods("GET").Path("/{id}").HandlerFunc(usersHandler.GetUser)
	umux.Methods("PATCH").Path("/{id}").HandlerFunc(usersHandler.PatchUser)
	umux.Methods("DELETE").Path("/{id}").HandlerFunc(usersHandler.DeleteUser)
	umux.Methods("GET").Path("/{userId}/memberships/").HandlerFunc(membershipsHandler.FindUserMemberships)

	umux.Methods("POST").Path("/{userId}/api_keys/").HandlerFunc(usersHandler.NewAPIKey)
	umux.Methods("GET").Path("/{userId}/api_keys/{keyId}").HandlerFunc(usersHandler.GetAPIKey)
	umux.Methods("GET").Path("/{userId}/api_keys/").HandlerFunc(usersHandler.GetAPIKeys)
	umux.Methods("DELETE").Path("/{userId}/api_keys/{keyId}").HandlerFunc(usersHandler.DeleteAPIKey)
	umux.Methods("GET").Path("/{userId}/api_keys/{keyId}/token").HandlerFunc(usersHandler.GetAPIToken)

	// Tenants API
	tmux := m.PathPrefix("/tenants").Subrouter()
	tmux.Use(jwtMiddleware.Handler)
	tmux.Use(serviceAPI.Handler)
	tmux.Use(aud.Handler)
	tmux.Use(membershipData.Handler)

	tmux.Methods("POST").Path("/").HandlerFunc(tenantsHandler.CreateTenant)
	tmux.Methods("GET").Path("/{id}").HandlerFunc(tenantsHandler.FindTenant)
	tmux.Methods("GET").Path("/").HandlerFunc(tenantsHandler.FindTenants)
	tmux.Methods("DELETE").Path("/{id}").HandlerFunc(tenantsHandler.DeleteTenant)
	tmux.Methods("PATCH").Path("/{id}").HandlerFunc(tenantsHandler.UpdateTenant)

	tmux.Methods("POST").Path("/{id}/members/").Handler(userdata.Handler(http.HandlerFunc(tenantsHandler.InviteExistingUser)))
	tmux.Methods("GET").Path("/{tenantId}/members/").HandlerFunc(membershipsHandler.FindTenantMemberships)
	tmux.Methods("PATCH").Path("/{tenantId}/members/{userId}").HandlerFunc(membershipsHandler.PatchMembership)
	tmux.Methods("DELETE").Path("/{tenantId}/members/{userId}").HandlerFunc(membershipsHandler.DeleteMembership)

	amux := m.PathPrefix("/tenants/accept_invite").Subrouter()

	amux.Methods("POST").Path("").HandlerFunc(tenantsHandler.AcceptInvite)

	// Members API
	mmux := m.PathPrefix("/members").Subrouter()
	mmux.Use(jwtMiddleware.Handler)
	mmux.Use(aud.Handler)
	mmux.Use(userdata.Handler)
	mmux.Use(membershipData.Handler)

	mmux.Methods("GET").Path("/{memberId}/user").HandlerFunc(usersHandler.FindUserByMembershipID)

	// Log API
	lmux := m.PathPrefix("/logs").Subrouter()
	lmux.Use(jwtMiddleware.Handler)
	lmux.Use(serviceAPI.Handler)
	lmux.Use(aud.Handler)
	lmux.Use(userdata.Handler)
	lmux.Use(membershipData.Handler)

	lmux.Methods("GET").Path("/").HandlerFunc(usersHandler.GetLogs)

	// Roles API
	rbacHandler := &handlers.RolesHandler{
		DB:      s.ac,
		Timeout: time.Duration(s.config.DBTimeout) * time.Second,
	}

	rmux := m.PathPrefix("/rbac").Subrouter()
	rmux.Use(jwtMiddleware.Handler)
	rmux.Use(serviceAPI.Handler)

	rmux.Methods("GET").Path("/roles/").HandlerFunc(rbacHandler.GetRoles)
	rmux.Methods("GET").Path("/roles/{id}").HandlerFunc(rbacHandler.GetRole)
	rmux.Methods("GET").Path("/permissions/").HandlerFunc(rbacHandler.GetPermissions)
	rmux.Methods("GET").Path("/permissions/{id}").HandlerFunc(rbacHandler.GetPermission)

	m.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		utils.JSONErrorResponse(w, errors.ErrResourceNotFound)
		if !s.enableLog {
			return
		}

		fields := log.Fields{
			"status":   errors.ErrResourceNotFound,
			"hostname": r.Host,
			"method":   r.Method,
			"path":     r.URL.Path,
		}
		log.WithFields(fields).Println(r.Method + " " + r.URL.Path)
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
	if s.enableLog {
		m.Use((&middleware.Logging{}).Handler)
	}
	m.Use((&middleware.Recover{}).Handler)

	m.Methods("GET").Path("/healthz").Handler(healthMon)
	m.Methods("GET").Path("/version").Handler(handlers.VersionHandler(version))
	m.Methods("GET").Path("/metrics").Handler(promhttp.Handler())

	return m
}
