package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/ecadlabs/auth/errors"
	"github.com/ecadlabs/auth/query"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
)

type userModel struct {
	ID               uuid.UUID      `db:"id"`
	Type             string         `db:"account_type"`
	Email            string         `db:"email"`
	EmailGen         int            `db:"email_gen"`
	PasswordHash     []byte         `db:"password_hash"`
	PasswordGen      int            `db:"password_gen"`
	Name             string         `db:"name"`
	Added            time.Time      `db:"added"`
	Modified         time.Time      `db:"modified"`
	EmailVerified    bool           `db:"email_verified"`
	SortedBy         string         `db:"sorted_by"` // Output only
	Membership       []byte         `db:"membership"`
	LoginAddr        string         `db:"login_addr"`
	LoginTimestamp   time.Time      `db:"login_ts"`
	RefreshAddr      string         `db:"refresh_addr"`
	RefreshTimestamp time.Time      `db:"refresh_ts"`
	AddressWhiteList pq.StringArray `db:"ip_whitelist"`
}

func (u *userModel) toUser() *User {
	ret := &User{
		ID:               u.ID,
		Type:             u.Type,
		Email:            u.Email,
		PasswordHash:     u.PasswordHash,
		Name:             u.Name,
		Added:            u.Added,
		Modified:         u.Modified,
		EmailVerified:    u.EmailVerified,
		PasswordGen:      u.PasswordGen,
		LoginAddr:        u.LoginAddr,
		RefreshAddr:      u.RefreshAddr,
		EmailGen:         u.EmailGen,
		AddressWhiteList: u.AddressWhiteList,
	}

	epoch := time.Unix(0, 0).UTC()

	if u.LoginTimestamp.UTC() != epoch {
		ret.LoginTimestamp = &u.LoginTimestamp
	}

	if u.RefreshTimestamp.UTC() != epoch {
		ret.RefreshTimestamp = &u.RefreshTimestamp
	}

	if len(u.Membership) != 0 {
		if err := json.Unmarshal(u.Membership, &ret.Membership); err != nil {
			log.Warn(err)
		}
	}

	return ret
}

// Storage service that manage database operation for the user resource
type Storage struct {
	DB *sqlx.DB
}

func (s *Storage) getUser(ctx context.Context, col string, val interface{}) (*User, error) {
	f := `
	SELECT
	  users.*,
	  m.membership,
	  ips.ip_whitelist
	FROM
	  users
	  LEFT JOIN (
	    SELECT
	      membership.user_id,
	      json_agg(
	        json_build_object(
	          'tenant_id',
	          membership.tenant_id,
	          'type',
	          membership.membership_type,
	          'roles',
	          r.roles
	        )
	      ) AS membership
	    FROM
	      membership
	      LEFT JOIN (
	        SELECT
	          membership_id,
	          json_agg(role) AS roles
	        FROM
	          roles
	        GROUP BY
	          membership_id
	      ) AS r ON membership.id = r.membership_id
	    WHERE
	      membership.membership_status = 'active'
	    GROUP BY
	      membership.user_id
	  ) AS m ON m.user_id = users.id
	  LEFT JOIN (
	    SELECT
	      user_id,
	      array_agg(addr) AS ip_whitelist
	    FROM
	      service_account_ip
	    GROUP BY
	      user_id
	  ) AS ips ON ips.user_id = users.id
	WHERE
	  users.%s = $1`

	var u userModel
	if err := s.DB.GetContext(ctx, &u, fmt.Sprintf(f, pq.QuoteIdentifier(col)), val); err != nil {
		if err == sql.ErrNoRows {
			err = errors.ErrUserNotFound
		}

		return nil, err
	}

	return u.toUser(), nil
}

// GetUserByID retrieve a user by his ID
func (s *Storage) GetUserByID(ctx context.Context, id uuid.UUID) (*User, error) {
	return s.getUser(ctx, "id", id)
}

// GetUserByEmail retrieve a user by his Email
func (s *Storage) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	return s.getUser(ctx, "email", email)
}

// GetUserByIPAddress retrieve a user by whitelisted IP address if any
func (s *Storage) GetUserByIPAddress(ctx context.Context, address string) (*User, error) {
	q := `
	SELECT
	  users.*,
	  m.membership
	FROM
	  users
	  INNER JOIN service_account_ip ON service_account_ip.user_id = users.id
	  LEFT JOIN (
	    SELECT
	      membership.user_id,
	      json_agg(
	        json_build_object(
	          'tenant_id',
	          membership.tenant_id,
	          'type',
	          membership.membership_type,
	          'roles',
	          r.roles
	        )
	      ) AS membership
	    FROM
	      membership
	      LEFT JOIN (
	        SELECT
	          membership_id,
	          json_agg(role) AS roles
	        FROM
	          roles
	        GROUP BY
	          membership_id
	      ) AS r ON membership.id = r.membership_id
	    WHERE
	      membership.membership_status = 'active'
	    GROUP BY
	      membership.user_id
	  ) AS m ON m.user_id = users.id
	WHERE
	  service_account_ip.addr = $1`

	var u userModel
	if err := s.DB.GetContext(ctx, &u, q, address); err != nil {
		if err == sql.ErrNoRows {
			err = errors.ErrUserNotFound
		}

		return nil, err
	}

	return u.toUser(), nil
}

