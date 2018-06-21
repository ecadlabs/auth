package handlers

import (
	"context"
	"encoding/json"
	"git.ecadlabs.com/ecad/auth/users"
	"git.ecadlabs.com/ecad/auth/utils"
	"github.com/auth0/go-jwt-middleware"
	"github.com/dgrijalva/jwt-go"
	"github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"strings"
	"time"
)

const (
	TokenContextKey = "token"
)

type TokenHandler struct {
	Storage          *users.Storage
	Timeout          time.Duration
	SessionMaxAge    time.Duration
	JWTSecretGetter  func() ([]byte, error)
	JWTSigningMethod jwt.SigningMethod
	Namespace        string
	BaseURL          func() string
	RefreshURL       func() string
	ResetURL         func() string
}

func (t *TokenHandler) context(parent context.Context) context.Context {
	if t.Timeout != 0 {
		ctx, _ := context.WithTimeout(parent, t.Timeout)
		return ctx
	}
	return parent
}

func (t *TokenHandler) writeTokenWithClaims(w http.ResponseWriter, claims jwt.Claims) {
	token := jwt.NewWithClaims(t.JWTSigningMethod, claims)
	secret, err := t.JWTSecretGetter()
	if err != nil {
		utils.JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tokenString, err := token.SignedString(secret)
	if err != nil {
		utils.JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := struct {
		Token      string `json:"token"`
		RefreshURL string `json:"refresh,omitempty"`
	}{
		Token:      tokenString,
		RefreshURL: t.RefreshURL(),
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")

	utils.JSONResponse(w, http.StatusOK, &response)
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
			utils.JSONError(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	if request.Name == "" || request.Password == "" {
		utils.JSONError(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	user, err := t.Storage.GetUserByEmail(t.context(r.Context()), request.Name)
	if err != nil {
		log.Error(err)
		utils.JSONError(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	if len(user.PasswordHash) == 0 {
		utils.JSONError(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(request.Password)); err != nil {
		log.Error(err)
		utils.JSONError(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	// Don't allow unverified users to log in
	if !user.EmailVerified {
		utils.JSONError(w, "Email is not verified", http.StatusForbidden)
		return
	}

	roles := make([]string, 0, len(user.Roles))
	for r := range user.Roles {
		roles = append(roles, r)
	}

	ns := t.Namespace
	now := time.Now()

	claims := jwt.MapClaims{
		"sub": user.ID,
		"exp": now.Add(t.SessionMaxAge).Unix(),
		"iat": now.Unix(),
		"iss": t.BaseURL(),
		"aud": t.BaseURL(),
		utils.NSClaim(ns, "email"): user.Email,
		utils.NSClaim(ns, "name"):  user.Name,
		utils.NSClaim(ns, "roles"): roles,
	}

	t.writeTokenWithClaims(w, claims)
}

func (t *TokenHandler) Refresh(w http.ResponseWriter, req *http.Request) {
	claims := req.Context().Value(TokenContextKey).(*jwt.Token).Claims.(jwt.MapClaims)

	// Update timestamp only
	now := time.Now()
	claims["exp"] = now.Add(t.SessionMaxAge).Unix()
	claims["iat"] = now.Unix()

	t.writeTokenWithClaims(w, claims)
}

func (t *TokenHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Token    string `json:"token" schema:"token"`
		Password string `json:"password" schema:"password"`
	}

	if strings.Contains(r.Header.Get("Content-Type"), "application/json") {
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			log.Error(err)
			utils.JSONError(w, err.Error(), http.StatusBadRequest)
			return
		}
	} else {
		// Allow GET requests for testing purposes
		r.ParseForm()

		if err := schemaDecoder.Decode(&request, r.Form); err != nil {
			log.Error(err)
			utils.JSONError(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	// Allow authorized requests too
	if token, err := jwtmiddleware.FromAuthHeader(r); err != nil {
		utils.JSONError(w, err.Error(), http.StatusUnauthorized)
		return
	} else if token != "" {
		request.Token = token
	}

	if request.Token == "" {
		utils.JSONError(w, "Token must not be empty", http.StatusBadRequest)
		return
	}

	if request.Password == "" {
		utils.JSONError(w, "Password must not be empty", http.StatusBadRequest)
		return
	}

	// Verify token
	token, err := jwt.Parse(request.Token, func(token *jwt.Token) (interface{}, error) { return t.JWTSecretGetter() })
	if err != nil {
		log.Error(err)
		utils.JSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	if t.JWTSigningMethod.Alg() != token.Header["alg"] {
		log.Errorf("Expected %s signing method but token specified %s", t.JWTSigningMethod.Alg(), token.Header["alg"])
		utils.JSONError(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	if !token.Valid {
		log.Errorln("Invalid token")
		utils.JSONError(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	// Verify audience
	claims := token.Claims.(jwt.MapClaims)
	if !claims.VerifyAudience(t.ResetURL(), true) {
		utils.JSONError(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	// Get issue time
	iat, ok := claims["iat"].(float64)
	if !ok {
		utils.JSONError(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	// Get user id
	sub, ok := claims["sub"].(string)
	if !ok {
		utils.JSONError(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	id, err := uuid.FromString(sub)
	if err != nil {
		utils.JSONError(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Error(err)
		utils.JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = t.Storage.UpdatePasswordHash(t.context(r.Context()), id, hash, time.Unix(int64(iat), 0))
	if err != nil {
		log.Error(err)
		utils.JSONError(w, err.Error(), errorHTTPStatus(err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
