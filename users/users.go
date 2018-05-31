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
	Role          Role           `db:"role"`
	EmailVerified bool           `db:"email_verified"`
	SortedBy      string         `db:"sorted_by"` // Output only
}

func (u *userModel) toUser() *User {
	return &User{
		ID:            u.ID,
		Email:         u.Email,
		PasswordHash:  u.PasswordHash,
		Name:          u.Name.String,
		Added:         u.Added,
		Modified:      u.Modified,
		Role:          u.Role,
		EmailVerified: u.EmailVerified,
	}
}

type Storage struct {
	DB *sqlx.DB
}

func (s *Storage) GetUserByID(ctx context.Context, id uuid.UUID) (*User, error) {
	var u userModel
	if err := s.DB.GetContext(ctx, &u, "SELECT * FROM users WHERE id = $1", id); err != nil {
		if err == sql.ErrNoRows {
			err = ErrNotFound
		}

		return nil, err
	}

	return u.toUser(), nil
}

func (s *Storage) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	var u userModel
	if err := s.DB.GetContext(ctx, &u, "SELECT * FROM users WHERE email = $1", email); err != nil {
		if err == sql.ErrNoRows {
			err = ErrNotFound
		}

		return nil, err
	}

	return u.toUser(), nil
}

var queryColumns = map[string]struct{}{
	"id":             struct{}{},
	"email":          struct{}{},
	"name":           struct{}{},
	"added":          struct{}{},
	"modified":       struct{}{},
	"role":           struct{}{},
	"email_verified": struct{}{},
}

func (s *Storage) GetUsers(ctx context.Context, q *query.Query) ([]*User, *query.Query, error) {
	if q.SortBy == "" {
		q.SortBy = DefaultSortColumn
	}

	selOpt := query.SelectOptions{
		Table:        "users",
		IDColumn:     "id",
		ReturnColumn: "sorted_by",
		ValidateColumn: func(col string) bool {
			_, ok := queryColumns[col]
			return ok
		},
	}

	stmt, args, err := q.SelectStmt(&selOpt)
	if err != nil {
		return nil, nil, &Error{err, http.StatusBadRequest}
	}

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
		Role:         user.Role,
	}

	rows, err := s.DB.NamedQueryContext(ctx, "INSERT INTO users (id, email, password_hash, name, role) VALUES (:id, :email, :password_hash, :name, :role) RETURNING added, modified, email_verified", &model)
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
	"role":  struct{}{},
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

func (s *Storage) Ping(ctx context.Context) error {
	if err := s.DB.PingContext(ctx); err != nil {
		return err
	}

	return nil
}
