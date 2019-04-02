package rbac

import (
	"context"
)

type PermissionDesc struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Roles       []string `json:"roles,omitempty"`
}

type RoleDesc struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions,omitempty"`
}

type RoleDB interface {
	GetDefaultRole() string
	GetRolesDesc(ctx context.Context, perm ...string) ([]*RoleDesc, error)
	GetPermissionsDesc(ctx context.Context, role ...string) ([]*PermissionDesc, error)
	GetRoleDesc(context.Context, string) (*RoleDesc, error)
	GetPermissionDesc(context.Context, string) (*PermissionDesc, error)
}
