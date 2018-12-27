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

type tenantModel struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Added     time.Time `json:"added" db:"added"`
	Modified  time.Time `json:"modified" db:"modified"`
	Protected bool      `json:"-" db:"protected"`
	SortedBy  string    `json:"-" db:"sorted_by"`
}

func (t *tenantModel) Clone() *tenantModel {
	return &tenantModel{
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

func (s *TenantStorage) CreateTenant(ctx context.Context, name string) (*tenantModel, error) {
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

	newTenant := tenantModel{
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

	return &newTenant, nil
}

func (s *TenantStorage) GetTenant(ctx context.Context, id uuid.UUID, self *User, onlySelf bool) (*tenantModel, error) {
	var queryExtension = ""
	if onlySelf {
		queryExtension += "AND id IN (SELECT tenant_id FROM membership WHERE user_id = '" + self.ID.String() + "')"
	}
	model := tenantModel{}
	err := s.DB.GetContext(ctx, &model, "SELECT tenants.* FROM tenants WHERE id = $1"+queryExtension, id)

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
}

func (s *TenantStorage) GetTenants(ctx context.Context, self *User, onlySelf bool, q *query.Query) (tenants []*tenantModel, count int, next *query.Query, err error) {
	var queryExtension = ""
	if onlySelf {
		queryExtension += "WHERE id IN (SELECT tenant_id FROM membership WHERE user_id = '" + self.ID.String() + "')"
	}

	if q.SortBy == "" {
		q.SortBy = TenantsDefaultSortColumn
	}

	selOpt := query.SelectOptions{
		SelectExpr: "tenants.*, tenants." + pq.QuoteIdentifier(q.SortBy) + " AS sorted_by",
		FromExpr:   "tenants " + queryExtension,
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

	tenantsSlice := []*tenantModel{}
	var lastItem *tenantModel

	for rows.Next() {
		var tenant tenantModel
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

func (s *TenantStorage) PatchTenant(ctx context.Context, id uuid.UUID, ops *TenantOps) (*tenantModel, error) {
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

	var tenant tenantModel
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

	_, err = tx.ExecContext(ctx, "DELETE tenants WHERE id = $1", id)

	if err != nil {
		return err
	}

	return nil
}