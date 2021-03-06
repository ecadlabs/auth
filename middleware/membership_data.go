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

type membershipContextKey struct{}

var MembershipContextKey interface{} = membershipContextKey{}

// Gets user data from DB
type MembershipData struct {
	Storage   storage.MembershipStorage
	Namespace string
}

func (m *MembershipData) Handler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := r.Context().Value(MembershipContextKey).(*storage.Membership); ok {
			// Already got from the service account key
			h.ServeHTTP(w, r)
			return
		}

		var err error
		if token, ok := r.Context().Value(TokenContextKey).(*jwt.Token); ok {
			claims := token.Claims.(jwt.MapClaims)

			if sub, ok := claims["sub"].(string); ok {
				var id uuid.UUID

				if id, err = uuid.FromString(sub); err == nil {

					if tenantIDStr, ok := claims[utils.NSClaim(m.Namespace, "tenant")].(string); ok {
						if tenantID, err := uuid.FromString(tenantIDStr); err == nil {
							membership, err := m.Storage.GetMembership(r.Context(), tenantID, id)

							if err == nil {
								req := r.WithContext(context.WithValue(r.Context(), MembershipContextKey, membership))
								h.ServeHTTP(w, req)
								return
							}
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
