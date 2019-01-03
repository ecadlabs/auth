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

type TenantModel struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Added       time.Time `json:"added" db:"added"`
	Modified    time.Time `json:"modified" db:"modified"`
	Protected   bool      `json:"-" db:"protected"`
	Archived    bool      `json:"-" db:"archived"`
	Tenant_type string    `json:"type" db:"tenant_type"`
	SortedBy    string    `json:"-" db:"sorted_by"`
}

func (t *TenantModel) Clone() *TenantModel {
	return &TenantModel{
		ID:        t.ID,
		Name:      t.Name,
		Added:     t.Added,
		Modified:  t.Modified,
		Protected: t.Protected,
	}
}

type TenantStorage struct {
	DB *sqlx.DB
}

func (s *TenantStorage) CreateTenant(ctx context.Context, name string, ownerId uuid.UUID) (*TenantModel, error) {
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

	newTenant := TenantModel{
		Name: name,
	}

	rows, err := sqlx.NamedQueryContext(ctx, tx, "INSERT INTO tenants (name) VALUES (:name) RETURNING id, added, modified, protected", &newTenant)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		if err = rows.StructScan(&newTenant); err != nil {
			return nil, err
		}
	}

	_, err = tx.ExecContext(ctx, "INSERT INTO membership (tenant_id, user_id, membership_status, membership_type) VALUES ($1, $2, $3, $4)", newTenant.ID, ownerId, ActiveState, OwnerMembership)

	if err != nil {
		return nil, err
	}

	_, err = tx.ExecContext(ctx, "INSERT INTO roles (tenant_id, user_id, role) VALUES ($1, $2, $3)", newTenant.ID, ownerId, OwnerRole)

	if err != nil {
		return nil, err
	}

	return &newTenant, nil
}

func (s *TenantStorage) GetTenant(ctx context.Context, tenant_id uuid.UUID, user_id uuid.UUID, onlySelf bool) (*TenantModel, error) {
	var queryExtension = ""
	if onlySelf {
		queryExtension += "AND id IN (SELECT tenant_id FROM membership WHERE user_id = '" + user_id.String() + "')"
	}
	model := TenantModel{}
	err := s.DB.GetContext(ctx, &model, "SELECT tenants.* FROM tenants WHERE id = $1"+queryExtension, tenant_id)

	if err != nil {
		if err == sql.ErrNoRows {
			err = errors.ErrTenantNotFound
		}
		return nil, err
	}

	return &model, nil
}

var tenantsQueryColumns = map[string]int{
	"id":       query.ColQuery | query.ColSort,
	"name":     query.ColQuery | query.ColSort,
	"added":    query.ColQuery | query.ColSort,
	"modified": query.ColQuery | query.ColSort,
	"archived": query.ColQuery | query.ColSort,
}

func (s *TenantStorage) GetTenantsSoleMember(ctx context.Context, user_id uuid.UUID) (tenants []*TenantModel, err error) {
	rows, err := s.DB.QueryContext(ctx, `
	SELECT * FROM tenants WHERE id IN (
		SELECT membership.tenant_id FROM membership 
		WHERE membership.tenant_id IN (
			SELECT tenant_id FROM membership WHERE user_id = $1
		)
	GROUP BY membership.tenant_id
	HAVING COUNT(user_id) = 1
	)`, user_id)

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
			&tenant.Tenant_type,
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

func (s *TenantStorage) GetTenants(ctx context.Context, user_id uuid.UUID, onlySelf bool, q *query.Query) (tenants []*TenantModel, count int, next *query.Query, err error) {
	var queryExtension = "tenants as scoped_tenants"
	if onlySelf {
		queryExtension = "(SELECT * FROM tenants WHERE id IN (SELECT tenant_id FROM membership WHERE user_id = '" + user_id.String() + "')) as scoped_tenants"
	}

	if q.SortBy == "" {
		q.SortBy = TenantsDefaultSortColumn
	}

	selOpt := query.SelectOptions{
		SelectExpr: "scoped_tenants.*, scoped_tenants." + pq.QuoteIdentifier(q.SortBy) + " AS sorted_by",
		FromExpr:   queryExtension,
		IDColumn:   "id",
		ColumnFlagsFunc: func(col string) int {
			if flags, ok := tenantsQueryColumns[col]; ok {
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
		stmt, args := q.CountStmt(&selOpt)
		if err = s.DB.Get(&count, stmt, args...); err != nil {
			return
		}
	}

	tenants = tenantsSlice

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

var tenantUpdatePaths = map[string]struct{}{
	"name": struct{}{},
}

func (s *TenantStorage) PatchTenant(ctx context.Context, id uuid.UUID, ops *Ops) (*TenantModel, error) {
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

func (s *TenantStorage) DeleteTenant(ctx context.Context, id uuid.UUID) error {
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
