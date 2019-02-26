package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/ecadlabs/auth/errors"
	"github.com/ecadlabs/auth/query"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	uuid "github.com/satori/go.uuid"
)

// TenantModel struct that represent tenant resource
type TenantModel struct {
	ID         uuid.UUID `json:"id" db:"id"`
	Name       string    `json:"name" db:"name"`
	Added      time.Time `json:"added" db:"added"`
	Modified   time.Time `json:"modified" db:"modified"`
	Protected  bool      `json:"-" db:"protected"`
	Archived   bool      `json:"-" db:"archived"`
	TenantType string    `json:"type" db:"tenant_type"`
	SortedBy   string    `json:"-" db:"sorted_by"`
}

// Clone clone a TenantModel struct
func (t *TenantModel) Clone() *TenantModel {
	return &TenantModel{
		ID:         t.ID,
		Name:       t.Name,
		Added:      t.Added,
		Modified:   t.Modified,
		Protected:  t.Protected,
		Archived:   t.Archived,
		TenantType: t.TenantType,
	}
}

func (s *Storage) CreateTenantInt(ctx context.Context, tx *sqlx.Tx, tenant *TenantModel) (*TenantModel, error) {
	var newTenant TenantModel
	err := sqlx.GetContext(ctx, tx, &newTenant, "INSERT INTO tenants (name) VALUES ($1) RETURNING *", tenant.Name)
	if err != nil {
		return nil, err
	}

	return &newTenant, nil
}

