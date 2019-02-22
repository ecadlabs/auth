package handlers

import (
	"net/http"

	"github.com/ecadlabs/auth/logger"
	"github.com/ecadlabs/auth/utils"
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
	EvEmailUpdate  = "email_update"
	EvNewAPIKey    = "create_api_key"
	EvDeleteAPIKey = "delete_api_key"
)

const (
	MembeshipIdType = "membership"
	TenantIdType    = "tenant"
	UserIdType      = "user"
)

var evSourceTypeMap = map[string]string{
	EvCreate:             MembeshipIdType,
	EvCreateTenant:       MembeshipIdType,
	EvUpdate:             MembeshipIdType,
	EvUpdateTenant:       MembeshipIdType,
	EvAddRole:            MembeshipIdType,
	EvRemoveRole:         MembeshipIdType,
	EvDelete:             MembeshipIdType,
	EvArchiveTenant:      MembeshipIdType,
	EvMembershipDelete:   MembeshipIdType,
	EvReset:              MembeshipIdType,
	EvResetRequest:       UserIdType,
	EvLogin:              MembeshipIdType,
	EvEmailUpdateRequest: UserIdType,
	EvEmailUpdate:        UserIdType,
}

var evTargetTypeMap = map[string]string{
	EvCreate:             UserIdType,
	EvCreateTenant:       TenantIdType,
	EvUpdate:             UserIdType,
	EvUpdateTenant:       TenantIdType,
	EvAddRole:            MembeshipIdType,
	EvRemoveRole:         MembeshipIdType,
	EvDelete:             UserIdType,
	EvArchiveTenant:      TenantIdType,
	EvMembershipDelete:   TenantIdType,
	EvReset:              UserIdType,
	EvResetRequest:       UserIdType,
	EvLogin:              MembeshipIdType,
	EvEmailUpdateRequest: UserIdType,
	EvEmailUpdate:        UserIdType,
}

func logFields(ev string, self, id uuid.UUID, r *http.Request) logrus.Fields {
	d := make(logrus.Fields, 4)

	if id != uuid.Nil {
		d[logger.DefaultTargetIDKey] = id
	}

	if self != uuid.Nil {
		d[logger.DefaultSourceIDKey] = self
	}

	d[logger.TargetIDType] = evTargetTypeMap[ev]
	d[logger.SourceIDType] = evSourceTypeMap[ev]

	d[logger.DefaultEventKey] = ev
	d[logger.DefaultAddrKey] = utils.GetRemoteAddr(r)

	return d
}
