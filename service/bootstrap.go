package service

import (
	"context"
	"database/sql"
	"errors"
	"git.ecadlabs.com/ecad/auth/handlers"
	"git.ecadlabs.com/ecad/auth/users"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
	"os"
)

const (
	envAdminEmail    = "AUTH_ADMIN_EMAIL"
	envAdminPassword = "AUTH_ADMIN_PASSWORD"
)

const (
	defaultAdminEmail    = "admin@admin"
	defaultAdminPassword = "admin"
)

var ErrNoBootstrap = errors.New("No bootstrapping")

func (s *Service) Bootstrap() (err error) {
	db := sqlx.NewDb(s.DB, "postgres")

	tx, err := db.Beginx()
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

	var val bool
	if err = tx.Get(&val, "SELECT val FROM bootstrap WHERE NOT val FOR UPDATE"); err != nil {
		if err == sql.ErrNoRows {
			err = ErrNoBootstrap
		}
		return
	}

	email := os.Getenv(envAdminEmail)
	if email == "" {
		email = defaultAdminEmail
	}

	password := os.Getenv(envAdminPassword)
	if password == "" {
		password = defaultAdminPassword
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	u := users.User{
		Email:        email,
		PasswordHash: hash,
		Roles: users.Roles{
			handlers.RoleAdmin: struct{}{},
		},
	}

	_, err = users.NewUserInt(context.Background(), tx, &u)
	if err != nil {
		return
	}

	_, err = tx.Exec("UPDATE bootstrap SET val = TRUE WHERE NOT val")
	return
}