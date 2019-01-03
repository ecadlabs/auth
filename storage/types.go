package storage

import (
	"time"

	"github.com/ecadlabs/auth/rbac"
	"github.com/satori/go.uuid"
)

// SortOrder int alias used for sort direction
type SortOrder int

const (
	// SortAsc sortOrder representing ascending sort direction
	SortAsc SortOrder = iota
	// SortDesc sortOrder representing descending sort direction
	SortDesc
)

const (
	// OwnerRole string representing the owner role
	OwnerRole = "owner"
)

const (
	// OwnerMembership string representing the owner membership
	OwnerMembership = "owner"
	// MemberMembership string representing the member membership
	MemberMembership = "member"
)

const (
	// InvitedState string representing the invited membership state
	InvitedState = "invited"
	// ActiveState string representing the active membership state
	ActiveState = "active"
)

// Roles type for holding roles map
type Roles map[string]interface{}

type membershipItem struct {
	MembershipType string    `json:"type"`
	TenantID       uuid.UUID `json:"tenantID"`
}

// CreateUser struct representing data necessary to create a new user
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
	Roles            Roles      `json:"roles,omitempty"`
}

// User struct representing a user
type User struct {
	ID               uuid.UUID         `json:"id" schema:"id"`
	Email            string            `json:"email" schema:"email"`
	EmailGen         int               `json:"-"`
	Name             string            `json:"name,omitempty" schema:"name"`
	PasswordHash     []byte            `json:"-" schema:"-"`
	Added            time.Time         `json:"added" schema:"added"`
	Modified         time.Time         `json:"modified" schema:"modified"`
	EmailVerified    bool              `json:"email_verified" schema:"email_verified"`
	Memberships      []*membershipItem `json:"memberships"`
	PasswordGen      int               `json:"-"`
	LoginAddr        string            `json:"login_addr,omitempty"`
	LoginTimestamp   *time.Time        `json:"login_ts,omitempty"`
	RefreshAddr      string            `json:"refresh_addr,omitempty"`
	RefreshTimestamp *time.Time        `json:"refresh_ts,omitempty"`
}

// GetDefaultMembership retrive the default membership of this user
func (u *User) GetDefaultMembership() (id uuid.UUID) {
	return u.Memberships[0].TenantID
}

// Membership struct representing a user
type Membership struct {
	ID               uuid.UUID `json:"id"`
	MembershipType   string    `json:"type"`
	TenantID         uuid.UUID `json:"tenant_id"`
	MembershipStatus string    `json:"status"`
	UserID           uuid.UUID `json:"user_id"`
	Added            time.Time `json:"added"`
	Modified         time.Time `json:"modified"`
	Roles            Roles     `json:"roles"`
}

// CanDelegate return a boolean if member can delegate this role
func (u *Membership) CanDelegate(role rbac.Role, roles Roles, prefix string) (bool, error) {
	delegate := make([]string, 0, len(roles))
	for r := range u.Roles {
		delegate = append(delegate, prefix+r)
	}

	return role.IsAllGranted(delegate...)
}

// Member struct representing a member
type Member struct {
	Email            string
	TenantID         uuid.UUID
	UserID           uuid.UUID
	MembershipType   string
	MembershipStatus string
	Added            time.Time
	Modified         time.Time
	Roles            Roles
}

// LogEntry struct representing a log entry
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

// Get retrieve a list of roles
func (r Roles) Get() (roles []string) {
	roles = make([]string, 0, len(r))

	for key := range r {
		roles = append(roles, key)
	}

	return
}

const (
	// UsersDefaultSortColumn default column for sorting users
	UsersDefaultSortColumn = "added"
	// TenantsDefaultSortColumn default column for sorting tenants
	TenantsDefaultSortColumn = "added"
	// MembershipsDefaultSortColumn default column for sorting memberships
	MembershipsDefaultSortColumn = "added"
	// LogDefaultSortColumn default column for sorting logs
	LogDefaultSortColumn = "ts"
)
