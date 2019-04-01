package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net"
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
		ID:            u.ID,
		Type:          u.Type,
		Email:         u.Email,
		PasswordHash:  u.PasswordHash,
		Name:          u.Name,
		Added:         u.Added,
		Modified:      u.Modified,
		EmailVerified: u.EmailVerified,
		PasswordGen:   u.PasswordGen,
		LoginAddr:     u.LoginAddr,
		RefreshAddr:   u.RefreshAddr,
		EmailGen:      u.EmailGen,
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

	if len(u.AddressWhiteList) != 0 {
		ret.AddressWhiteList = make(StringSet, len(u.AddressWhiteList))

		for _, a := range u.AddressWhiteList {
			ret.AddressWhiteList[a] = true
		}
	}

	return ret
}

// Storage service that manage database operation for the user resource
type Storage struct {
	DB *sqlx.DB
}

const getUserQuery = `
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
			  'tenant_type',
			  tenants.tenant_type,
			  'roles',
			  r.roles
			)
		  ) AS membership
		FROM
		  membership
		  INNER JOIN tenants ON tenants.id = membership.tenant_id
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
	  ) AS ips ON ips.user_id = users.id`

// GetUserByID retrieve a user by his ID
func (s *Storage) GetUserByID(ctx context.Context, typ string, id uuid.UUID) (*User, error) {
	q := getUserQuery + " WHERE users.id = $1"
	args := []interface{}{id}

	if typ != "" {
		q += " AND users.account_type = $2"
		args = append(args, typ)
	}

	var u userModel
	if err := s.DB.GetContext(ctx, &u, q, args...); err != nil {
		if err == sql.ErrNoRows {
			err = errors.ErrUserNotFound
		}

		return nil, err
	}

	return u.toUser(), nil
}

// GetUserByEmail retrieve a user by his Email
func (s *Storage) GetUserByEmail(ctx context.Context, typ, email string) (*User, error) {
	q := getUserQuery + " WHERE users.email = $1"
	args := []interface{}{email}

	if typ != "" {
		q += " AND users.account_type = $2"
		args = append(args, typ)
	}

	var u userModel
	if err := s.DB.GetContext(ctx, &u, q, args...); err != nil {
		if err == sql.ErrNoRows {
			err = errors.ErrUserNotFound
		}

		return nil, err
	}

	return u.toUser(), nil
}

// GetServiceAccountByAddress retrieve a user by whitelisted IP address if any
func (s *Storage) GetServiceAccountByAddress(ctx context.Context, address string) (*User, error) {
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
			  'tenant_type',
			  tenants.tenant_type,
			  'roles',
			  r.roles
	        )
	      ) AS membership
	    FROM
	      membership
		  INNER JOIN tenants ON tenants.id = membership.tenant_id
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
	  users.account_type = 'service' AND
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

var userQueryColumns = query.Columns{
	"id":             {Name: "users.id", Flags: query.ColSort},
	"email":          {Name: "users.email", Flags: query.ColSort},
	"name":           {Name: "users.name", Flags: query.ColSort},
	"added":          {Name: "users.added", Flags: query.ColSort},
	"modified":       {Name: "users.modified", Flags: query.ColSort},
	"email_verified": {Name: "users.email_verified", Flags: query.ColSort},
	"login_addr":     {Name: "users.login_addr", Flags: query.ColSort},
	"login_ts":       {Name: "users.login_ts", Flags: query.ColSort},
	"refresh_addr":   {Name: "users.refresh_addr", Flags: query.ColSort},
	"refresh_ts":     {Name: "users.refresh_ts", Flags: query.ColSort},
	"account_type":   {Name: "users.account_type", Flags: query.ColSort},
}

