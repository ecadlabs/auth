package rbac

import (
	"bufio"
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

type yamlRole struct {
	Description string   `yaml:"description"`
	Permissions []string `yaml:"permissions"`
	Default     bool     `yaml:"default"`
}

type yamlFile struct {
	Permissions map[string]string    `yaml:"permissions"`
	Roles       map[string]*yamlRole `yaml:"roles"`
}

func LoadYAML(name string) (*StaticRBAC, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	d := yaml.NewDecoder(bufio.NewReader(f))

	var data yamlFile
	if err := d.Decode(&data); err != nil {
		return nil, err
	}

	roles := make(map[string]*StaticRole)

	defaultRoleCount := 0
	var defaultRoleName string

	for name, role := range data.Roles {
		if role.Default {
			defaultRoleCount++
			defaultRoleName = name
		}

		// verify
		perms := make(map[string]struct{})

		for _, p := range role.Permissions {
			if _, ok := data.Permissions[p]; !ok {
				return nil, fmt.Errorf("YAML RBAC: unknown permission %s", p)
			}

			perms[p] = struct{}{}
		}

		role := StaticRole{
			RoleName:        name,
			Description:     role.Description,
			RolePermissions: perms,
		}

		roles[name] = &role
	}

	if defaultRoleCount != 1 {
		return nil, fmt.Errorf("YAML RBAC: Only One default role must be defined, got %d", defaultRoleCount)
	}

	res := StaticRBAC{
		Roles:       roles,
		Permissions: data.Permissions,
		DefaultRole: defaultRoleName,
	}

	return &res, nil
}
