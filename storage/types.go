package storage

import (
	"time"

	"github.com/satori/go.uuid"
)

type SortOrder int

const (
	SortAsc SortOrder = iota
	SortDesc
)

const (
	OwnerMembership  = "owner"
	MemberMembership = "member"
)

type Roles map[string]interface{}

type Membership struct {
	Membership_type string    `json:"type"`
	TenantID        uuid.UUID `json:"tenantID"`
}

type User struct {
	ID               uuid.UUID     `json:"id" schema:"id"`
	Email            string        `json:"email" schema:"email"`
	EmailGen         int           `json:"-"`
	PasswordHash     []byte        `json:"-" schema:"-"`
	Name             string        `json:"name,omitempty" schema:"name"`
	Added            time.Time     `json:"added" schema:"added"`
	Modified         time.Time     `json:"modified" schema:"modified"`
	EmailVerified    bool          `json:"email_verified" schema:"email_verified"`
	Roles            Roles         `json:"roles,omitempty" schema:"roles"`
	Memberships      []*Membership `json:"memberships"`
	PasswordGen      int           `json:"-"`
	LoginAddr        string        `json:"login_addr,omitempty"`
	LoginTimestamp   *time.Time    `json:"login_ts,omitempty"`
	RefreshAddr      string        `json:"refresh_addr,omitempty"`
	RefreshTimestamp *time.Time    `json:"refresh_ts,omitempty"`
}

func (u *User) IsMember(id uuid.UUID) bool {
	for _, mem := range u.Memberships {
		if mem.TenantID == id {
			return true
		}
	}
	return false
}

func (u *User) IsOwner(id uuid.UUID) bool {
	for _, mem := range u.Memberships {
		if mem.TenantID == id && mem.Membership_type == OwnerMembership {
			return true
		}
	}
	return false
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

func (r Roles) Get() (roles []string) {
	roles = make([]string, 0, len(r))

	for key := range r {
		roles = append(roles, key)
	}

	return
}

const (
	UsersDefaultSortColumn   = "added"
	TenantsDefaultSortColumn = "added"
	LogDefaultSortColumn     = "ts"
)
