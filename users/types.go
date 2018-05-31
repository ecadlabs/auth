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

type Role int

const (
	RoleRegular Role = iota
	RoleAdmin
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
	Role          Role      `json:"role,omitempty" schema:"role"`
	EmailVerified bool      `json:"email_verified" schema:"email_verified"`
}

func (u *User) ColumnToString(column string) string {
	switch column {
	case ColumnID:
		return u.ID.String()
	case ColumnEmail:
		return u.Email
	case ColumnName:
		return u.Name
	case ColumnAdded:
		v, _ := u.Added.MarshalText()
		return string(v)
	case ColumnModified:
		v, _ := u.Modified.MarshalText()
		return string(v)
	default:
		return ""
	}
}

const (
	ColumnID       = "id"
	ColumnEmail    = "email"
	ColumnName     = "name"
	ColumnAdded    = "added"
	ColumnModified = "modified"
)

const DefaultSortColumn = ColumnAdded
