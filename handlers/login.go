package handlers

import (
	"context"
	"encoding/json"
	"git.ecadlabs.com/ecad/auth/users"
	"github.com/dgrijalva/jwt-go"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"time"
)

type TokenHandler struct {
	Storage          *users.Storage
	Timeout          time.Duration
	SessionMaxAge    time.Duration
	JWTSecretGetter  func() ([]byte, error)
	JWTSigningMethod jwt.SigningMethod
	Namespace        string
	RefreshURL       string
}

func (t *TokenHandler) context() context.Context {
	ctx := context.Background()
	if t.Timeout != 0 {
		ctx, _ = context.WithTimeout(ctx, t.Timeout)
	}

	return ctx
}

func (t *TokenHandler) writeTokenWithClaims(w http.ResponseWriter, claims jwt.Claims) {
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

	w.Header().Set("Access-Control-Allow-Origin", "*")

	JSONResponse(w, http.StatusOK, &response)
}

// Login is a login endpoint handler
func (t *TokenHandler) Login(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Name     string `json:"name"`
		Password string `json:"password"`
	}

	if name, password, ok := r.BasicAuth(); ok {
		request.Name = name
		request.Password = password
	} else {
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			JSONError(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	if request.Name == "" || request.Password == "" {
		JSONError(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}

	user, err := t.Storage.GetUserByEmail(t.context(), request.Name)
	if err != nil {
		log.Error(err)
		JSONError(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}

	if err := bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(request.Password)); err != nil {
		log.Error(err)
		JSONError(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}

	now := time.Now()

	claims := jwt.MapClaims{
		"sub": user.ID,
		"exp": now.Add(t.SessionMaxAge).Unix(),
		"iat": now.Unix(),
	}

	if t.Namespace != "" {
		claims["iss"] = t.Namespace
		claims[t.Namespace+"/email"] = user.Email
		claims[t.Namespace+"/name"] = user.Name
	}

	t.writeTokenWithClaims(w, claims)
}

func (t *TokenHandler) Refresh(w http.ResponseWriter, req *http.Request) {
	claims := req.Context().Value("user").(*jwt.Token).Claims.(jwt.MapClaims)

	// Update timestamp only
	now := time.Now()
	claims["exp"] = now.Add(t.SessionMaxAge).Unix()
	claims["iat"] = now.Unix()

	t.writeTokenWithClaims(w, claims)
}
