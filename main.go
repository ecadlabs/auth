package main

import (
	"context"
	"flag"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"git.ecadlabs.com/ecad/auth/router"
	"github.com/dgrijalva/jwt-go"
)

const (
	version = "0.0.1"
)

var jwtSigningMethod = jwt.SigningMethodHS256

func main() {
	httpAddr := flag.String("http", ":8000", "HTTP service address.")
	healthAddr := flag.String("health", ":8001", "Health service address.")
	secret := flag.String("secret", "secret", "JWT signing secret.")
	dbURL := flag.String("db", "postgres://localhost/auth?connect_timeout=10&sslmode=disable", "PostgreSQL server URL")
	flag.Parse()

	log.Println("Starting Auth service...")
	log.Printf("Health service listening on %s", *healthAddr)
	log.Printf("HTTP service listening on %s", *httpAddr)

	conf := router.Config{
		PostgresURL:      *dbURL,
		JWTSecret:        []byte(*secret),
		JWTSigningMethod: jwtSigningMethod,
	}

	r, err := conf.Handlers()
	if err != nil {
		log.Fatal(err)
	}

	// Start servers
	healthServer := &http.Server{
		Addr:    *healthAddr,
		Handler: r.Health,
	}

	errChan := make(chan error, 10)

	go func() {
		errChan <- healthServer.ListenAndServe()
	}()

	defer healthServer.Shutdown(context.Background())

	httpServer := &http.Server{
		Addr:    *httpAddr,
		Handler: r.Login,
	}

	go func() {
		errChan <- httpServer.ListenAndServe()
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
