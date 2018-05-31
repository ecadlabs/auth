package users

import (
	"errors"
	"github.com/satori/go.uuid"
	"net/http"
	"time"
)

type Error struct {
	error
	HTTPStatus int
}

var (
	ErrNotFound = Error{errors.New("User not found"), http.StatusNotFound}
	ErrEmail    = Error{errors.New("Email is in use"), http.StatusConflict}
)

const (
	RoleRegular = "com.ecadlabs.regular"
	RoleAdmin   = "com.ecadlabs.admin"
)

type SortOrder int

const (
	SortAsc SortOrder = iota
	SortDesc
)

type User struct {
	ID            uuid.UUID `json:"id" schema:"id"`
	Email         string    `json:"email" schema:"email"`
	PasswordHash  []byte    `json:"-" schema:"-"`
	Password      string    `json:"password,omitempty" schema:"password"` // Create user request
	Name          string    `json:"name,omitempty" schema:"name"`
	Added         time.Time `json:"added" schema:"added"`
	Modified      time.Time `json:"modified" schema:"modified"`
	EmailVerified bool      `json:"email_verified" schema:"email_verified"`
	Roles         []string  `json:"roles,omitempty" schema:"roles"`
}

const (
	ColumnID       = "id"
	ColumnEmail    = "email"
	ColumnName     = "name"
	ColumnAdded    = "added"
	ColumnModified = "modified"
)

const DefaultSortColumn = ColumnAdded
