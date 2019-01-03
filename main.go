package main

import (
	"context"
	"crypto/tls"
	"database/sql"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/ecadlabs/auth/migrations"
	"github.com/ecadlabs/auth/rbac"
	"github.com/ecadlabs/auth/service"
	"github.com/golang-migrate/migrate"
	log "github.com/sirupsen/logrus"
)

const (
	version = "0.0.1"
)

var jwtSigningMethod = jwt.SigningMethodHS256

func doMigrate(db *sql.DB) error {
	m, err := migrations.NewDB(db)
	if err != nil {
		return err
	}

	ver, _, err := m.Version()
	if err == nil {
		log.WithField("version", ver).Println("Current DB Version")
	} else if err != migrate.ErrNilVersion {
		return err
	}

	if err := m.Up(); err == nil {
		ver, _, err := m.Version()
		if err != nil {
			return err
		}
		log.WithField("version", ver).Println("Migrated successfully")
	} else if err == migrate.ErrNoChange {
		log.Println("No migrations")
	} else {
		return err
	}

	return nil
}

func main() {
	var (
		config      service.Config
		configFile  string
		migrateOnly bool
		bootstrap   bool
		rbacFile    string
	)

	flag.StringVar(&configFile, "c", "", "Config file.")
	flag.StringVar(&rbacFile, "r", "", "RBAC file.")
	flag.BoolVar(&migrateOnly, "migrate", false, "Migrate and exit immediately.")
	flag.BoolVar(&bootstrap, "bootstrap", false, "Bootstrap DB.")

	flag.StringVar(&config.BaseURL, "base", "http://localhost:8000", "Base URL.")
	flag.StringVar(&config.Address, "http", ":8000", "HTTP service address.")
	flag.StringVar(&config.HealthAddress, "health", ":8001", "Health service address.")
	flag.StringVar(&config.JWTSecret, "secret", "", "JWT signing secret.")
	flag.StringVar(&config.JWTNamespace, "namespace", service.DefaultNamespace, "JWT namespace prefix.")
	flag.IntVar(&config.SessionMaxAge, "max_age", 60*60*72, "Session max age, sec.")
	flag.IntVar(&config.ResetTokenMaxAge, "reset_token_max_age", 60*60*3, "Password reset token max age, sec.")
	flag.IntVar(&config.TenantInviteMaxAge, "tenant_invite_max_age", 60*60*24, "Tenant invite token max age, sec.")
	flag.IntVar(&config.EmailUpdateTokenMaxAge, "email_token_max_age", 60*60*3, "Email update token max age, sec.")
	flag.StringVar(&config.PostgresURL, "db", "postgres://localhost/users?connect_timeout=10&sslmode=disable", "PostgreSQL server URL.")
	flag.IntVar(&config.PostgresRetriesNum, "db_retries_num", 5, "Number of attempts to establish PostgreSQL connection")
	flag.IntVar(&config.PostgresRetryDelay, "db_retry_delay", 10, "Delay between connection attempts attempts")
	flag.IntVar(&config.DBTimeout, "timeout", 10, "DB timeout, sec.")
	flag.BoolVar(&config.TLS, "tls", false, "Enable TLS.")
	flag.StringVar(&config.TLSCert, "tlscert", "", "TLS certificate file.")
	flag.StringVar(&config.TLSKey, "tlskey", "", "TLS private key file.")

	flag.Parse()

	if configFile != "" {
		if err := config.Load(configFile); err != nil {
			log.Fatal(err)
		}

		// Override from command line
		flag.Parse()
	}

	var ac rbac.RBAC
	if rbacFile != "" {
		var err error
		if ac, err = rbac.LoadYAML(rbacFile); err != nil {
			log.Fatal(err)
		}
	}

	if (config.JWTSecret == "" || ac == nil) && !migrateOnly {
		flag.Usage()
		os.Exit(0)
	}

	var tlsConfig *tls.Config

	if config.TLS && config.TLSCert != "" && config.TLSKey != "" {
		cert, err := tls.LoadX509KeyPair(config.TLSCert, config.TLSKey)
		if err != nil {
			log.Fatal(err)
		}

		tlsConfig = &tls.Config{
			Certificates: []tls.Certificate{cert},
		}
	}

	if config.Email.Driver == "" {
		config.Email.Driver = "debug"
	}

	// Service instance
	svc, err := service.New(&config, ac, true)
	if err != nil {
		log.Fatal(err)
	}

	if migrateOnly {
		if err := doMigrate(svc.DB); err != nil {
			log.Fatal(err)
		}

		os.Exit(0)
	}

	log.Println("Starting Auth service...")
	log.Printf("Health service listening on %s", config.HealthAddress)

	// Start servers
	healthServer := &http.Server{
		Addr:    config.HealthAddress,
		Handler: svc.HealthHandler(),
	}

	errChan := make(chan error, 10)

	go func() {
		errChan <- healthServer.ListenAndServe()
	}()

	defer healthServer.Shutdown(context.Background())

	// Wait for connection
	var i int
	for {
		err = svc.DB.Ping()

		if err != nil {
			log.Warningln(err)
		}

		if err == nil || i >= config.PostgresRetriesNum {
			break
		}

		time.Sleep(time.Duration(config.PostgresRetryDelay) * time.Second)

		i++
	}

	if err != nil {
		log.Fatal(err)
	}

	if err := doMigrate(svc.DB); err != nil {
		log.Fatal(err)
	}

	if bootstrap {
		if _, err := svc.Bootstrap(); err != nil {
			if err != service.ErrNoBootstrap {
				log.Fatal(err)
			}
		} else {
			log.Println("DB bootstrapped successfully")
		}
	}

	log.Printf("HTTP service listening on %s", config.Address)
	httpServer := &http.Server{
		Addr:      config.Address,
		Handler:   svc.APIHandler(),
		TLSConfig: tlsConfig,
	}

	go func() {
		if httpServer.TLSConfig != nil {
			errChan <- httpServer.ListenAndServeTLS("", "")
		} else {
			errChan <- httpServer.ListenAndServe()
		}
	}()

	defer httpServer.Shutdown(context.Background())

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case err := <-errChan:
			if err != nil {
				log.Fatal(err)
			}

		case s := <-signalChan:
			log.Printf("Captured %v. Exiting...\n", s)
			return
		}
	}
}
