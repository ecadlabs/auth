package users

import (
	"errors"
	"github.com/satori/go.uuid"
	"time"
)

var (
	ErrNotFound = errors.New("User not found")
	ErrColumn   = errors.New("Wrong column name")
)

type Role int

const (
	UserRegular Role = iota
	UserAdmin
)

type SortOrder int

const (
	SortAsc SortOrder = iota
	SortDesc
)

type User struct {
	ID            uuid.UUID `json:"id"`
	Email         string    `json:"email"`
	PasswordHash  []byte    `json:"-"`
	Name          string    `json:"name,omitempty"`
	Added         time.Time `json:"added"`
	Modified      time.Time `json:"modified"`
	Role          Role      `json:"role,omitempty"`
	EmailVerified bool      `json:"email_verified"`
}

type Column string

const (
	ColumnID       Column = "id"
	ColumnEmail    Column = "email"
	ColumnName     Column = "name"
	ColumnAdded    Column = "added"
	ColumnModified Column = "modified"
)

var validSortColumns = map[string]struct{}{
	"id":       struct{}{},
	"email":    struct{}{},
	"name":     struct{}{},
	"added":    struct{}{},
	"modified": struct{}{},
}

func ColumnFromString(s string) (Column, error) {
	if _, ok := validSortColumns[s]; !ok {
		return Column(""), ErrColumn
	}

	return Column(s), nil
}
