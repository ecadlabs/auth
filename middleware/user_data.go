package middleware

import (
	"context"
	"net/http"

	"git.ecadlabs.com/ecad/auth/errors"
	"git.ecadlabs.com/ecad/auth/storage"
	"git.ecadlabs.com/ecad/auth/utils"
	"github.com/dgrijalva/jwt-go"
	"github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
)

// Extracts user data from token itself (only ID and Roles are set)
type TokenUserData struct {
	Namespace       string
	TokenContextKey string
	UserContextKey  string
	DefaultRole     string
	RolePrefix      string
}

func (t *TokenUserData) Handler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := storage.User{
			Roles: make(storage.Roles),
		}
		req := r.WithContext(context.WithValue(r.Context(), t.UserContextKey, &user))

		token, ok := r.Context().Value(t.TokenContextKey).(*jwt.Token)

		if !ok {
			user.Roles[t.DefaultRole] = struct{}{}
			h.ServeHTTP(w, req)
			return
		}

		claims := token.Claims.(jwt.MapClaims)

		if names, ok := claims[utils.NSClaim(t.Namespace, "roles")].([]interface{}); ok {
			for _, name := range names {
				if s, ok := name.(string); ok {
					user.Roles[s] = struct{}{}
				}
			}
		}

		if !user.Roles.HasPrefix(t.RolePrefix) {
			user.Roles[t.DefaultRole] = struct{}{}
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

		utils.JSONError(w, "", errors.CodeUnauthorized)
	})
}

// Gets user data from DB
type UserData struct {
	Storage         *storage.Storage
	TokenContextKey string
	UserContextKey  string
}

func (u *UserData) Handler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error

		if token, ok := r.Context().Value(u.TokenContextKey).(*jwt.Token); ok {
			claims := token.Claims.(jwt.MapClaims)

			if sub, ok := claims["sub"].(string); ok {
				var id uuid.UUID

				if id, err = uuid.FromString(sub); err == nil {
					var user *storage.User

					if user, err = u.Storage.GetUserByID(r.Context(), id); err == nil {
						if user.EmailVerified {
							req := r.WithContext(context.WithValue(r.Context(), u.UserContextKey, user))
							h.ServeHTTP(w, req)
							return
						}
					}
				}
			}
		}

		if err != nil {
			log.Errorln(err)
		}

		utils.JSONError(w, "", errors.CodeUnauthorized)
	})
}
