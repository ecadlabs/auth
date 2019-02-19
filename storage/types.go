package storage

import (
	"context"
	"encoding/json"
	"net"
	"sort"
	"time"

	"github.com/ecadlabs/auth/query"
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

const (
	// AccountRegular represents regulat user account
	AccountRegular = "regular"
	// AccountService represents service account
	AccountService = "service"
)

func unmarshalJSONSet(data []byte, dst *map[string]interface{}) error {
	if err := json.Unmarshal(data, dst); err == nil {
		return nil
	}

	var tmp []string
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*dst = make(map[string]interface{}, len(tmp))
	for _, v := range tmp {
		(*dst)[v] = true
	}

	return nil
}

// Roles type for holding roles map
type Roles map[string]interface{}

func (r *Roles) UnmarshalJSON(data []byte) error {
	return unmarshalJSONSet(data, (*map[string]interface{})(r))
}

// StringSet type for holding list of unique strings
type StringSet map[string]interface{}

func (s *StringSet) UnmarshalJSON(data []byte) error {
	return unmarshalJSONSet(data, (*map[string]interface{})(s))
}

type MembershipItem struct {
	Type     string    `json:"type"`
	TenantID uuid.UUID `json:"tenant_id"`
	Roles    Roles     `json:"roles,omitempty"`
}

// CreateUser struct representing data necessary to create a new user
type CreateUser struct {
	Email            string   `json:"email,omitempty" schema:"email"`
	Name             string   `json:"name,omitempty" schema:"name"`
	PasswordHash     []byte   `json:"-" schema:"-"`
	EmailVerified    bool     `json:"email_verified" schema:"email_verified"`
	Roles            Roles    `json:"roles,omitempty"`
	Type             string   `json:"account_type" schema:"account_type"`
	AddressWhiteList []net.IP `json:"address_whitelist"`
}

// User struct representing a user
type User struct {
	ID               uuid.UUID         `json:"id" schema:"id"`
	Type             string            `json:"account_type" schema:"account_type"`
	Email            string            `json:"email,omitempty" schema:"email"`
	EmailGen         int               `json:"-"`
	Name             string            `json:"name,omitempty" schema:"name"`
	PasswordHash     []byte            `json:"-" schema:"-"`
	Added            time.Time         `json:"added" schema:"added"`
	Modified         time.Time         `json:"modified" schema:"modified"`
	EmailVerified    bool              `json:"email_verified" schema:"email_verified"`
	Membership       []*MembershipItem `json:"membership,omitempty"`
	PasswordGen      int               `json:"-"`
	LoginAddr        string            `json:"login_addr,omitempty"`
	LoginTimestamp   *time.Time        `json:"login_ts,omitempty"`
	RefreshAddr      string            `json:"refresh_addr,omitempty"`
	RefreshTimestamp *time.Time        `json:"refresh_ts,omitempty"`
	AddressWhiteList StringSet         `json:"address_whitelist,omitempty"`
}

// GetDefaultMembership retrive the default membership of this user
func (u *User) GetDefaultMembership() (id uuid.UUID) {
	return u.Membership[0].TenantID
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

	sort.Strings(roles)

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

// APIKey represents service account API key
type APIKey struct {
	ID           uuid.UUID `db:"id" json:"id"`
	MembershipID uuid.UUID `db:"membership_id" json:"membership_id"`
	UserID       uuid.UUID `db:"user_id" json:"user_id"`
	TenantID     uuid.UUID `db:"tenant_id" json:"tenant_id"`
	Added        time.Time `db:"added" json:"added"`
}

type TenantStorage interface {
	CreateTenant(ctx context.Context, name string, ownerID uuid.UUID) (*TenantModel, error)
	GetTenant(ctx context.Context, tenantID, userID uuid.UUID, onlySelf bool) (*TenantModel, error)
	GetTenantsSoleMember(ctx context.Context, userID uuid.UUID) (tenants []*TenantModel, err error)
	GetTenants(ctx context.Context, userID uuid.UUID, onlySelf bool, q *query.Query) (tenants []*TenantModel, count int, next *query.Query, err error)
	PatchTenant(ctx context.Context, id uuid.UUID, ops *Ops) (*TenantModel, error)
	DeleteTenant(ctx context.Context, id uuid.UUID) error
}

type MembershipStorage interface {
	AddMembership(ctx context.Context, id uuid.UUID, user *User, status string, membershipType string, role Roles) error
	GetMembership(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*Membership, error)
	UpdateMembership(ctx context.Context, id uuid.UUID, userID uuid.UUID, ops *Ops) (*Membership, error)
	GetMemberships(ctx context.Context, q *query.Query) (memberships []*Membership, count int, next *query.Query, err error)
	DeleteMembership(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
}

type APIKeyStorage interface {
	GetKey(ctx context.Context, keyID, userID uuid.UUID) (*APIKey, error)
	GetKeys(ctx context.Context, uid uuid.UUID) ([]*APIKey, error)
	NewKey(ctx context.Context, userID, tenantID uuid.UUID) (*APIKey, error)
	DeleteKey(ctx context.Context, keyID, userID uuid.UUID) error
}

type UserStorage interface {
	GetUserByID(ctx context.Context, typ string, id uuid.UUID) (*User, error)
	GetUserByEmail(ctx context.Context, typ, email string) (*User, error)
	GetServiceAccountByAddress(ctx context.Context, address string) (*User, error)
	GetUsers(ctx context.Context, typ string, q *query.Query) (users []*User, count int, next *query.Query, err error)
	NewUser(ctx context.Context, user *CreateUser) (res *User, err error)
	UpdateUser(ctx context.Context, typ string, id uuid.UUID, ops *Ops) (user *User, err error)
	DeleteUser(ctx context.Context, typ string, id uuid.UUID) (err error)
	UpdatePasswordWithGen(ctx context.Context, id uuid.UUID, hash []byte, expectedGen int) (err error)
	UpdateEmailWithGen(ctx context.Context, id uuid.UUID, email string, expectedGen int) (user *User, oldEmail string, err error)
	UpdateLoginInfo(ctx context.Context, id uuid.UUID, addr string) error
	UpdateRefreshInfo(ctx context.Context, id uuid.UUID, addr string) error
}

type LogStorage interface {
	GetLogs(ctx context.Context, q *query.Query) (entries []*LogEntry, count int, next *query.Query, err error)
}
