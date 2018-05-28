package users

import (
	"errors"
	"github.com/satori/go.uuid"
	"time"
)

var (
	ErrNotFound = errors.New("User not found")
	ErrEmail    = errors.New("Email is in use")
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

const (
	ColumnID       = "id"
	ColumnEmail    = "email"
	ColumnName     = "name"
	ColumnAdded    = "added"
	ColumnModified = "modified"
)

const DefaultSortColumn = ColumnAdded
