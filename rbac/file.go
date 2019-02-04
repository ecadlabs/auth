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

	for name, role := range data.Roles {
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

	res := StaticRBAC{
		Roles:       roles,
		Permissions: data.Permissions,
	}

	return &res, nil
}
