package rbac

import (
	"context"
	"sort"
	"strings"

	"github.com/ecadlabs/auth/errors"
)

type Role interface {
	Name() string
	IsAnyGranted(...string) (bool, error)
	IsAllGranted(...string) (bool, error)
	Permissions() []string
}

// e.g. user data or parsed token
type Subject interface {
	Role() (Role, error)
}

// e.g. database
type Enforcer interface {
	GetRole(ctx context.Context, ids ...string) (Role, error)
}

type RBAC interface {
	RoleDB
	Enforcer
}

type StaticRole struct {
	RoleName        string
	Description     string
	RolePermissions map[string]struct{}
}

func (s *StaticRole) Permissions() []string {
	res := make([]string, 0, len(s.RolePermissions))
	for p := range s.RolePermissions {
		res = append(res, p)
	}
	sort.Strings(res)

	return res
}

func (s *StaticRole) Name() string {
	return s.RoleName
}

func (s *StaticRole) IsAllGranted(perm ...string) (bool, error) {
	for _, p := range perm {
		if _, ok := s.RolePermissions[p]; !ok {
			return false, nil
		}
	}

	return true, nil
}

func (s *StaticRole) IsAnyGranted(perm ...string) (bool, error) {
	for _, p := range perm {
		if _, ok := s.RolePermissions[p]; ok {
			return true, nil
		}
	}

	return false, nil
}

type RoleList []Role

func (r RoleList) Name() string {
	names := make([]string, len(r))
	for i, role := range r {
		names[i] = role.Name()
	}

	return "[" + strings.Join(names, ",") + "]"
}

func (r RoleList) IsAllGranted(perm ...string) (bool, error) {
PermLoop:
	for _, p := range perm {
		for _, role := range r {
			ok, err := role.IsAllGranted(p)
			if err != nil {
				return false, err
			}

			if ok {
				continue PermLoop
			}
		}

		return false, nil
	}

	return true, nil
}

func (r RoleList) IsAnyGranted(perm ...string) (bool, error) {
	for _, p := range perm {
		for _, role := range r {
			if ok, err := role.IsAnyGranted(p); ok || err != nil {
				return ok, err
			}
		}
	}

	return false, nil
}

func (r RoleList) Permissions() []string {
	tmp := make(map[string]struct{})
	for _, role := range r {
		for _, p := range role.Permissions() {
			tmp[p] = struct{}{}
		}
	}

	res := make([]string, 0, len(tmp))
	for p := range tmp {
		res = append(res, p)
	}
	sort.Strings(res)

	return res
}

type StaticRBAC struct {
	Roles       map[string]*StaticRole
	Permissions map[string]string
	DefaultRole string
}

func (s *StaticRBAC) GetDefaultRole() string {
	return s.DefaultRole
}

func (s *StaticRBAC) GetRole(ctx context.Context, ids ...string) (Role, error) {
	res := make([]Role, 0, len(ids))

	for _, id := range ids {
		if r, ok := s.Roles[id]; ok {
			res = append(res, r)
		}
	}

	if len(res) == 0 {
		return nil, errors.ErrRoleNotFound
	} else if len(res) == 1 {
		return res[0], nil
	}

	return RoleList(res), nil
}

func (s *StaticRBAC) GetRolesDesc(ctx context.Context, perm ...string) ([]*RoleDesc, error) {
	roles := make([]*RoleDesc, 0, len(s.Roles))

RolesLoop:
	for _, r := range s.Roles {
		for _, p := range perm {
			if _, ok := r.RolePermissions[p]; !ok {
				continue RolesLoop
			}
		}

		desc := RoleDesc{
			Name:        r.RoleName,
			Description: r.Description,
			Permissions: r.Permissions(),
		}

		roles = append(roles, &desc)
	}

	sort.Slice(roles, func(i int, j int) bool {
		return roles[i].Name < roles[j].Name
	})

	return roles, nil
}

func (s *StaticRBAC) GetPermissionsDesc(ctx context.Context, role ...string) ([]*PermissionDesc, error) {
	perms := make([]*PermissionDesc, 0, len(s.Permissions))

PermissionsLoop:
	for p, d := range s.Permissions {
		rolesList := make(map[string]struct{})

		// build roles list
		for _, r := range s.Roles {
			if _, ok := r.RolePermissions[p]; ok {
				rolesList[r.RoleName] = struct{}{}
			}
		}

		// filter
		for _, r := range role {
			if _, ok := rolesList[r]; !ok {
				continue PermissionsLoop
			}
		}

		desc := PermissionDesc{
			Name:        p,
			Description: d,
			Roles:       make([]string, 0, len(rolesList)),
		}

		for r := range rolesList {
			desc.Roles = append(desc.Roles, r)
		}
		sort.Strings(desc.Roles)

		perms = append(perms, &desc)
	}

	sort.Slice(perms, func(i int, j int) bool {
		return perms[i].Name < perms[j].Name
	})

	return perms, nil
}

func (s *StaticRBAC) GetRoleDesc(ctx context.Context, role string) (*RoleDesc, error) {
	r, ok := s.Roles[role]
	if !ok {
		return nil, errors.ErrRoleNotFound
	}

	desc := RoleDesc{
		Name:        r.RoleName,
		Description: r.Description,
		Permissions: r.Permissions(),
	}

	return &desc, nil
}

func (s *StaticRBAC) GetPermissionDesc(ctx context.Context, perm string) (*PermissionDesc, error) {
	d, ok := s.Permissions[perm]
	if !ok {
		return nil, errors.ErrPermissionNotFound
	}

	rolesList := make(map[string]struct{})

	// build roles list
	for _, r := range s.Roles {
		if _, ok := r.RolePermissions[perm]; ok {
			rolesList[r.RoleName] = struct{}{}
		}
	}

	desc := PermissionDesc{
		Name:        perm,
		Description: d,
		Roles:       make([]string, 0, len(rolesList)),
	}

	for r := range rolesList {
		desc.Roles = append(desc.Roles, r)
	}
	sort.Strings(desc.Roles)

	return &desc, nil
}

var (
	_ Role = &StaticRole{}
	_ Role = RoleList{}
	_ RBAC = &StaticRBAC{}
)
