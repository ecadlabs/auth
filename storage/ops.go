package storage

import (
	"fmt"
	"strings"

	"github.com/ecadlabs/auth/errors"
	"github.com/ecadlabs/auth/jsonpatch"
)

type UserOps struct {
	Update      map[string]interface{}
	AddRoles    []string
	RemoveRoles []string
}

type TenantOps struct {
	Update map[string]interface{}
}

func errPatchOp(o *jsonpatch.Op) error {
	return errors.Wrap(fmt.Errorf("Incorrect JSON patch op `%s' for path `%s'", o.Op, o.Path), errors.CodePatchFormat)
}

func TenantOpsFromPatch(patch jsonpatch.Patch) (*TenantOps, error) {
	ret := TenantOps{
		Update: make(map[string]interface{}, len(patch)),
	}

	for _, o := range patch {
		if o.Path != "" && o.Op == "replace" && o.Path[0] == '/' && strings.IndexByte(o.Path[1:], '/') < 0 {
			if o.Value == nil {
				return nil, errors.ErrPatchValue
			}

			ret.Update[o.Path[1:]] = o.Value
		} else {
			return nil, errPatchOp(o)
		}
	}

	return &ret, nil
}

func OpsFromPatch(patch jsonpatch.Patch) (*UserOps, error) {
	ret := UserOps{
		Update:      make(map[string]interface{}, len(patch)),
		AddRoles:    make([]string, 0, len(patch)),
		RemoveRoles: make([]string, 0, len(patch)),
	}

	for _, o := range patch {
		if strings.HasPrefix(o.Path, "/roles/") {
			role := strings.TrimPrefix(o.Path, "/roles/")

			switch o.Op {
			case "add":
				ret.AddRoles = append(ret.AddRoles, role)
			case "remove":
				ret.RemoveRoles = append(ret.RemoveRoles, role)
			default:
				return nil, errPatchOp(o)
			}
		} else if o.Path != "" && o.Op == "replace" && o.Path[0] == '/' && strings.IndexByte(o.Path[1:], '/') < 0 {
			if o.Value == nil {
				return nil, errors.ErrPatchValue
			}

			ret.Update[o.Path[1:]] = o.Value
		} else {
			return nil, errPatchOp(o)
		}
	}

	return &ret, nil
}