// GetUsers retrieve a users according to a query and return a paged results
func (s *Storage) GetUsers(ctx context.Context, typ string, q *query.Query) (users []*User, count int, next *query.Query, err error) {
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
				    'tenant_type',
					tenants.tenant_type,
				    'roles',
				    r.roles
				  )
				) AS membership
			  FROM
				membership
				INNER JOIN tenants ON tenants.id = membership.tenant_id
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
		IDColumn:   "users.id",
		ColumnFunc: userQueryColumns.Func,
	}

	tmp := *q
	if typ != "" {
		tmp.Match = append(tmp.Match, query.Expr{
			Col:   "account_type",
			Op:    query.OpEq,
			Value: typ,
		})
	}

	stmt, args, err := tmp.SelectStmt(&selOpt)
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
	if tmp.TotalCount {
		if stmt, args, err = tmp.CountStmt(&selOpt); err != nil {
			return
		}

		if err = s.DB.Get(&count, stmt, args...); err != nil {
			return
		}
	}

	users = usersSlice

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

func isUniqueViolation(err error, constraint string) bool {
	e, ok := err.(*pq.Error)
	return ok && e.Code.Name() == "unique_violation" && e.Constraint == constraint
}

// NewUserInt insert a new user in the database along with his initial tenant
func NewUserInt(ctx context.Context, tx *sqlx.Tx, user *CreateUser) (res *User, err error) {
	model := userModel{
		ID:            user.ID,
		Email:         user.Email,
		PasswordHash:  user.PasswordHash,
		Name:          user.Name,
		EmailVerified: user.EmailVerified,
		Type:          user.Type,
	}

	// Create user

	var idVal string
	if model.ID != uuid.Nil {
		idVal = ":id"
	} else {
		idVal = "DEFAULT"
	}

	q := "INSERT INTO users (id, account_type, email, password_hash, name, email_verified) VALUES (" + idVal + ", :account_type, :email, :password_hash, :name, :email_verified) RETURNING id, added, modified, password_gen"
	rows, err := sqlx.NamedQueryContext(ctx, tx, q, &model)
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

	tModel := TenantModel{}

	// Create tenant
	if err = tx.GetContext(ctx, &tModel, "INSERT INTO tenants (name, tenant_type) VALUES ($1, 'individual') RETURNING *", model.Email); err != nil {
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

	if _, err = tx.ExecContext(ctx, "INSERT INTO roles (membership_id, role) VALUES "+strings.Join(valuesExprs, ", "), args...); err != nil {
		return
	}

	if len(user.AddressWhiteList) != 0 {
		// Create whitelist
		valuesExprs = make([]string, len(user.AddressWhiteList))
		args = make([]interface{}, len(user.AddressWhiteList)+1)

		args[0] = model.ID
		i = 0

		for _, a := range user.AddressWhiteList {
			valuesExprs[i] = fmt.Sprintf("($1, $%d)", i+2)
			args[i+1] = a.String()
			i++
		}

		if _, err = tx.ExecContext(ctx, "INSERT INTO service_account_ip (user_id, addr) VALUES "+strings.Join(valuesExprs, ", "), args...); err != nil {
			return
		}
	}

	res = model.toUser()

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
	}()

	user.ID = uuid.NewV4()

	tmp, err := NewUserInt(ctx, tx, user)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return s.GetUserByID(ctx, "", tmp.ID)
}

func errPatchPath(p string) error {
	return errors.Wrap(fmt.Errorf("Invalid property `%s'", p), errors.CodeBadRequest)
}

var updatePaths = map[string]struct{}{
	"name":          struct{}{},
	"password_hash": struct{}{},
}

