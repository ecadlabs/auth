package middleware

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/ecadlabs/auth/errors"
	"github.com/ecadlabs/auth/notification"
	"github.com/ecadlabs/auth/utils"
	log "github.com/sirupsen/logrus"
)

type DomainConfigData struct {
	SessionMaxAge          time.Duration                  `yaml:"session_max_age"`
	ResetTokenMaxAge       time.Duration                  `yaml:"reset_token_max_age"`
	TenantInviteMaxAge     time.Duration                  `yaml:"tenant_invite_max_age"`
	EmailUpdateTokenMaxAge time.Duration                  `yaml:"email_update_token_max_age"`
	BaseURL                string                         `yaml:"base_url"`
	TemplateData           notification.EmailTemplateData `yaml:"template"`
	BaseURLFunc            func() string                  `yaml:"-"` // Testing only
}

func (c *DomainConfigData) GetBaseURL() string {
	if c.BaseURLFunc != nil {
		return c.BaseURLFunc()
	}

	return c.BaseURL
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

		parsedDomain, _, err := net.SplitHostPort(domain)

		if err != nil {
			switch err.(type) {
			case *net.AddrError:
				parsedDomain = domain
			default:
				log.Error(err)
				utils.JSONError(w, err.Error(), errors.CodeBadRequest)
				return
			}
		}

		domain = parsedDomain
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
