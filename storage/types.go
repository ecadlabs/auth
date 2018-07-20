package storage

import (
	"strings"
	"time"

	"git.ecadlabs.com/ecad/auth/roles"
	"github.com/satori/go.uuid"
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

func (r Roles) Get() roles.Roles {
	s := make([]string, len(r))
	var i int
	for role := range r {
		s[i] = role
		i++
	}
	return roles.GetKnownRoles(s)
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
	ID               uuid.UUID  `json:"id" schema:"id"`
	Email            string     `json:"email" schema:"email"`
	EmailGen         int        `json:"-"`
	PasswordHash     []byte     `json:"-" schema:"-"`
	Name             string     `json:"name,omitempty" schema:"name"`
	Added            time.Time  `json:"added" schema:"added"`
	Modified         time.Time  `json:"modified" schema:"modified"`
	EmailVerified    bool       `json:"email_verified" schema:"email_verified"`
	Roles            Roles      `json:"roles,omitempty" schema:"roles"`
	PasswordGen      int        `json:"-"`
	LoginAddr        string     `json:"login_addr,omitempty"`
	LoginTimestamp   *time.Time `json:"login_ts,omitempty"`
	RefreshAddr      string     `json:"refresh_addr,omitempty"`
	RefreshTimestamp *time.Time `json:"refresh_ts,omitempty"`
}

type LogEntry struct {
	ID        uuid.UUID              `json:"id"`
	Timestamp time.Time              `json:"ts"`
	Event     string                 `json:"event"`
	UserID    uuid.UUID              `json:"user_id,omitempty"`
	TargerID  uuid.UUID              `json:"target_id,omitempty"`
	Address   string                 `json:"addr,omitempty"`
	Message   string                 `json:"msg,omitempty"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

const (
	UsersDefaultSortColumn = "added"
	LogDefaultSortColumn   = "ts"
)
