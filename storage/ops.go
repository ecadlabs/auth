package storage

import (
	"fmt"
	"strings"

	"github.com/ecadlabs/auth/errors"
	"github.com/ecadlabs/auth/jsonpatch"
)

// Ops struct that represent patch
type Ops struct {
	Update map[string]interface{}
	Add    map[string][]string
	Remove map[string][]string
}

func errPatchOp(o *jsonpatch.Op) error {
	return errors.Wrap(fmt.Errorf("Incorrect JSON patch op `%s' for path `%s'", o.Op, o.Path), errors.CodePatchFormat)
}

// OpsFromPatch create a RoleOps struct from a patch request
func OpsFromPatch(patch jsonpatch.Patch) (*Ops, error) {
	ret := Ops{
		Update: make(map[string]interface{}, len(patch)),
		Add:    make(map[string][]string, len(patch)),
		Remove: make(map[string][]string, len(patch)),
	}

	for _, o := range patch {
		if o.Path == "" || o.Path[0] != '/' {
			return nil, errPatchOp(o)
		}

		path := o.Path[1:]

		switch o.Op {
		case "add", "remove":
			p := strings.SplitN(path, "/", 2)
			if len(p) < 2 {
				return nil, errPatchOp(o)
			}

			if o.Op == "add" {
				ret.Add[p[0]] = append(ret.Add[p[0]], p[1])
			} else {
				ret.Remove[p[0]] = append(ret.Remove[p[0]], p[1])
			}

		case "replace":
			if o.Value == nil {
				return nil, errors.ErrPatchValue
			}

			if strings.IndexByte(path, '/') >= 0 {
				return nil, errPatchOp(o)
			}

			ret.Update[path] = o.Value

		default:
			return nil, errPatchOp(o)
		}
	}

	return &ret, nil
}
