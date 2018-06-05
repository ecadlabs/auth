package roles

import (
	"errors"
	"fmt"
	"sync"
)

type AssertFunc func(map[string]interface{}) bool

type Role struct {
	name        string
	permissions map[string]AssertFunc
	parent      *Role
}

var noArgs = make(map[string]interface{})

func (r *Role) isGranted(top *Role, perm string, args map[string]interface{}) error {
	if args == nil {
		args = noArgs
	}

	if assert, ok := r.permissions[perm]; ok {
		if assert != nil && !assert(args) {
			return fmt.Errorf("Assertion associated with permission `%s' of role `%s' has been failed", perm, top.Name())
		}
		return nil
	}

	if r.parent != nil {
		return r.parent.isGranted(top, perm, args)
	}

	return fmt.Errorf("Role `%s' doesn't have `%s' permission", top.Name(), perm)
}

func (r *Role) IsGranted(perm string, args map[string]interface{}) error {
	return r.isGranted(r, perm, args)
}

func (r *Role) Name() string {
	return r.name
}

type Registry struct {
	roles map[string]*Role
	sync.RWMutex
}

func NewRegistry() *Registry {
	return &Registry{
		roles: make(map[string]*Role),
	}
}

func (r *Registry) NewRole(name string, permissions map[string]AssertFunc, parent ...*Role) *Role {
	r.Lock()
	defer r.Unlock()

	if role, ok := r.roles[name]; ok {
		return role
	}

	role := &Role{
		name:        name,
		permissions: permissions,
	}

	if len(parent) > 0 {
		role.parent = parent[0]
	}

	r.roles[name] = role

	return role
}

func (r *Registry) GetRole(name string) *Role {
	r.RLock()
	defer r.RUnlock()

	return r.roles[name]
}

type Roles []*Role

func (r *Registry) GetKnownRoles(names []string) Roles {
	roles := make(Roles, 0, len(names))
	for _, n := range names {
		if role := r.GetRole(n); role != nil {
			roles = append(roles, role)
		}
	}
	return roles
}

func (r Roles) IsGranted(perm string, args map[string]interface{}) error {
	if len(r) == 0 {
		return errors.New("Empty roles slice")
	}

	var err error
	for _, role := range r {
		if err = role.IsGranted(perm, args); err == nil {
			return nil
		}
	}

	return err
}

var DefaultRegistry = NewRegistry()

func NewRole(name string, permissions map[string]AssertFunc, parent ...*Role) *Role {
	return DefaultRegistry.NewRole(name, permissions, parent...)
}

func GetRole(name string) *Role {
	return DefaultRegistry.GetRole(name)
}

func GetKnownRoles(names []string) Roles {
	return DefaultRegistry.GetKnownRoles(names)
}
