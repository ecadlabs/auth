package middleware

import (
	"net"
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/ecadlabs/auth/errors"
	"github.com/ecadlabs/auth/utils"
)

type Audience struct {
	Value     func(r *http.Request) string
	Namespace string
}

func (a *Audience) Handler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if token, ok := r.Context().Value(TokenContextKey).(*jwt.Token); ok {
			claims := token.Claims.(jwt.MapClaims)

			if addr, ok := claims[utils.NSClaim(a.Namespace, "address")].(string); ok {
				expected := net.ParseIP(addr)
				if expected == nil {
					utils.JSONErrorResponse(w, errors.ErrInvalidToken)
					return
				}

				remote := net.ParseIP(utils.GetRemoteAddr(r))
				if !expected.Equal(remote) {
					utils.JSONError(w, "", errors.CodeForbidden)
					return
				}
			}

			if !claims.VerifyAudience(a.Value(r), true) {
				utils.JSONErrorResponse(w, errors.ErrAudience)
				return
			}

			h.ServeHTTP(w, r)
			return
		}

		utils.JSONError(w, "", errors.CodeForbidden)
	})
}
