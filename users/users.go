package users

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/satori/go.uuid"
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

type GetOptions struct {
	SortBy string
	Order  SortOrder
	Start  interface{}
	Limit  int
}

func (s *Storage) GetUsers(ctx context.Context, opt *GetOptions) ([]*User, error) {
	var col string
	if opt.SortBy == "" {
		col = DefaultSortColumn
	} else {
		col = opt.SortBy
	}

	col = pq.QuoteIdentifier(col)

	var so string
	if opt.Order == SortDesc {
		so = "DESC"
	} else {
		so = "ASC"
	}

	q := "SELECT * FROM users"

	if opt.Start != nil {
		q += fmt.Sprintf(" WHERE %s > :start", col)
	}

	q += fmt.Sprintf(" ORDER BY %s %s", col, so)

	if opt.Limit > 0 {
		q += " LIMIT :limit"
	}

	rows, err := s.DB.NamedQueryContext(ctx, q, map[string]interface{}{
		"start": opt.Start,
		"limit": opt.Limit,
	})

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := []*User{}

	for rows.Next() {
		var user userModel
		if err := rows.StructScan(&user); err != nil {
			return nil, err
		}

		users = append(users, user.toUser())
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func isUniqueViolation(err error, constraint string) bool {
	if e, ok := err.(*pq.Error); ok && e.Code.Name() == "unique_violation" && e.Constraint == constraint {
		return true
	}
	return false
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

func (s *Storage) Ping(ctx context.Context) error {
	if err := s.DB.PingContext(ctx); err != nil {
		return err
	}

	return nil
}
