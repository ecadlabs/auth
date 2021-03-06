package middleware

import (
	"context"
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/ecadlabs/auth/errors"
	"github.com/ecadlabs/auth/storage"
	"github.com/ecadlabs/auth/utils"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
)

type userContextKey struct{}

var UserContextKey interface{} = userContextKey{}

// Gets user data from DB
type UserData struct {
	Storage storage.UserStorage
}

func (u *UserData) Handler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error

		if token, ok := r.Context().Value(TokenContextKey).(*jwt.Token); ok {
			claims := token.Claims.(jwt.MapClaims)

			if sub, ok := claims["sub"].(string); ok {
				var id uuid.UUID

				if id, err = uuid.FromString(sub); err == nil {
					var user *storage.User

					if user, err = u.Storage.GetUserByID(r.Context(), "", id); err == nil {
						if user.Type == storage.AccountService || user.EmailVerified {
							req := r.WithContext(context.WithValue(r.Context(), UserContextKey, user))
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
