package storage

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/ecadlabs/auth/errors"
	"github.com/ecadlabs/auth/query"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	uuid "github.com/satori/go.uuid"
)

type membershipModel struct {
	ID               uuid.UUID      `db:"id"`
	MembershipType   string         `db:"membership_type"`
	TenantID         uuid.UUID      `db:"tenant_id"`
	MembershipStatus string         `db:"membership_status"`
	UserID           uuid.UUID      `db:"user_id"`
	Added            time.Time      `db:"added"`
	Modified         time.Time      `db:"modified"`
	Roles            pq.StringArray `db:"roles"`
	SortedBy         string         `json:"-" db:"sorted_by"`
}

func (m *membershipModel) toMembership() *Membership {
	ret := &Membership{
		ID:               m.ID,
		MembershipStatus: m.MembershipStatus,
		MembershipType:   m.MembershipType,
		UserID:           m.UserID,
		TenantID:         m.TenantID,
		Added:            m.Added,
		Modified:         m.Modified,
		Roles:            make(Roles, len(m.Roles)),
	}

	for _, r := range m.Roles {
		ret.Roles[r] = true
	}

	return ret
}

// MembershipStorage is a storage service for memberships
type MembershipStorage struct {
	DB *sqlx.DB
}

// AddMembership insert a new membership in the database
func (s *MembershipStorage) AddMembership(ctx context.Context, id uuid.UUID, user *User, status string, membershipType string, role Roles) error {
	// TODO: Refactor this to use the membership struct as a parameter
	tx, err := s.DB.Beginx()
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}

		err = tx.Commit()
	}()

	_, err = tx.ExecContext(ctx, "INSERT INTO membership (tenant_id, user_id, membership_status, membership_type) VALUES ($1, $2, $3, $4)", id, user.ID, status, membershipType)

	if err != nil {
		if isUniqueViolation(err, "membership_pkey") {
			err = errors.ErrMembershipExisits
		}
		return err
	}

	// Create roles
	valuesExprs := make([]string, len(role))
	args := make([]interface{}, len(role)+2)

	args[0] = user.ID
	args[1] = id
	var i int

	for r := range role {
		valuesExprs[i] = fmt.Sprintf("($1, $2, $%d)", i+3)
		args[i+2] = r
		i++
	}

	_, err = tx.ExecContext(ctx, "INSERT INTO roles (user_id, tenant_id, role) VALUES "+strings.Join(valuesExprs, ", "), args...)

	if err != nil {
		if isUniqueViolation(err, "membership_pkey") {
			err = errors.ErrRoleExists
		}
		return err
	}

	return nil
}

// GetMembership retrive a membership from the database
func (s *MembershipStorage) GetMembership(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*Membership, error) {
	model := membershipModel{}
	err := s.DB.GetContext(ctx, &model, "SELECT membership.*, ra.roles FROM membership LEFT JOIN (SELECT user_id, tenant_id, array_agg(role) AS roles FROM roles GROUP BY user_id, tenant_id) AS ra ON ra.user_id = membership.user_id AND ra.tenant_id = membership.tenant_id WHERE membership.tenant_id = $1 AND membership.user_id = $2", id, userID)

	if err != nil {
		if err == sql.ErrNoRows {
			err = errors.ErrMembershipNotFound
		}
		return nil, err
	}

	return model.toMembership(), nil
}

var membershipUpdatePath = map[string]struct{}{
	"membership_type":   struct{}{},
	"membership_status": struct{}{},
}

// UpdateMembership update a membership
func (s *MembershipStorage) UpdateMembership(ctx context.Context, id uuid.UUID, userID uuid.UUID, ops *RoleOps) (*Membership, error) {
	// Verify columns
	for k := range ops.Update {
		if _, ok := membershipUpdatePath[k]; !ok {
			return nil, errPatchPath(k)
		}
	}

	tx, err := s.DB.Beginx()
	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}

		err = tx.Commit()
	}()

	// Update properties
	var i int
	expr := "UPDATE membership SET "
	args := make([]interface{}, len(ops.Update)+2)

	for k, v := range ops.Update {
		if i != 0 {
			expr += ", "
		}
		expr += fmt.Sprintf("%s = $%d", pq.QuoteIdentifier(k), i+1)
		args[i] = v
		i++
	}

	if len(ops.Update) != 0 {
		expr += ", "
	}

	updateCount := len(ops.Update)
	expr += fmt.Sprintf("modified = DEFAULT WHERE user_id = $%d AND tenant_id = $%d RETURNING *", updateCount+1, updateCount+2)

	args[updateCount] = userID
	args[updateCount+1] = id

	var u membershipModel
	if err = tx.GetContext(ctx, &u, expr, args...); err != nil {
		if err == sql.ErrNoRows {
			err = errors.ErrUserNotFound
		}
		return nil, err
	}

	// Update roles
	if len(ops.AddRoles) != 0 {
		expr := "INSERT INTO roles (user_id, tenant_id, role) VALUES "
		args := make([]interface{}, len(ops.AddRoles)+2)

		for i, r := range ops.AddRoles {
			if i != 0 {
				expr += ", "
			}
			expr += fmt.Sprintf("($1, $2, $%d)", i+3)
			args[i+2] = r
		}

		args[0] = userID
		args[1] = id

		if _, err = tx.ExecContext(ctx, expr, args...); err != nil {
			if isUniqueViolation(err, "roles_pkey") {
				err = errors.ErrRoleExists
			}
			return nil, err
		}
	}

	if len(ops.RemoveRoles) != 0 {
		expr := "DELETE FROM roles WHERE user_id = $1 AND tenant_id = $2 AND ("
		args := make([]interface{}, len(ops.RemoveRoles)+2)

		for i, r := range ops.RemoveRoles {
			if i != 0 {
				expr += " OR "
			}
			expr += fmt.Sprintf("role = $%d", i+3)
			args[i+2] = r
		}

		expr += ")"
		args[0] = userID
		args[1] = id

		if _, err = tx.ExecContext(ctx, expr, args...); err != nil {
			return nil, err
		}
	}

	// Get roles back

	if err = tx.GetContext(ctx, &u, "SELECT array_agg(role) AS roles FROM roles WHERE user_id = $1 AND tenant_id = $2 GROUP BY user_id, tenant_id", userID, id); err != nil {
		if err != sql.ErrNoRows {
			return nil, err
		}
		return nil, errors.ErrRolesEmpty
	}

	// Safe guard to always have one owner
	err = s.hasAMinimumOfOneOwner(ctx, tx, id)
	if err != nil {
		return nil, err
	}

	return u.toMembership(), nil
}

