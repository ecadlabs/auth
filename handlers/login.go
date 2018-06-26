package handlers

import (
	"encoding/json"
	"git.ecadlabs.com/ecad/auth/utils"
	"github.com/dgrijalva/jwt-go"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"net"
	"net/http"
	"strings"
	"time"
)

const (
	TokenContextKey = "token"
)

func xff(r *http.Request) string {
	if fh := r.Header.Get("Forwarded"); fh != "" {
		chunks := strings.Split(fh, ",")

		for _, c := range chunks {
			opts := strings.Split(strings.TrimSpace(c), ";")

			for _, o := range opts {
				v := strings.SplitN(strings.TrimSpace(o), "=", 2)
				if len(v) == 2 && v[0] == "for" && v[1] != "" {
					return v[1]
				}
			}
		}
	}

	if xfh := r.Header.Get("X-Forwarded-For"); xfh != "" {
		chunks := strings.Split(xfh, ",")
		for _, c := range chunks {
			if c = strings.TrimSpace(c); c != "" {
				return c
			}
		}
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}

	return r.RemoteAddr
}

func (u *Users) writeTokenWithClaims(w http.ResponseWriter, claims jwt.Claims) error {
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

	if err := u.writeTokenWithClaims(w, claims); err != nil {
		utils.JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Log
	if u.AuxLogger != nil {
		u.AuxLogger.WithFields(logFields(log.Fields{
			"address": xff(r),
			"email":   user.Email,
		}, EvLogin, user.ID, user.ID)).Printf("User %v logged in", user.ID)
	}
}

func (u *Users) Refresh(w http.ResponseWriter, req *http.Request) {
	claims := req.Context().Value(TokenContextKey).(*jwt.Token).Claims.(jwt.MapClaims)

	// Update timestamp only
	now := time.Now()
	claims["exp"] = now.Add(u.SessionMaxAge).Unix()
	claims["iat"] = now.Unix()

	u.writeTokenWithClaims(w, claims)
}
