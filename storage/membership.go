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
	UserID           uuid.UUID      `db:"user_id"`
	TenantID         uuid.UUID      `db:"tenant_id"`
	MembershipType   string         `db:"membership_type"`
	MembershipStatus string         `db:"membership_status"`
	Added            time.Time      `db:"added"`
	Modified         time.Time      `db:"modified"`
	Roles            pq.StringArray `db:"roles"`
	SortedBy         string         `db:"sorted_by"`
	Email            string         `db:"email"`
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
		Email:            m.Email,
		Roles:            make(Roles, len(m.Roles)),
	}

	for _, r := range m.Roles {
		ret.Roles[r] = true
	}

	return ret
}

// AddMembership insert a new membership in the database
func (s *Storage) AddMembership(ctx context.Context, id uuid.UUID, user *User, status string, membershipType string, role Roles) error {
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

	err = s.AddMembershipInt(ctx, tx, id, user.ID, status, membershipType, role)
	if err != nil {
		return err
	} else {
		return nil
	}
}

// AddMembershipInt insert a new membership in the database
func (s *Storage) AddMembershipInt(ctx context.Context, tx *sqlx.Tx, tenantID uuid.UUID, userID uuid.UUID, status string, membershipType string, role Roles) error {
	var mid uuid.UUID
	err := tx.GetContext(ctx, &mid, "INSERT INTO membership (tenant_id, user_id, membership_status, membership_type) VALUES ($1, $2, $3, $4) RETURNING id", tenantID, userID, status, membershipType)

	if err != nil {
		if isUniqueViolation(err, "membership_pkey") {
			err = errors.ErrMembershipExisits
		}
		return err
	}

	// Create roles
	valuesExprs := make([]string, len(role))
	args := make([]interface{}, len(role)+1)

	args[0] = mid
	var i int

	for r := range role {
		valuesExprs[i] = fmt.Sprintf("($1, $%d)", i+2)
		args[i+1] = r
		i++
	}

	_, err = tx.ExecContext(ctx, "INSERT INTO roles (membership_id, role) VALUES "+strings.Join(valuesExprs, ", "), args...)

	if err != nil {
		if isUniqueViolation(err, "membership_pkey") {
			err = errors.ErrRoleExists
		}
		return err
	}

	return nil
}

// GetMembership retrive a membership from the database
func (s *Storage) GetMembership(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*Membership, error) {
	q := `
	SELECT
	  membership.*,
	  r.roles,
	  users.email
	FROM
	  membership
	  INNER JOIN users ON membership.user_id = users.id
	  LEFT JOIN (
	    SELECT
	      membership_id,
	      array_agg(role) AS roles
	    FROM
	      roles
	    GROUP BY
	      membership_id
	  ) AS r ON r.membership_id = membership.id
	WHERE
	  membership.tenant_id = $1
	  AND membership.user_id = $2`

	var model membershipModel
	err := s.DB.GetContext(ctx, &model, q, id, userID)

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
func (s *Storage) UpdateMembership(ctx context.Context, id uuid.UUID, userID uuid.UUID, ops *Ops) (*Membership, error) {
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
	if roles := ops.Add["roles"]; len(roles) != 0 {
		expr := "INSERT INTO roles (membership_id, role) VALUES "
		args := make([]interface{}, len(roles)+1)

		args[0] = u.ID

		for i, r := range roles {
			if i != 0 {
				expr += ", "
			}
			expr += fmt.Sprintf("($1, $%d)", i+2)
			args[i+1] = r
		}

		if _, err = tx.ExecContext(ctx, expr, args...); err != nil {
			if isUniqueViolation(err, "roles_membership_id_role_key") {
				err = errors.ErrRoleExists
			}
			return nil, err
		}
	}

	if roles := ops.Remove["roles"]; len(roles) != 0 {
		expr := "DELETE FROM roles WHERE membership_id = $1 AND ("
		args := make([]interface{}, len(roles)+1)

		args[0] = u.ID

		for i, r := range roles {
			if i != 0 {
				expr += " OR "
			}
			expr += fmt.Sprintf("role = $%d", i+2)
			args[i+1] = r
		}

		expr += ")"

		if _, err = tx.ExecContext(ctx, expr, args...); err != nil {
			return nil, err
		}
	}

	// Get roles back
	if err = tx.GetContext(ctx, &u, "SELECT array_agg(role) AS roles FROM roles WHERE membership_id = $1 GROUP BY membership_id", u.ID); err != nil {
		if err != sql.ErrNoRows {
			return nil, err
		}
		return nil, errors.ErrRolesEmpty
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
	"roles":             query.ColQuery,
}

// GetMemberships get memberships from the database as a paged result
func (s *Storage) GetMemberships(ctx context.Context, q *query.Query) (memberships []*Membership, count int, next *query.Query, err error) {
	if q.SortBy == "" {
		q.SortBy = MembershipsDefaultSortColumn
	}

	selOpt := query.SelectOptions{
		SelectExpr: "membership.*, r.roles, users.email, membership." + pq.QuoteIdentifier(q.SortBy) + " AS sorted_by",
		FromExpr: `
			membership
			INNER JOIN tenants ON tenants.id = membership.tenant_id
			AND tenants.archived = FALSE
			INNER JOIN users ON users.id = membership.user_id
			LEFT JOIN (
		  	SELECT
		    	membership_id,
		    	array_agg(role) AS roles
		  	FROM
		    	roles
		  	GROUP BY
		    	membership_id
				) AS r ON r.membership_id = membership.id`,
		IDColumn: "id",
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
func (s *Storage) DeleteMembership(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
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

	return nil
}