var membershipsQueryColumns = map[string]int{
	"user_id":           query.ColQuery | query.ColSort,
	"tenant_id":         query.ColQuery | query.ColSort,
	"added":             query.ColQuery | query.ColSort,
	"modified":          query.ColQuery | query.ColSort,
	"membership_type":   query.ColQuery | query.ColSort,
	"membership_status": query.ColQuery | query.ColSort,
}

// GetMemberships get memberships from the database as a paged result
func (s *MembershipStorage) GetMemberships(ctx context.Context, q *query.Query) (memberships []*Membership, count int, next *query.Query, err error) {
	if q.SortBy == "" {
		q.SortBy = MembershipsDefaultSortColumn
	}

	selOpt := query.SelectOptions{
		SelectExpr: "mem_role.*, " + pq.QuoteIdentifier(q.SortBy) + " AS sorted_by",
		FromExpr:   "(SELECT mem.*, ra.roles FROM (SELECT membership.* FROM membership LEFT JOIN tenants ON membership.tenant_id = tenants.id WHERE tenants.archived = FALSE) AS mem LEFT JOIN (SELECT user_id, tenant_id, array_agg(role) AS roles FROM roles GROUP BY user_id, tenant_id) AS ra ON ra.user_id = mem.user_id AND ra.tenant_id = mem.tenant_id) AS mem_role",
		IDColumn:   "id",
		ColumnFlagsFunc: func(col string) int {
			if flags, ok := membershipsQueryColumns[col]; ok {
				return flags
			}
			return 0
		},
	}

	stmt, args, err := q.SelectStmt(&selOpt)
	if err != nil {
		err = errors.Wrap(err, errors.CodeQuerySyntax)
		return
	}

	rows, err := s.DB.QueryxContext(ctx, stmt, args...)
	if err != nil {
		return
	}
	defer rows.Close()

	membershipsSlice := []*Membership{}
	var lastItem *membershipModel

	for rows.Next() {
		var membership membershipModel
		if err = rows.StructScan(&membership); err != nil {
			return
		}

		lastItem = &membership
		membershipsSlice = append(membershipsSlice, membership.toMembership())
	}

	if err = rows.Err(); err != nil {
		return
	}

	// Count
	if q.TotalCount {
		stmt, args := q.CountStmt(&selOpt)
		if err = s.DB.Get(&count, stmt, args...); err != nil {
			return
		}
	}

	memberships = membershipsSlice

	if lastItem != nil {
		// Update query
		ret := *q
		ret.LastID = lastItem.ID.String()
		ret.Last = lastItem.SortedBy
		ret.TotalCount = false

		next = &ret
	}

	return
}

// DeleteMembership deletes a membership from the database
func (s *MembershipStorage) DeleteMembership(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	tx, err := s.DB.Beginx()
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}

		err = tx.Commit()
	}()

	res, err := tx.ExecContext(ctx, "DELETE FROM membership WHERE user_id = $1 AND tenant_id = $2", userID, id)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return errors.ErrMembershipNotFound
	}

	_, err = tx.ExecContext(ctx, "DELETE FROM roles WHERE user_id = $1 AND tenant_id = $2", userID, id)
	if err != nil {
		return err
	}

	// Safe guard to always have one owner
	err = s.hasAMinimumOfOneOwner(ctx, tx, id)
	if err != nil {
		return err
	}

	return nil
}

func (s *MembershipStorage) hasAMinimumOfOneOwner(ctx context.Context, tx *sqlx.Tx, id uuid.UUID) error {
	rows, err := tx.QueryContext(ctx, "SELECT * FROM roles LEFT JOIN membership ON membership.tenant_id = roles.tenant_id AND membership.user_id = roles.user_id WHERE roles.tenant_id = $1 AND membership_status = $2 AND membership_type = $3 AND roles.role = $4", id, ActiveState, OwnerMembership, OwnerMembership)
	if err != nil {
		return err
	}

	defer rows.Close()
	if !rows.Next() {
		return errors.ErrMembershipNotFound
	}

	return nil
}
