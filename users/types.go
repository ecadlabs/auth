package users

import (
	"errors"
	"github.com/satori/go.uuid"
	"net/http"
	"strings"
	"time"
)

type Error struct {
	error
	HTTPStatus int
}

var (
	ErrNotFound   = Error{errors.New("User not found"), http.StatusNotFound}
	ErrEmail      = Error{errors.New("Email is in use"), http.StatusConflict}
	ErrPatchValue = Error{errors.New("Patch value is missed"), http.StatusBadRequest}
	ErrRoleExists = Error{errors.New("Role exists"), http.StatusConflict}
)

type SortOrder int

const (
	SortAsc SortOrder = iota
	SortDesc
)

type Roles map[string]interface{}

func (r Roles) HasPrefix(prefix string) bool {
	for role := range r {
		if strings.HasPrefix(role, prefix) {
			return true
		}
	}

	return false
}

func (r Roles) Prefixed(prefix string) Roles {
	ret := make(Roles)
	for role := range r {
		if strings.HasPrefix(role, prefix) {
			ret[role] = struct{}{}
		}
	}
	return ret
}

func (r Roles) Has(role string) bool {
	_, ok := r[role]
	return ok
}

func (r *Roles) Add(role string) {
	if *r == nil {
		*r = make(Roles)
	}
	(*r)[role] = true
}

func (r Roles) Delete(role string) {
	delete(r, role)
}

type User struct {
	ID            uuid.UUID `json:"id" schema:"id"`
	Email         string    `json:"email" schema:"email"`
	PasswordHash  []byte    `json:"-" schema:"-"`
	Password      string    `json:"password,omitempty" schema:"password"` // Create user request
	Name          string    `json:"name,omitempty" schema:"name"`
	Added         time.Time `json:"added" schema:"added"`
	Modified      time.Time `json:"modified" schema:"modified"`
	EmailVerified bool      `json:"email_verified" schema:"email_verified"`
	Roles         Roles     `json:"roles,omitempty" schema:"roles"`
}

const (
	ColumnID       = "id"
	ColumnEmail    = "email"
	ColumnName     = "name"
	ColumnAdded    = "added"
	ColumnModified = "modified"
)

const DefaultSortColumn = ColumnAdded
