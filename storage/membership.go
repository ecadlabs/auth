package storage

import (
	"context"
	"database/sql"
	"time"

	"github.com/ecadlabs/auth/errors"
	"github.com/jmoiron/sqlx"
	uuid "github.com/satori/go.uuid"
)

type Membership struct {
	Membership_type   string    `db:"mem_type"`
	TenantID          uuid.UUID `db:"tenant_id"`
	Membership_status string    `db:"mem_status"`
	UserID            uuid.UUID `db:"user_id"`
	Added             time.Time `db:"added"`
	Modified          time.Time `db:"modified"`
}

type MembershipStorage struct {
	DB *sqlx.DB
}

type addMembership struct {
	TenantId uuid.UUID `db:tenant_id schema:tenant_id`
	UserId   uuid.UUID `db:user_id schema:user_id`
	Status   string    `db:mem_status schema:mem_status`
}

func (s *MembershipStorage) AddMembership(ctx context.Context, id uuid.UUID, user *User, status string) error {
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

	_, err = tx.ExecContext(ctx, "INSERT INTO membership (tenant_id, user_id, mem_status) VALUES ($1, $2, $3)", id, user.ID, status)

	if err != nil {
		if isUniqueViolation(err, "membership_pkey") {
			err = errors.ErrMembershipExisits
		}
		return err
	}

	return nil
}

func (s *MembershipStorage) GetMembership(ctx context.Context, id uuid.UUID, userId uuid.UUID) (*Membership, error) {
	model := Membership{}
	err := s.DB.GetContext(ctx, &model, "SELECT membership.* FROM membership WHERE tenant_id = $1 AND user_id = $2", id, userId)

	if err != nil {
		if err == sql.ErrNoRows {
			err = errors.ErrMembershipNotFound
		}
		return nil, err
	}

	return &model, nil
}

func (s *MembershipStorage) UpdateMembership(ctx context.Context, id uuid.UUID, userId uuid.UUID, status string) error {
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

	_, err = tx.ExecContext(ctx, "UPDATE membership SET mem_status = $3 WHERE tenant_id = $1 AND user_id = $2", id, userId, status)

	if err != nil {
		return err
	}

	return nil
}
