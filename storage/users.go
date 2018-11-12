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
	"github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
)

type userModel struct {
	ID               uuid.UUID      `db:"id"`
	Email            string         `db:"email"`
	EmailGen         int            `db:"email_gen"`
	PasswordHash     []byte         `db:"password_hash"`
	PasswordGen      int            `db:"password_gen"`
	Name             string         `db:"name"`
	Added            time.Time      `db:"added"`
	Modified         time.Time      `db:"modified"`
	EmailVerified    bool           `db:"email_verified"`
	SortedBy         string         `db:"sorted_by"` // Output only
	Roles            pq.StringArray `db:"roles"`
	LoginAddr        string         `db:"login_addr"`
	LoginTimestamp   time.Time      `db:"login_ts"`
	RefreshAddr      string         `db:"refresh_addr"`
	RefreshTimestamp time.Time      `db:"refresh_ts"`
}

func (u *userModel) toUser() *User {
	ret := &User{
		ID:            u.ID,
		Email:         u.Email,
		PasswordHash:  u.PasswordHash,
		Name:          u.Name,
		Added:         u.Added,
		Modified:      u.Modified,
		Roles:         make(Roles, len(u.Roles)),
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

	for _, r := range u.Roles {
		ret.Roles[r] = true
	}

	return ret
}

type Storage struct {
	DB *sqlx.DB
}

func (s *Storage) getUser(ctx context.Context, col string, val interface{}) (*User, error) {
	var u userModel

	q := "SELECT users.*, ra.roles FROM users LEFT JOIN (SELECT user_id, array_agg(role) AS roles FROM roles GROUP BY user_id) AS ra ON ra.user_id = users.id WHERE users." + pq.QuoteIdentifier(col) + " = $1"
	if err := s.DB.GetContext(ctx, &u, q, val); err != nil {
		if err == sql.ErrNoRows {
			err = errors.ErrUserNotFound
		}

		return nil, err
	}

	return u.toUser(), nil
}

func (s *Storage) GetUserByID(ctx context.Context, id uuid.UUID) (*User, error) {
	return s.getUser(ctx, "id", id)
}

func (s *Storage) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	return s.getUser(ctx, "email", email)
}

var userQueryColumns = map[string]int{
	"id":             query.ColQuery | query.ColSort,
	"email":          query.ColQuery | query.ColSort,
	"name":           query.ColQuery | query.ColSort,
	"added":          query.ColQuery | query.ColSort,
	"modified":       query.ColQuery | query.ColSort,
	"roles":          query.ColQuery,
	"email_verified": query.ColQuery | query.ColSort,
	"login_addr":     query.ColQuery | query.ColSort,
	"login_ts":       query.ColQuery | query.ColSort,
	"refresh_addr":   query.ColQuery | query.ColSort,
	"refresh_ts":     query.ColQuery | query.ColSort,
}

func (s *Storage) GetUsers(ctx context.Context, q *query.Query) (users []*User, count int, next *query.Query, err error) {
	if q.SortBy == "" {
		q.SortBy = UsersDefaultSortColumn
	}

	selOpt := query.SelectOptions{
		SelectExpr: "users.*, ra.roles, users." + pq.QuoteIdentifier(q.SortBy) + " AS sorted_by",
		FromExpr:   "users LEFT JOIN (SELECT user_id, array_agg(role) AS roles FROM roles GROUP BY user_id) AS ra ON ra.user_id = users.id",
		IDColumn:   "id",
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

func NewUserInt(ctx context.Context, tx *sqlx.Tx, user *User) (res *User, err error) {
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

	if len(user.Roles) == 0 {
		return
	}

	res.Roles = user.Roles

	// Create roles
	valuesExprs := make([]string, len(user.Roles))
	args := make([]interface{}, len(user.Roles)+1)

	args[0] = model.ID
	var i int

	for r := range user.Roles {
		valuesExprs[i] = fmt.Sprintf("($1, $%d)", i+2)
		args[i+1] = r
		i++
	}

	_, err = tx.ExecContext(ctx, "INSERT INTO roles (user_id, role) VALUES "+strings.Join(valuesExprs, ", "), args...)
	return
}

func (s *Storage) NewUser(ctx context.Context, user *User) (res *User, err error) {
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

func (s *Storage) UpdateUser(ctx context.Context, id uuid.UUID, ops *UserOps) (user *User, err error) {
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

	// Update roles
	if len(ops.AddRoles) != 0 {
		expr := "INSERT INTO roles (user_id, role) VALUES "
		args := make([]interface{}, len(ops.AddRoles)+1)

		for i, r := range ops.AddRoles {
			if i != 0 {
				expr += ", "
			}
			expr += fmt.Sprintf("($1, $%d)", i+2)
			args[i+1] = r
		}

		args[0] = id

		if _, err = tx.ExecContext(ctx, expr, args...); err != nil {
			if isUniqueViolation(err, "roles_pkey") {
				err = errors.ErrRoleExists
			}
			return nil, err
		}
	}

	if len(ops.RemoveRoles) != 0 {
		expr := "DELETE FROM roles WHERE user_id = $1 AND ("
		args := make([]interface{}, len(ops.RemoveRoles)+1)

		for i, r := range ops.RemoveRoles {
			if i != 0 {
				expr += " OR "
			}
			expr += fmt.Sprintf("role = $%d", i+2)
			args[i+1] = r
		}

		expr += ")"
		args[0] = id

		if _, err = tx.ExecContext(ctx, expr, args...); err != nil {
			return nil, err
		}
	}

	// Get roles back
	if err = tx.GetContext(ctx, &u, "SELECT array_agg(role) AS roles FROM roles WHERE user_id = $1 GROUP BY user_id", id); err != nil {
		if err != sql.ErrNoRows {
			return nil, err
		}
	}

	return u.toUser(), nil
}

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

	_, err = tx.ExecContext(ctx, "DELETE FROM roles WHERE user_id = $1", id)
	if err != nil {
		return err
	}

	return nil
}

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

	// Get roles back
	if err = tx.GetContext(ctx, &u, "SELECT array_agg(role) AS roles FROM roles WHERE user_id = $1 GROUP BY user_id", id); err != nil {
		if err != sql.ErrNoRows {
			return nil, "", err
		}
	}

	return u.toUser(), prev.Email, nil
}

func (s *Storage) UpdateLoginInfo(ctx context.Context, id uuid.UUID, addr string) error {
	_, err := s.DB.ExecContext(ctx, "UPDATE users SET login_addr = $1, login_ts = NOW() WHERE id = $2", addr, id)
	return err
}

func (s *Storage) UpdateRefreshInfo(ctx context.Context, id uuid.UUID, addr string) error {
	_, err := s.DB.ExecContext(ctx, "UPDATE users SET refresh_addr = $1, refresh_ts = NOW() WHERE id = $2", addr, id)
	return err
}

func (s *Storage) Ping(ctx context.Context) error {
	if err := s.DB.PingContext(ctx); err != nil {
		return err
	}

	return nil
}
