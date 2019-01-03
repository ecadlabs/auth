package storage

import (
	"time"

	"github.com/ecadlabs/auth/rbac"
	"github.com/satori/go.uuid"
)

type SortOrder int

const (
	SortAsc SortOrder = iota
	SortDesc
)

const (
	OwnerRole = "owner"
)

const (
	OwnerMembership  = "owner"
	MemberMembership = "member"
)

const (
	InvitedState = "invited"
	ActiveState  = "active"
)

type Roles map[string]interface{}

type MembershipItem struct {
	Membership_type string    `json:"type"`
	TenantID        uuid.UUID `json:"tenantID"`
}

type CreateUser struct {
	Email            string     `json:"email" schema:"email"`
	Name             string     `json:"name,omitempty" schema:"name"`
	ID               uuid.UUID  `json:"id" schema:"id"`
	PasswordHash     []byte     `json:"-" schema:"-"`
	Added            time.Time  `json:"added" schema:"added"`
	Modified         time.Time  `json:"modified" schema:"modified"`
	EmailVerified    bool       `json:"email_verified" schema:"email_verified"`
	LoginAddr        string     `json:"login_addr,omitempty"`
	LoginTimestamp   *time.Time `json:"login_ts,omitempty"`
	RefreshAddr      string     `json:"refresh_addr,omitempty"`
	RefreshTimestamp *time.Time `json:"refresh_ts,omitempty"`
	Roles            Roles      `json:"roles,omitempty`
}

type User struct {
	ID               uuid.UUID         `json:"id" schema:"id"`
	Email            string            `json:"email" schema:"email"`
	EmailGen         int               `json:"-"`
	Name             string            `json:"name,omitempty" schema:"name"`
	PasswordHash     []byte            `json:"-" schema:"-"`
	Added            time.Time         `json:"added" schema:"added"`
	Modified         time.Time         `json:"modified" schema:"modified"`
	EmailVerified    bool              `json:"email_verified" schema:"email_verified"`
	Memberships      []*MembershipItem `json:"memberships"`
	PasswordGen      int               `json:"-"`
	LoginAddr        string            `json:"login_addr,omitempty"`
	LoginTimestamp   *time.Time        `json:"login_ts,omitempty"`
	RefreshAddr      string            `json:"refresh_addr,omitempty"`
	RefreshTimestamp *time.Time        `json:"refresh_ts,omitempty"`
}

func (u *User) GetDefaultMembership() (id uuid.UUID) {
	return u.Memberships[0].TenantID
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

type Membership struct {
	ID                uuid.UUID `json:"id"`
	Membership_type   string    `json:"type"`
	TenantID          uuid.UUID `json:"tenant_id"`
	Membership_status string    `json:"status"`
	UserID            uuid.UUID `json:"user_id"`
	Added             time.Time `json:"added"`
	Modified          time.Time `json:"modified"`
	Roles             Roles     `json:"roles"`
}

func (u *Membership) CanDelegate(role rbac.Role, roles Roles, prefix string) (bool, error) {
	delegate := make([]string, 0, len(roles))
	for r := range u.Roles {
		delegate = append(delegate, prefix+r)
	}

	return role.IsAllGranted(delegate...)
}

type Member struct {
	Email             string
	TenantID          uuid.UUID
	UserID            uuid.UUID
	Membership_type   string
	Membership_status string
	Added             time.Time
	Modified          time.Time
	Roles             Roles
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
	UsersDefaultSortColumn       = "added"
	TenantsDefaultSortColumn     = "added"
	MembershipsDefaultSortColumn = "added"
	LogDefaultSortColumn         = "ts"
)
