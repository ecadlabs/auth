package handlers

import (
	"net"
	"net/http"
	"strings"

	"github.com/ecadlabs/auth/logger"
	"github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
)

const (
	EvCreate             = "create"
	EvCreateTenant       = "create_tenant"
	EvUpdate             = "update"
	EvUpdateTenant       = "update_tenant"
	EvAddRole            = "add_role"
	EvRemoveRole         = "remove_role"
	EvDelete             = "delete"
	EvArchiveTenant      = "archive_tenant"
	EvReset              = "reset"
	EvResetRequest       = "reset_request"
	EvLogin              = "login"
	EvEmailUpdateRequest = "email_update_request"
	EvEmailUpdate        = "email_update"
)

func getRemoteAddr(r *http.Request) string {
	if fh := r.Header.Get("Forwarded"); fh != "" {
		chunks := strings.Split(fh, ",")

		for _, c := range chunks {
			opts := strings.Split(strings.TrimSpace(c), ";")

			for _, o := range opts {
				v := strings.SplitN(strings.TrimSpace(o), "=", 2)

				if len(v) == 2 && v[0] == "for" {
					if addr := strings.Trim(v[1], "\"[]"); addr != "" {
						return addr
					}
				}
			}
		}
	}

	if xfh := r.Header.Get("X-Forwarded-For"); xfh != "" {
		chunks := strings.Split(xfh, ",")
		for _, c := range chunks {
			if c = strings.Trim(strings.TrimSpace(c), "\"[]"); c != "" {
				return c
			}
		}
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}

	return r.RemoteAddr
}

func logFields(ev string, self, id uuid.UUID, r *http.Request) logrus.Fields {
	d := make(logrus.Fields, 4)

	if id != uuid.Nil {
		d[logger.DefaultTargetIDKey] = id
	}

	if self != uuid.Nil {
		d[logger.DefaultUserIDKey] = self
	}

	d[logger.DefaultEventKey] = ev
	d[logger.DefaultAddrKey] = getRemoteAddr(r)

	return d
}
