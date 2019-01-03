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
	//EvCreate constant for the create user event
	EvCreate = "create"
	//EvCreateTenant constant for the create tenant event
	EvCreateTenant = "create_tenant"
	//EvUpdate constant for the update user event
	EvUpdate = "update"
	//EvUpdateTenant constant for the update tenant event
	EvUpdateTenant = "update_tenant"
	//EvAddRole constant for the add role event
	EvAddRole = "add_role"
	//EvRemoveRole constant for the remove role event
	EvRemoveRole = "remove_role"
	//EvDelete constant for the delete user event
	EvDelete = "delete"
	//EvArchiveTenant constant for the archive tenant event
	EvArchiveTenant = "archive_tenant"
	//EvMembershipDelete constant for the delete membership event
	EvMembershipDelete = "delete_membership"
	//EvReset constant for the reset password event
	EvReset = "reset"
	//EvResetRequest constant for the request reset password event
	EvResetRequest = "reset_request"
	//EvLogin constant for the login event
	EvLogin = "login"
	//EvEmailUpdateRequest constant for request email update event
	EvEmailUpdateRequest = "email_update_request"
	//EvEmailUpdate constant for email update event
	EvEmailUpdate = "email_update"
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
