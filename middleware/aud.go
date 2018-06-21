package middleware

import (
	"git.ecadlabs.com/ecad/auth/utils"
	"github.com/dgrijalva/jwt-go"
	"net/http"
)

type Audience struct {
	TokenContextKey string
	Value           func() string
}

func (a *Audience) Handler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if token, ok := r.Context().Value(a.TokenContextKey).(*jwt.Token); ok {
			claims := token.Claims.(jwt.MapClaims)
			if claims.VerifyAudience(a.Value(), true) {
				h.ServeHTTP(w, r)
				return
			}
		}

		utils.JSONError(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
	})
}
