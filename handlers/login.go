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

type TokenProducer struct {
	SessionMaxAge    time.Duration
	JWTSecretGetter  func() ([]byte, error)
	Namespace        string
	JWTSigningMethod jwt.SigningMethod
	Logger           log.FieldLogger
	RefreshURL       string
}

func (t *TokenProducer) log() log.FieldLogger {
	if t.Logger != nil {
		return t.Logger
	}
	return log.StandardLogger()
}

func (t *TokenProducer) writeTokenWithClaims(w http.ResponseWriter, claims jwt.Claims) {
	token := jwt.NewWithClaims(t.JWTSigningMethod, claims)
	secret, err := t.JWTSecretGetter()
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
		Token      string `json:"token"`
		RefreshURL string `json:"refresh,omitempty"`
	}{
		Token:      tokenString,
		RefreshURL: t.RefreshURL,
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if err := json.NewEncoder(w).Encode(&response); err != nil {
		t.log().Println(err)
	}
}

// Login is a login endpoint handler
type Login struct {
	*TokenProducer
	Authenticator authenticator.Authenticator
	AuthTimeout   time.Duration
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

	l.writeTokenWithClaims(w, jwt.MapClaims(claims))
}

type Refresh struct {
	*TokenProducer
}

func (r *Refresh) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	claims := req.Context().Value("user").(*jwt.Token).Claims.(jwt.MapClaims)

	// Update timestamp only
	now := time.Now()
	claims["exp"] = now.Add(r.SessionMaxAge).Unix()
	claims["iat"] = now.Unix()

	r.writeTokenWithClaims(w, claims)
}
