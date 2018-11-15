package middleware

import (
	"net/http"

	"github.com/ecadlabs/auth/errors"
	"github.com/ecadlabs/auth/utils"
	"github.com/dgrijalva/jwt-go"
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
			} else {
				utils.JSONErrorResponse(w, errors.ErrAudience)
				return
			}
		}

		utils.JSONError(w, "", errors.CodeForbidden)
	})
}
