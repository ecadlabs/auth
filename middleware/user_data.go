package middleware

import (
	"context"
	"net/http"

	"github.com/ecadlabs/auth/errors"
	"github.com/ecadlabs/auth/storage"
	"github.com/ecadlabs/auth/utils"
	"github.com/dgrijalva/jwt-go"
	"github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
)

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