var userQueryColumns = map[string]int{
	"id":             query.ColQuery | query.ColSort,
	"email":          query.ColQuery | query.ColSort,
	"name":           query.ColQuery | query.ColSort,
	"added":          query.ColQuery | query.ColSort,
	"modified":       query.ColQuery | query.ColSort,
	"email_verified": query.ColQuery | query.ColSort,
	"login_addr":     query.ColQuery | query.ColSort,
	"login_ts":       query.ColQuery | query.ColSort,
	"refresh_addr":   query.ColQuery | query.ColSort,
	"refresh_ts":     query.ColQuery | query.ColSort,
}

// GetUsers retrieve a users according to a query and return a paged results
func (s *Storage) GetUsers(ctx context.Context, q *query.Query) (users []*User, count int, next *query.Query, err error) {
	if q.SortBy == "" {
		q.SortBy = UsersDefaultSortColumn
	}

	selOpt := query.SelectOptions{
		SelectExpr: "users.*, m.membership, ips.ip_whitelist, users." + pq.QuoteIdentifier(q.SortBy) + " AS sorted_by",
		FromExpr: `
			users
			LEFT JOIN (
			  SELECT
				membership.user_id,
				json_agg(
				  json_build_object(
					'tenant_id',
					membership.tenant_id,
					'type',
					membership.membership_type,
					'roles',
					r.roles
				  )
				) AS membership
			  FROM
				membership
				LEFT JOIN (
				  SELECT
					membership_id,
					json_agg(role) AS roles
				  FROM
					roles
				  GROUP BY
					membership_id
				) AS r ON membership.id = r.membership_id
			  WHERE
				membership.membership_status = 'active'
			  GROUP BY
				membership.user_id
			) AS m ON m.user_id = users.id
			LEFT JOIN (
			  SELECT
				user_id,
				array_agg(addr) AS ip_whitelist
			  FROM
				service_account_ip
			  GROUP BY
				user_id
			) AS ips ON ips.user_id = users.id`,
		IDColumn: "id",
		ColumnFlagsFunc: func(col string) int {
			if flags, ok := userQueryColumns[col]; ok {
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

	usersSlice := []*User{}
	var lastItem *userModel

	for rows.Next() {
		var user userModel
		if err = rows.StructScan(&user); err != nil {
			return
		}

		lastItem = &user
		usersSlice = append(usersSlice, user.toUser())
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

	users = usersSlice

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

func isUniqueViolation(err error, constraint string) bool {
	e, ok := err.(*pq.Error)
	return ok && e.Code.Name() == "unique_violation" && e.Constraint == constraint
}

// NewUserInt insert a new user in the database along with his initial tenant
func NewUserInt(ctx context.Context, tx *sqlx.Tx, user *CreateUser) (res *User, err error) {
	model := userModel{
		ID:            uuid.NewV4(),
		Email:         user.Email,
		PasswordHash:  user.PasswordHash,
		Name:          user.Name,
		EmailVerified: user.EmailVerified,
	}

	// Create user
	rows, err := sqlx.NamedQueryContext(ctx, tx, "INSERT INTO users (id, email, password_hash, name, email_verified) VALUES (:id, :email, :password_hash, :name, :email_verified) RETURNING added, modified, password_gen", &model)
	if err != nil {
		if isUniqueViolation(err, "users_email_key") {
			err = errors.ErrEmailInUse
		}
		return
	}
	defer rows.Close()

	for rows.Next() {
		if err = rows.StructScan(&model); err != nil {
			return
		}
	}

	if err = rows.Err(); err != nil {
		return
	}

	res = model.toUser()

	tModel := TenantModel{}

	// Create tenant
	rows, err = sqlx.NamedQueryContext(ctx, tx, "INSERT INTO tenants (name, tenant_type) VALUES (:email, 'individual') RETURNING *", &model)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		if err = rows.StructScan(&tModel); err != nil {
			return
		}
	}

	if err = rows.Err(); err != nil {
		return
	}

	var mid uuid.UUID
	if err = tx.GetContext(ctx, &mid, "INSERT INTO membership (user_id, tenant_id, membership_type) VALUES ($1, $2, $3) RETURNING id", model.ID, tModel.ID, OwnerMembership); err != nil {
		return
	}

	user.Roles["owner"] = true

	// Create roles
	valuesExprs := make([]string, len(user.Roles))
	args := make([]interface{}, len(user.Roles)+1)

	args[0] = mid
	var i int

	for r := range user.Roles {
		valuesExprs[i] = fmt.Sprintf("($1, $%d)", i+2)
		args[i+1] = r
		i++
	}

	_, err = tx.ExecContext(ctx, "INSERT INTO roles (membership_id, role) VALUES "+strings.Join(valuesExprs, ", "), args...)
	return
}

// NewUser insert a new user in the database along with his initial tenant
// Wrap the database queries with a transaction
func (s *Storage) NewUser(ctx context.Context, user *CreateUser) (res *User, err error) {
	tx, err := s.DB.Beginx()
	if err != nil {
		return
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}

		err = tx.Commit()
	}()

	res, err = NewUserInt(ctx, tx, user)
	return
}

func errPatchPath(p string) error {
	return errors.Wrap(fmt.Errorf("Invalid property `%s'", p), errors.CodeBadRequest)
}

var updatePaths = map[string]struct{}{
	"name":          struct{}{},
	"password_hash": struct{}{},
}

// UpdateUser update user according to patch operations
func (s *Storage) UpdateUser(ctx context.Context, id uuid.UUID, ops *Ops) (user *User, err error) {
	// Verify columns
	for k := range ops.Update {
		if _, ok := updatePaths[k]; !ok {
			return nil, errPatchPath(k)
		}
	}

	tx, err := s.DB.Beginx()
	if err != nil {
		return
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
	expr := "UPDATE users SET "
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

	var u userModel
	if err = tx.GetContext(ctx, &u, expr, args...); err != nil {
		if err == sql.ErrNoRows {
			err = errors.ErrUserNotFound
		}
		return nil, err
	}

	return u.toUser(), nil
}

// DeleteUser delete user with the specified ID
func (s *Storage) DeleteUser(ctx context.Context, id uuid.UUID) (err error) {
	tx, err := s.DB.Beginx()
	if err != nil {
		return
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}

		err = tx.Commit()
	}()

	res, err := tx.ExecContext(ctx, "DELETE FROM users WHERE id = $1", id)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return errors.ErrUserNotFound
	}

	return nil
}

// UpdatePasswordWithGen set a new password according to the hash parameter
func (s *Storage) UpdatePasswordWithGen(ctx context.Context, id uuid.UUID, hash []byte, expectedGen int) (err error) {
	tx, err := s.DB.Beginx()
	if err != nil {
		return
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}

		err = tx.Commit()
	}()

	var gen int
	if err := tx.GetContext(ctx, &gen, "SELECT password_gen FROM users WHERE id = $1", id); err != nil {
		if err == sql.ErrNoRows {
			err = errors.ErrUserNotFound
		}
		return err
	}

	if gen != expectedGen {
		log.WithFields(log.Fields{"token": expectedGen, "db": gen}).Println("Reset token expired")
		return errors.ErrTokenExpired
	}

	res, err := tx.ExecContext(ctx, "UPDATE users SET password_hash = $1, email_verified = TRUE, modified = DEFAULT, password_gen = password_gen + 1 WHERE id = $2 AND password_gen = $3", hash, id, gen)
	if err != nil {
		return err
	}

	v, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if v == 0 {
		return errors.ErrUserNotFound // Unlikely
	}

	return nil
}

// UpdateEmailWithGen update the email
func (s *Storage) UpdateEmailWithGen(ctx context.Context, id uuid.UUID, email string, expectedGen int) (user *User, oldEmail string, err error) {
	tx, err := s.DB.Beginx()
	if err != nil {
		return
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}

		err = tx.Commit()
	}()

	var prev userModel
	if err := tx.GetContext(ctx, &prev, "SELECT email_gen, email FROM users WHERE id = $1", id); err != nil {
		if err == sql.ErrNoRows {
			err = errors.ErrUserNotFound
		}
		return nil, "", err
	}

	if prev.EmailGen != expectedGen {
		log.WithFields(log.Fields{"token": expectedGen, "db": prev.EmailGen}).Println("Email update token expired")
		return nil, "", errors.ErrTokenExpired
	}

	var u userModel
	if err = tx.GetContext(ctx, &u, "UPDATE users SET email = $1, email_verified = TRUE, modified = DEFAULT, email_gen = email_gen + 1 WHERE id = $2 AND email_gen = $3 RETURNING *", email, id, prev.EmailGen); err != nil {
		if err == sql.ErrNoRows {
			err = errors.ErrUserNotFound
		}
		return nil, "", err
	}

	return u.toUser(), prev.Email, nil
}

// UpdateLoginInfo update the address and time of last login
func (s *Storage) UpdateLoginInfo(ctx context.Context, id uuid.UUID, addr string) error {
	_, err := s.DB.ExecContext(ctx, "UPDATE users SET login_addr = $1, login_ts = NOW() WHERE id = $2", addr, id)
	return err
}

// UpdateRefreshInfo update the address and time of last refresh
func (s *Storage) UpdateRefreshInfo(ctx context.Context, id uuid.UUID, addr string) error {
	_, err := s.DB.ExecContext(ctx, "UPDATE users SET refresh_addr = $1, refresh_ts = NOW() WHERE id = $2", addr, id)
	return err
}

// Ping ping the database
func (s *Storage) Ping(ctx context.Context) error {
	if err := s.DB.PingContext(ctx); err != nil {
		return err
	}

	return nil
}
