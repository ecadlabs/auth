package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"git.ecadlabs.com/ecad/auth/storage"
	"git.ecadlabs.com/ecad/auth/utils"
	"github.com/dgrijalva/jwt-go"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

const (
	TokenContextKey = "token"
)

func (u *Users) writeUserToken(w http.ResponseWriter, user *storage.User) error {
	roles := make([]string, 0, len(user.Roles))
	for r := range user.Roles {
		roles = append(roles, r)
	}

	now := time.Now()

	claims := jwt.MapClaims{
		"sub": user.ID,
		"exp": now.Add(u.SessionMaxAge).Unix(),
		"iat": now.Unix(),
		"iss": u.BaseURL(),
		"aud": u.BaseURL(),
		utils.NSClaim(u.Namespace, "email"): user.Email,
		utils.NSClaim(u.Namespace, "name"):  user.Name,
		utils.NSClaim(u.Namespace, "roles"): roles,
	}

	token := jwt.NewWithClaims(u.JWTSigningMethod, claims)
	secret, err := u.JWTSecretGetter()
	if err != nil {
		return err
	}

	tokenString, err := token.SignedString(secret)
	if err != nil {
		return err
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

	return nil
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

	if err := u.Storage.UpdateLoginInfo(u.context(r), user.ID, getRemoteAddr(r)); err != nil {
		utils.JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := u.writeUserToken(w, user); err != nil {
		utils.JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Log
	if u.AuxLogger != nil {
		u.AuxLogger.WithFields(logFields(EvLogin, user.ID, user.ID, r)).WithField("email", user.Email).Printf("User %v logged in", user.ID)
	}
}

func (u *Users) Refresh(w http.ResponseWriter, r *http.Request) {
	self := r.Context().Value(UserContextKey).(*storage.User)

	if err := u.Storage.UpdateRefreshInfo(u.context(r), self.ID, getRemoteAddr(r)); err != nil {
		utils.JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := u.writeUserToken(w, self); err != nil {
		utils.JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
