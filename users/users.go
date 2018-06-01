package users

import (
	"context"
	"database/sql"
	"git.ecadlabs.com/ecad/auth/jsonpatch"
	"git.ecadlabs.com/ecad/auth/query"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

type userModel struct {
	ID            uuid.UUID      `db:"id"`
	Email         string         `db:"email"`
	PasswordHash  []byte         `db:"password_hash"`
	Name          sql.NullString `db:"name"`
	Added         time.Time      `db:"added"`
	Modified      time.Time      `db:"modified"`
	EmailVerified bool           `db:"email_verified"`
	SortedBy      string         `db:"sorted_by"` // Output only
	Roles         pq.StringArray `db:"roles"`
}

func (u *userModel) toUser() *User {
	return &User{
		ID:            u.ID,
		Email:         u.Email,
		PasswordHash:  u.PasswordHash,
		Name:          u.Name.String,
		Added:         u.Added,
		Modified:      u.Modified,
		Roles:         u.Roles,
		EmailVerified: u.EmailVerified,
	}
}

type Storage struct {
	DB *sqlx.DB
}

func (s *Storage) getUser(ctx context.Context, col string, val interface{}) (*User, error) {
	var u userModel

	q := "SELECT users.*, ra.roles FROM users LEFT JOIN (SELECT user_id, array_agg(role) AS roles FROM roles GROUP BY user_id) AS ra ON ra.user_id = users.id WHERE users." + pq.QuoteIdentifier(col) + " = $1"
	if err := s.DB.GetContext(ctx, &u, q, val); err != nil {
		if err == sql.ErrNoRows {
			err = ErrNotFound
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

var queryColumns = map[string]struct{}{
	"id":             struct{}{},
	"email":          struct{}{},
	"name":           struct{}{},
	"added":          struct{}{},
	"modified":       struct{}{},
	"roles":          struct{}{},
	"email_verified": struct{}{},
}

func (s *Storage) GetUsers(ctx context.Context, q *query.Query) ([]*User, *query.Query, error) {
	if q.SortBy == "" {
		q.SortBy = DefaultSortColumn
	}

	selOpt := query.SelectOptions{
		SelectExpr: "users.*, ra.roles, users." + pq.QuoteIdentifier(q.SortBy) + " AS sorted_by",
		FromExpr:   "users LEFT JOIN (SELECT user_id, array_agg(role) AS roles FROM roles GROUP BY user_id) AS ra ON ra.user_id = users.id",
		IDColumn:   "id",
		ValidateColumn: func(col string) bool {
			_, ok := queryColumns[col]
			return ok
		},
	}

	stmt, args, err := q.SelectStmt(&selOpt)
	if err != nil {
		return nil, nil, &Error{err, http.StatusBadRequest}
	}
	log.Println(stmt)

	rows, err := s.DB.QueryxContext(ctx, stmt, args...)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	users := []*User{}
	var lastItem *userModel

	for rows.Next() {
		var user userModel
		if err := rows.StructScan(&user); err != nil {
			return nil, nil, err
		}

		lastItem = &user
		users = append(users, user.toUser())
	}

	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	ret := *q

	if lastItem != nil {
		// Update query
		ret.LastID = lastItem.ID.String()
		ret.Last = lastItem.SortedBy
	}

	return users, &ret, nil
}

func isUniqueViolation(err error, constraint string) bool {
	e, ok := err.(*pq.Error)
	return ok && e.Code.Name() == "unique_violation" && e.Constraint == constraint
}

func (s *Storage) NewUser(ctx context.Context, user *User) (*User, error) {
	model := userModel{
		ID:           uuid.NewV4(),
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		Name:         sql.NullString{String: user.Name, Valid: user.Name != ""},
		//Role:         user.Role, TODO
	}

	rows, err := s.DB.NamedQueryContext(ctx, "INSERT INTO users (id, email, password_hash, name) VALUES (:id, :email, :password_hash, :name) RETURNING added, modified, email_verified", &model)
	if err != nil {
		if isUniqueViolation(err, "users_email_key") {
			err = ErrEmail
		}

		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.StructScan(&model); err != nil {
			return nil, err
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return model.toUser(), nil
}

var patchColumns = map[string]struct{}{
	"email": struct{}{},
	"name":  struct{}{},
	//"role":  struct{}{}, //TODO
}

func (s *Storage) PatchUser(ctx context.Context, id uuid.UUID, patch jsonpatch.Patch) (*User, error) {
	updateOpt := jsonpatch.UpdateOptions{
		Table:         "users",
		IDColumn:      "id",
		ID:            id,
		ReturnUpdated: true,
		ValidateColumn: func(col string) bool {
			_, ok := patchColumns[col]
			return ok
		},
		SetDefaultColumns: []string{"modified"},
	}

	stmt, args, err := patch.UpdateStmt(&updateOpt)
	if err != nil {
		return nil, &Error{err, http.StatusBadRequest}
	}

	log.Println(stmt)

	var u userModel
	if err := s.DB.GetContext(ctx, &u, stmt, args...); err != nil {
		if err == sql.ErrNoRows {
			err = ErrNotFound
		}
		return nil, err
	}

	return u.toUser(), nil
}

func (s *Storage) DeleteUser(ctx context.Context, id uuid.UUID) error {
	res, err := s.DB.ExecContext(ctx, "DELETE FROM users WHERE id = $1", id)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrNotFound
	}

	return nil
}

func (s *Storage) Ping(ctx context.Context) error {
	if err := s.DB.PingContext(ctx); err != nil {
		return err
	}

	return nil
}
