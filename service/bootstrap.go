package service

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"strings"

	"github.com/ecadlabs/auth/storage"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

const (
	envAdminEmail    = "AUTH_ADMIN_EMAIL"
	envAdminPassword = "AUTH_ADMIN_PASSWORD"
	envAdminRoles    = "AUTH_ADMIN_ROLES"
)

const (
	defaultAdminEmail    = "admin@admin"
	defaultAdminPassword = "admin"
	defaultAdminRoles    = "admin"
)

var ErrNoBootstrap = errors.New("No bootstrapping")

func (s *Service) Bootstrap() (user *storage.User, err error) {
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

	rolesList := os.Getenv(envAdminRoles)
	if rolesList == "" {
		rolesList = defaultAdminRoles
	}

	roles := make(storage.Roles)
	for _, r := range strings.Split(rolesList, ":") {
		roles[r] = struct{}{}
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	u := storage.CreateUser{
		Email:         email,
		PasswordHash:  hash,
		EmailVerified: true, // Allow logging in !!!
		Roles:         roles,
	}

	user, err = storage.NewUserInt(context.Background(), tx, &u)
	if err != nil {
		return
	}

	_, err = tx.Exec("UPDATE bootstrap SET val = TRUE WHERE NOT val")
	return
}
