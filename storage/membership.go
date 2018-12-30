package storage

import (
	"context"
	"database/sql"
	"time"

	"github.com/ecadlabs/auth/errors"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	uuid "github.com/satori/go.uuid"
)

type membershipModel struct {
	Membership_type   string         `db:"mem_type"`
	TenantID          uuid.UUID      `db:"tenant_id"`
	Membership_status string         `db:"mem_status"`
	UserID            uuid.UUID      `db:"user_id"`
	Added             time.Time      `db:"added"`
	Modified          time.Time      `db:"modified"`
	Roles             pq.StringArray `db:"roles"`
}

func (m *membershipModel) toMembership() *Membership {
	ret := &Membership{
		Membership_status: m.Membership_status,
		Membership_type:   m.Membership_status,
		UserID:            m.UserID,
		TenantID:          m.TenantID,
		Added:             m.Added,
		Modified:          m.Modified,
		Roles:             make(Roles, len(m.Roles)),
	}

	for _, r := range m.Roles {
		ret.Roles[r] = true
	}

	return ret
}

type Membership struct {
	Membership_type   string
	TenantID          uuid.UUID
	Membership_status string
	UserID            uuid.UUID
	Added             time.Time
	Modified          time.Time
	Roles             Roles
}

type MembershipStorage struct {
	DB *sqlx.DB
}

func (s *MembershipStorage) AddMembership(ctx context.Context, id uuid.UUID, user *User, status string, mem_type string) error {
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

	_, err = tx.ExecContext(ctx, "INSERT INTO membership (tenant_id, user_id, mem_status, mem_type) VALUES ($1, $2, $3, $4)", id, user.ID, status, mem_type)

	if err != nil {
		if isUniqueViolation(err, "membership_pkey") {
			err = errors.ErrMembershipExisits
		}
		return err
	}

	return nil
}

func (s *MembershipStorage) GetMembership(ctx context.Context, id uuid.UUID, userId uuid.UUID) (*Membership, error) {
	model := membershipModel{}
	err := s.DB.GetContext(ctx, &model, "SELECT membership.*, ra.roles FROM membership LEFT JOIN (SELECT user_id, array_agg(role) AS roles FROM roles GROUP BY user_id) AS ra ON ra.user_id = user_id AND ra.tenant_id = tenant_id WHERE tenant_id = $1 AND user_id = $2", id, userId)

	if err != nil {
		if err == sql.ErrNoRows {
			err = errors.ErrMembershipNotFound
		}
		return nil, err
	}

	return model.toMembership(), nil
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
