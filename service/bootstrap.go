package service

import (
	"context"
	"database/sql"
	"errors"

	"github.com/satori/go.uuid"

	"github.com/ecadlabs/auth/storage"
	"github.com/jmoiron/sqlx"
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

func (s *Service) Bootstrap(c *BootstrapConfig) (user *storage.User, err error) {
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

	for _, cUser := range c.Users {
		roles := make(storage.Roles)
		roles[cUser.Role] = true
		u := storage.CreateUser{
			ID:            uuid.FromStringOrNil(cUser.ID),
			Email:         cUser.Email,
			PasswordHash:  ([]byte)(cUser.Hash),
			EmailVerified: true, // Allow logging in !!!
			Roles:         roles,
			Type:          storage.AccountRegular,
		}

		_, err = storage.NewUserInt(context.Background(), tx, &u)
		if err != nil {
			return
		}
	}

	for _, tenant := range c.Tenants {
		u := storage.TenantModel{
			ID:   uuid.FromStringOrNil(tenant.ID),
			Name: tenant.Name,
		}

		_, err := s.storage.CreateTenantInt(context.Background(), tx, &u)
		if err != nil {
			return nil, err
		}
	}

	for _, member := range c.Membership {
		roles := make(storage.Roles)
		roles[member.Role] = true

		err := s.storage.AddMembershipInt(
			context.Background(),
			tx,
			uuid.FromStringOrNil(member.TenantID),
			uuid.FromStringOrNil(member.UserID),
			storage.ActiveState,
			storage.MemberMembership,
			roles,
		)

		if err != nil {
			return nil, err
		}
	}

	_, err = tx.Exec("UPDATE bootstrap SET val = TRUE WHERE NOT val")
	return
}
