package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"context"
	"git.ecadlabs.com/ecad/auth/authenticator"
	"github.com/dgrijalva/jwt-go"
	log "github.com/sirupsen/logrus"
)

const (
	defaultIss = "auth.service"
)

// Login is a login endpoint handler
type Login struct {
	SessionMaxAge    time.Duration
	Authenticator    authenticator.Authenticator
	JWTSecretGetter  func() ([]byte, error)
	Namespace        string
	AuthTimeout      time.Duration
	JWTSigningMethod jwt.SigningMethod
	Logger           log.FieldLogger
}

func (l *Login) log() log.FieldLogger {
	if l.Logger != nil {
		return l.Logger
	}
	return log.StandardLogger()
}

func (l *Login) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Name     string `json:"name"`
		Password string `json:"password"`
	}

	if name, password, ok := r.BasicAuth(); ok {
		request.Name = name
		request.Password = password
	} else {
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			JSONError(w, err.Error(), http.StatusUnauthorized)
			return
		}
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")

	ctx, _ := context.WithTimeout(r.Context(), l.AuthTimeout)
	result, err := l.Authenticator.Authenticate(ctx, &authenticator.Credentials{
		ID:     request.Name,
		Secret: []byte(request.Password),
	})

	if err != nil {
		if ae, ok := err.(authenticator.Error); ok && ae.Rejected() {
			JSONError(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		} else {
			JSONError(w, err.Error(), http.StatusServiceUnavailable)
		}

		l.log().WithField("err", err).Println("Authentication error")
		return
	}

	now := time.Now()
	iss := l.Namespace
	if iss == "" {
		iss = defaultIss
	}
	defaultClaims := map[string]interface{}{
		"exp": now.Add(l.SessionMaxAge).Unix(),
		"iss": iss, // Should be URL according to RFC
		"iat": now.Unix(),
		"sub": request.Name,
	}

	claims := result.Claims()

	if claims == nil {
		claims = make(map[string]interface{})
	}

	for k, v := range defaultClaims {
		// Claims may be set by backend
		if _, ok := claims[k]; !ok {
			claims[k] = v
		}
	}

	token := jwt.NewWithClaims(l.JWTSigningMethod, jwt.MapClaims(claims))
	secret, err := l.JWTSecretGetter()
	if err != nil {
		JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tokenString, err := token.SignedString(secret)
	if err != nil {
		JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := struct {
		Token string `json:"token"`
	}{
		Token: tokenString,
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(&response)
}