// UpdateUser update user according to patch operations
func (s *Storage) UpdateUser(ctx context.Context, typ string, id uuid.UUID, ops *Ops) (user *User, err error) {
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
	}()

	// Update properties
	var i int
	expr := "UPDATE users SET "
	args := make([]interface{}, 0, len(ops.Update)+2)

	for k, v := range ops.Update {
		if i != 0 {
			expr += ", "
		}
		expr += fmt.Sprintf("%s = $%d", pq.QuoteIdentifier(k), i+1)
		args = append(args, v)
		i++
	}

	if len(ops.Update) != 0 {
		expr += ", "
	}

	expr += "modified = DEFAULT WHERE "

	expr += fmt.Sprintf("id = $%d", i+1)
	args = append(args, id)
	i++

	if typ != "" {
		expr += fmt.Sprintf(" AND account_type = $%d", i+1)
		args = append(args, typ)
	}

	res, err := tx.ExecContext(ctx, expr, args...)
	if err != nil {
		return nil, err
	}

	v, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}

	if v == 0 {
		return nil, errors.ErrUserNotFound
	}

	// Update white list
	addAddr := ops.Add["address_whitelist"]
	removeAddr := ops.Remove["address_whitelist"]

	if len(addAddr) != 0 {
		expr := "INSERT INTO service_account_ip (user_id, addr) VALUES "
		args := make([]interface{}, len(addAddr)+1)

		args[0] = id

		for i, r := range addAddr {
			if net.ParseIP(r) == nil {
				return nil, errors.ErrAddrSyntax
			}

			if i != 0 {
				expr += ", "
			}
			expr += fmt.Sprintf("($1, $%d)", i+2)
			args[i+1] = r
		}

		if _, err = tx.ExecContext(ctx, expr, args...); err != nil {
			if isUniqueViolation(err, "service_account_ip_addr_key") {
				err = errors.ErrAddrExists
			}
			return nil, err
		}
	}

	if len(removeAddr) != 0 {
		expr := "DELETE FROM service_account_ip WHERE user_id = $1 AND ("
		args := make([]interface{}, len(removeAddr)+1)

		args[0] = id

		for i, r := range removeAddr {
			if i != 0 {
				expr += " OR "
			}
			expr += fmt.Sprintf("addr = $%d", i+2)
			args[i+1] = r
		}

		expr += ")"

		if _, err = tx.ExecContext(ctx, expr, args...); err != nil {
			return nil, err
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return s.GetUserByID(ctx, typ, id)
}

// DeleteUser delete user with the specified ID
func (s *Storage) DeleteUser(ctx context.Context, typ string, id uuid.UUID) (err error) {
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

	q := "DELETE FROM users WHERE id = $1"
	args := []interface{}{id}

	if typ != "" {
		q += " AND account_type = $2"
		args = append(args, typ)
	}

	res, err := tx.ExecContext(ctx, q, args...)
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
	if err := tx.GetContext(ctx, &gen, "SELECT password_gen FROM users WHERE id = $1 AND account_type = 'regular'", id); err != nil {
		if err == sql.ErrNoRows {
			err = errors.ErrUserNotFound
		}
		return err
	}

	if gen != expectedGen {
		log.WithFields(log.Fields{"token": expectedGen, "db": gen}).Println("Reset token expired")
		return errors.ErrTokenExpired
	}

	res, err := tx.ExecContext(ctx, "UPDATE users SET password_hash = $1, email_verified = TRUE, modified = DEFAULT, password_gen = password_gen + 1 WHERE account_type = 'regular' AND id = $2 AND password_gen = $3", hash, id, gen)
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
	if err := tx.GetContext(ctx, &prev, "SELECT email_gen, email FROM users WHERE id = $1 AND account_type = 'regular'", id); err != nil {

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
	if err = tx.GetContext(ctx, &u, "UPDATE users SET email = $1, email_verified = TRUE, modified = DEFAULT, email_gen = email_gen + 1 WHERE account_type = 'regular' AND id = $2 AND email_gen = $3 RETURNING *", email, id, prev.EmailGen); err != nil {
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

func (s *Storage) GetUserIDByMembershipID(ctx context.Context, account_type string, membershipID uuid.UUID) (uuid.UUID, error) {
	var uid uuid.UUID
	if err := s.DB.GetContext(ctx, &uid, "SELECT user_id from membership WHERE id = $1", membershipID); err != nil {
		if err == sql.ErrNoRows {
			err = errors.ErrMembershipNotFound
		}
		return uid, err
	}
	return uid, nil
}
