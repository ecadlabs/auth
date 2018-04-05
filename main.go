package main

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/base64"
	"flag"
	"fmt"
	"git.ecadlabs.com/ecad/auth/authenticator/postgresql/migrations"
	"git.ecadlabs.com/ecad/auth/router"
	"github.com/dgrijalva/jwt-go"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

const (
	version = "0.0.1"
)

var jwtSigningMethod = jwt.SigningMethodHS256

func main() {
	var (
		config      Config
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

	flag.StringVar(&config.Address, "http", ":8000", "HTTP service address.")
	flag.StringVar(&config.HealthAddress, "health", ":8001", "Health service address.")
	flag.StringVar(&config.JWTSecret, "secret", "", "JWT signing secret.")
	flag.StringVar(&config.PostgresURL, "db", "postgres://localhost/auth?connect_timeout=10&sslmode=disable", "PostgreSQL server URL.")
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

	if config.JWTSecret == "" {
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

	log.Println("Running migrations...")
	if err := migrations.Migrate(config.PostgresURL); err != nil {
		log.Fatal(err)
	}

	if migrateOnly {
		os.Exit(0)
	}

	log.Println("Starting Auth service...")
	log.Printf("Health service listening on %s", config.HealthAddress)
	log.Printf("HTTP service listening on %s", config.Address)

	conf := router.Config{
		PostgresURL:      config.PostgresURL,
		JWTSecret:        []byte(config.JWTSecret),
		JWTSigningMethod: jwtSigningMethod,
		BaseURL:          config.BaseURL,
	}

	r, err := conf.Handlers()
	if err != nil {
		log.Fatal(err)
	}

	// Start servers
	healthServer := &http.Server{
		Addr:    config.HealthAddress,
		Handler: r.Health,
	}

	errChan := make(chan error, 10)

	go func() {
		errChan <- healthServer.ListenAndServe()
	}()

	defer healthServer.Shutdown(context.Background())

	httpServer := &http.Server{
		Addr:      config.Address,
		Handler:   r.Login,
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
