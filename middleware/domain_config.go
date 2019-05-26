package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/ecadlabs/auth/utils"
	log "github.com/sirupsen/logrus"
)

type DomainConfigData struct {
	SessionMaxAge          time.Duration `yaml:"session_max_age"`
	ResetTokenMaxAge       time.Duration `yaml:"reset_token_max_age"`
	TenantInviteMaxAge     time.Duration `yaml:"tenant_invite_max_age"`
	EmailUpdateTokenMaxAge time.Duration `yaml:"email_update_token_max_age"`
	// TODO different URLs
}

type DomainConfigStorage interface {
	GetDomainConfig(domain string) (*DomainConfigData, error)
}

type DomainConfig struct {
	Storage DomainConfigStorage
}

type domainConfigContextKey struct{}

var DomainConfigContextKey interface{} = domainConfigContextKey{}

func (d *DomainConfig) Handler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var domain string
		if domain = r.FormValue("_domain"); domain == "" {
			if r.URL.IsAbs() {
				domain = r.URL.Host
			} else {
				domain = r.Host
			}
		}

		conf, err := d.Storage.GetDomainConfig(domain)
		if err != nil {
			log.Error(err)
			utils.JSONErrorResponse(w, err)
			return
		}

		req := r.WithContext(context.WithValue(r.Context(), DomainConfigContextKey, conf))
		h.ServeHTTP(w, req)
	})
}
