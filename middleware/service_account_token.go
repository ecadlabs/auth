package middleware

import (
	"context"
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/ecadlabs/auth/errors"
	"github.com/ecadlabs/auth/storage"
	"github.com/ecadlabs/auth/utils"
	"github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
)

type KeyMembershipStorage interface {
	storage.MembershipStorage
	storage.APIKeyStorage
}

type ServiceAPI struct {
	Storage              KeyMembershipStorage
	TokenContextKey      interface{}
	MembershipContextKey interface{}
	Namespace            string
}

func (s *ServiceAPI) Handler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, ok := r.Context().Value(s.TokenContextKey).(*jwt.Token)
		if !ok {
			utils.JSONError(w, "", errors.CodeUnauthorized)
			return
		}

		claims := token.Claims.(jwt.MapClaims)

		keyIdStr, ok := claims[utils.NSClaim(s.Namespace, "api_key")].(string)
		if !ok {
			// Just regular token
			h.ServeHTTP(w, r)
			return
		}

		kid, err := uuid.FromString(keyIdStr)
		if err != nil {
			utils.JSONError(w, "", errors.CodeUnauthorized)
			return
		}

		sub, ok := claims["sub"].(string)
		uid, err := uuid.FromString(sub)
		if !ok || err != nil {
			utils.JSONError(w, "", errors.CodeUnauthorized)
			return
		}

		key, err := s.Storage.GetKey(r.Context(), uid, kid)
		if err != nil {
			log.Errorln(err)
			utils.JSONError(w, "", errors.CodeUnauthorized)
			return
		}

		membership, err := s.Storage.GetMembership(r.Context(), key.TenantID, uid)
		if err != nil {
			log.Errorln(err)
			utils.JSONError(w, "", errors.CodeUnauthorized)
			return
		}

		req := r.WithContext(context.WithValue(r.Context(), s.MembershipContextKey, membership))
		h.ServeHTTP(w, req)
	})
}
