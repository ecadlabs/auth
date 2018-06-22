package handlers

import (
	"encoding/json"
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

func (u *Users) writeTokenWithClaims(w http.ResponseWriter, claims jwt.Claims) {
	token := jwt.NewWithClaims(u.JWTSigningMethod, claims)
	secret, err := u.JWTSecretGetter()
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
		RefreshURL: u.RefreshURL(),
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")

	utils.JSONResponse(w, http.StatusOK, &response)
}

// Login is a login endpoint handler
func (u *Users) Login(w http.ResponseWriter, r *http.Request) {
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

	user, err := u.Storage.GetUserByEmail(u.context(r), request.Name)
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

	ns := u.Namespace
	now := time.Now()

	claims := jwt.MapClaims{
		"sub": user.ID,
		"exp": now.Add(u.SessionMaxAge).Unix(),
		"iat": now.Unix(),
		"iss": u.BaseURL(),
		"aud": u.BaseURL(),
		utils.NSClaim(ns, "email"): user.Email,
		utils.NSClaim(ns, "name"):  user.Name,
		utils.NSClaim(ns, "roles"): roles,
	}

	u.writeTokenWithClaims(w, claims)
}

func (u *Users) Refresh(w http.ResponseWriter, req *http.Request) {
	claims := req.Context().Value(TokenContextKey).(*jwt.Token).Claims.(jwt.MapClaims)

	// Update timestamp only
	now := time.Now()
	claims["exp"] = now.Add(u.SessionMaxAge).Unix()
	claims["iat"] = now.Unix()

	u.writeTokenWithClaims(w, claims)
}

func (u *Users) ResetPassword(w http.ResponseWriter, r *http.Request) {
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
	token, err := jwt.Parse(request.Token, func(token *jwt.Token) (interface{}, error) { return u.JWTSecretGetter() })
	if err != nil {
		log.Error(err)
		utils.JSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	if u.JWTSigningMethod.Alg() != token.Header["alg"] {
		log.Errorf("Expected %s signing method but token specified %s", u.JWTSigningMethod.Alg(), token.Header["alg"])
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
	if !claims.VerifyAudience(u.ResetURL(), true) {
		utils.JSONError(w, "Not a session token", http.StatusUnauthorized)
		return
	}

	// Get password generation
	gen, ok := claims[utils.NSClaim(u.Namespace, "gen")].(float64)
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

	err = u.Storage.UpdatePasswordHash(u.context(r), id, hash, int(gen))
	if err != nil {
		log.Error(err)
		utils.JSONError(w, err.Error(), errorHTTPStatus(err))
		return
	}

	w.WriteHeader(http.StatusNoContent)

	// Log
	if u.AuxLogger != nil {
		u.AuxLogger.WithFields(logFields(nil, EvReset, uuid.Nil, id)).Printf("Password for account %v reset", id)
	}
}
