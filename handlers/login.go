package handlers

import (
	"encoding/json"
	"git.ecadlabs.com/ecad/auth/utils"
	"github.com/dgrijalva/jwt-go"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"net/http"
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
