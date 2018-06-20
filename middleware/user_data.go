package middleware

import (
	"context"
	"git.ecadlabs.com/ecad/auth/handlers"
	"git.ecadlabs.com/ecad/auth/users"
	"github.com/dgrijalva/jwt-go"
	"github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"net/http"
)

type UserData struct {
	Namespace       string
	TokenContextKey string
	UserContextKey  string
	DefaultRole     string
	RolePrefix      string
}

func (u *UserData) Handler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := users.User{
			Roles: make(users.Roles),
		}
		req := r.WithContext(context.WithValue(r.Context(), u.UserContextKey, &user))

		token, ok := r.Context().Value(u.TokenContextKey).(*jwt.Token)

		if !ok {
			user.Roles[u.DefaultRole] = struct{}{}
			h.ServeHTTP(w, req)
			return
		}

		claims := token.Claims.(jwt.MapClaims)

		if names, ok := claims[nsClaim(u.Namespace, "roles")].([]interface{}); ok {
			for _, name := range names {
				if s, ok := name.(string); ok {
					user.Roles[s] = struct{}{}
				}
			}
		}

		if !user.Roles.HasPrefix(u.RolePrefix) {
			user.Roles[u.DefaultRole] = struct{}{}
		}

		if sub, ok := claims["sub"].(string); ok {
			if id, err := uuid.FromString(sub); err == nil {
				user.ID = id
				h.ServeHTTP(w, req)
				return
			} else {
				log.Errorln(err)
			}
		}

		handlers.JSONError(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
	})
}