func (s *Storage) CreateTenant(ctx context.Context, name string) (*TenantModel, error) {
	tenant := TenantModel{
		Name: name,
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

	newTenant, err := s.CreateTenantInt(ctx, tx, &tenant)
	return newTenant, err
}

// CreateTenant insert a tenant in the database and return it
func (s *Storage) CreateTenantWithOwner(ctx context.Context, name string, ownerID uuid.UUID) (*TenantModel, error) {
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

	tenant := &TenantModel{
		Name: name,
	}

	newTenant, err := s.CreateTenantInt(ctx, tx, tenant)
	if err != nil {
		return nil, err
	}

	var memId uuid.UUID
	err = tx.GetContext(ctx, &memId, "INSERT INTO membership (tenant_id, user_id, membership_status, membership_type) VALUES ($1, $2, $3, $4) RETURNING id", newTenant.ID, ownerID, ActiveState, OwnerMembership)
	if err != nil {
		return nil, err
	}

	_, err = tx.ExecContext(ctx, "INSERT INTO roles (membership_id, role) VALUES ($1, $2)", memId, OwnerRole)
	if err != nil {
		return nil, err
	}

	return newTenant, nil
}

// GetTenant fetch a tenant from the database and return it
func (s *Storage) GetTenant(ctx context.Context, tenantID, userID uuid.UUID, onlySelf bool) (*TenantModel, error) {
	var queryExtension = ""
	if onlySelf {
		queryExtension += "AND id IN (SELECT tenant_id FROM membership WHERE user_id = '" + userID.String() + "')"
	}
	model := TenantModel{}
	err := s.DB.GetContext(ctx, &model, "SELECT tenants.* FROM tenants WHERE id = $1"+queryExtension, tenantID)

	if err != nil {
		if err == sql.ErrNoRows {
			err = errors.ErrTenantNotFound
		}
		return nil, err
	}

	return &model, nil
}

var tenantsQueryColumns = query.Columns{
	"id":       {Name: "id", Flags: query.ColSort},
	"name":     {Name: "name", Flags: query.ColSort},
	"added":    {Name: "added", Flags: query.ColSort},
	"modified": {Name: "modified", Flags: query.ColSort},
	"archived": {Name: "archived", Flags: query.ColSort},
}

// GetTenantsSoleMember get a list of tenant where the user is the only member
func (s *Storage) GetTenantsSoleMember(ctx context.Context, userID uuid.UUID) (tenants []*TenantModel, err error) {
	rows, err := s.DB.QueryContext(ctx, `
	SELECT * FROM tenants WHERE id IN (
		SELECT membership.tenant_id FROM membership
		WHERE membership.tenant_id IN (
			SELECT tenant_id FROM membership WHERE user_id = $1
		)
	GROUP BY membership.tenant_id
	HAVING COUNT(user_id) = 1
	)`, userID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tenantsSlice := []*TenantModel{}

	for rows.Next() {
		var tenant TenantModel
		if err = rows.Scan(
			&tenant.ID,
			&tenant.Name,
			&tenant.Added,
			&tenant.Modified,
			&tenant.Protected,
			&tenant.Archived,
			&tenant.TenantType,
		); err != nil {
			return
		}

		tenantsSlice = append(tenantsSlice, &tenant)
	}

	if err = rows.Err(); err != nil {
		return
	}

	tenants = tenantsSlice
	return
}

// GetTenants get a list of tenant which are paged
func (s *Storage) GetTenants(ctx context.Context, userID uuid.UUID, onlySelf bool, q *query.Query) (tenants []*TenantModel, count int, next *query.Query, err error) {
	var queryExtension = "tenants as scoped_tenants"
	if onlySelf {
		queryExtension = "(SELECT * FROM tenants WHERE id IN (SELECT tenant_id FROM membership WHERE user_id = '" + userID.String() + "')) as scoped_tenants"
	}

	if q.SortBy == "" {
		q.SortBy = TenantsDefaultSortColumn
	}

	selOpt := query.SelectOptions{
		SelectExpr: "scoped_tenants.*, scoped_tenants." + pq.QuoteIdentifier(q.SortBy) + " AS sorted_by",
		FromExpr:   queryExtension,
		IDColumn:   "id",
		ColumnFunc: tenantsQueryColumns.Func,
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

	tenantsSlice := []*TenantModel{}
	var lastItem *TenantModel

	for rows.Next() {
		var tenant TenantModel
		if err = rows.StructScan(&tenant); err != nil {
			return
		}

		lastItem = &tenant
		tenantsSlice = append(tenantsSlice, &tenant)
	}

	if err = rows.Err(); err != nil {
		return
	}

	// Count
	if q.TotalCount {
		if stmt, args, err = q.CountStmt(&selOpt); err != nil {
			return
		}

		if err = s.DB.Get(&count, stmt, args...); err != nil {
			return
		}
	}

	tenants = tenantsSlice

	if lastItem != nil {
		// Update query
		lastId := lastItem.ID.String()
		ret := *q
		ret.LastID = &lastId
		ret.Last = &lastItem.SortedBy
		ret.TotalCount = false

		next = &ret
	}

	return
}

var tenantUpdatePaths = map[string]struct{}{
	"name": struct{}{},
}

// PatchTenant update a tenant
func (s *Storage) PatchTenant(ctx context.Context, id uuid.UUID, ops *Ops) (*TenantModel, error) {
	// Verify columns
	for k := range ops.Update {
		if _, ok := tenantUpdatePaths[k]; !ok {
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
	expr := "UPDATE tenants SET "
	args := make([]interface{}, len(ops.Update)+1)

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

	expr += fmt.Sprintf("modified = DEFAULT WHERE id = $%d RETURNING *", len(ops.Update)+1)
	args[len(ops.Update)] = id

	var tenant TenantModel
	if err = tx.GetContext(ctx, &tenant, expr, args...); err != nil {
		if err == sql.ErrNoRows {
			err = errors.ErrTenantNotFound
		}
		return nil, err
	}

	return &tenant, nil
}

// DeleteTenant delete a tenant from the database
func (s *Storage) DeleteTenant(ctx context.Context, id uuid.UUID) error {
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

	_, err = tx.ExecContext(ctx, "UPDATE tenants SET archived = TRUE WHERE id = $1", id)

	if err != nil {
		return err
	}

	return nil
}
