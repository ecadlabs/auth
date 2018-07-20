package handlers

import (
	"git.ecadlabs.com/ecad/auth/roles"
	"git.ecadlabs.com/ecad/auth/storage"
	"github.com/satori/go.uuid"
)

const (
	RoleAnonymous = "com.ecadlabs.auth.anonymous"
	RoleRegular   = "com.ecadlabs.auth.regular"
	RoleAdmin     = "com.ecadlabs.auth.admin"
	RolePrefix    = "com.ecadlabs.auth."
)

const (
	permissionCreate     = "create"
	permissionGet        = "get"
	permissionList       = "list"
	permissionModify     = "modify"
	permissionDelete     = "delete"
	permissionAddRole    = "add_role"
	permissionDeleteRole = "delete_role"
	permissionLog        = "log"
)

func assertNonAdminUser(args map[string]interface{}) bool {
	// Create regular only
	user, ok := args["user"].(*storage.User)
	return ok && !user.Roles.Has(RoleAdmin)
}

func assertSelf(args map[string]interface{}) bool {
	// Self only
	self, ok := args["self"].(uuid.UUID)
	if !ok {
		return false
	}

	id, ok := args["id"].(uuid.UUID)
	return ok && self == id
}

func assertNonAdminRole(args map[string]interface{}) bool {
	// Manipulate non admin roles
	role, ok := args["role"].(string)
	return ok && role != RoleAdmin
}

var (
	roleRegular = roles.NewRole(RoleRegular, map[string]roles.AssertFunc{
		permissionGet:        assertSelf,
		permissionModify:     assertSelf,
		permissionAddRole:    assertNonAdminRole,
		permissionDeleteRole: assertNonAdminRole,
	}, nil)

	roleAdmin = roles.NewRole(RoleAdmin, map[string]roles.AssertFunc{
		permissionCreate:     nil,
		permissionGet:        nil,
		permissionDelete:     func(args map[string]interface{}) bool { return !assertSelf(args) }, // Others
		permissionModify:     nil,
		permissionAddRole:    nil,
		permissionDeleteRole: assertNonAdminRole,
		permissionList:       nil,
		permissionLog:        nil,
	}, nil)
)
