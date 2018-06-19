package main

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"database/sql"
	"encoding/base64"
	"flag"
	"fmt"
	"git.ecadlabs.com/ecad/auth/migrations"
	"git.ecadlabs.com/ecad/auth/service"
	"github.com/dgrijalva/jwt-go"
	"github.com/golang-migrate/migrate"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
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
		genPwd      string
		genSecret   int
		useBase64   bool
		migrateOnly bool
	)

	flag.StringVar(&genPwd, "bcrypt", "", "Generate password hash.")
	flag.IntVar(&genSecret, "gen_secret", 0, "Generate random JWT secret of N bytes.")
	flag.BoolVar(&useBase64, "base64_secret", false, "Encode generated secret using base64.")
	flag.StringVar(&configFile, "c", "", "Config file.")
	flag.BoolVar(&migrateOnly, "migrate", false, "Migrate and exit immediately.")

	flag.StringVar(&config.BaseURL, "base", "http://localhost:8000", "Base URL.")
	flag.StringVar(&config.Address, "http", ":8000", "HTTP service address.")
	flag.StringVar(&config.HealthAddress, "health", ":8001", "Health service address.")
	flag.StringVar(&config.JWTSecret, "secret", "", "JWT signing secret.")
	flag.IntVar(&config.SessionMaxAge, "max_age", 259200, "Session max age, sec.")
	flag.StringVar(&config.PostgresURL, "db", "postgres://localhost/users?connect_timeout=10&sslmode=disable", "PostgreSQL server URL.")
	flag.IntVar(&config.PostgresRetriesNum, "db_retries_num", 5, "Number of attempts to establish PostgreSQL connection")
	flag.IntVar(&config.PostgresRetryDelay, "db_retry_delay", 10, "Delay between connection attempts attempts")
	flag.IntVar(&config.DBTimeout, "timeout", 10, "DB timeout, sec.")
	flag.BoolVar(&config.TLS, "tls", false, "Enable TLS.")
	flag.StringVar(&config.TLSCert, "tlscert", "", "TLS certificate file.")
	flag.StringVar(&config.TLSKey, "tlskey", "", "TLS private key file.")

	flag.Parse()

	if genSecret != 0 {
		buf := make([]byte, genSecret)
		if _, err := rand.Read(buf); err != nil {
			log.Fatal(err)
		}

		if !useBase64 {
			os.Stdout.Write(buf)
			os.Exit(0)
		}

		s := base64.StdEncoding.EncodeToString(buf)
		fmt.Println(s)
		os.Exit(0)
	}

	if genPwd != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(genPwd), bcrypt.DefaultCost)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(string(hash))
		os.Exit(0)
	}

	if configFile != "" {
		if err := config.Load(configFile); err != nil {
			log.Fatal(err)
		}

		// Override from command line
		flag.Parse()
	}

	if config.JWTSecret == "" && !migrateOnly {
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

	// Service instance
	svc, err := config.New()
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
