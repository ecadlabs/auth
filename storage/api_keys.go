package storage

import (
	"context"
	"database/sql"

	"github.com/ecadlabs/auth/errors"
	"github.com/satori/go.uuid"
)

const apiKeyQuery = `
    SELECT
      service_account_keys.id,
      service_account_keys.membership_id,
      membership.user_id,
      membership.tenant_id,
      service_account_keys.added
    FROM
      service_account_keys
      INNER JOIN membership ON service_account_keys.membership_id = membership.id
      INNER JOIN users ON membership.user_id = users.id
      AND users.account_type = 'service'`

func (s *Storage) GetKey(ctx context.Context, userID, keyID uuid.UUID) (*APIKey, error) {
	var key APIKey
	if err := s.DB.GetContext(ctx, &key, apiKeyQuery+" WHERE service_account_keys.id = $1 AND users.id = $2", keyID, userID); err != nil {
		if err == sql.ErrNoRows {
			err = errors.ErrKeyNotFound
		}

		return nil, err
	}

	return &key, nil
}

func (s *Storage) GetKeys(ctx context.Context, uid uuid.UUID) ([]*APIKey, error) {
	var keys []*APIKey
	if err := s.DB.SelectContext(ctx, &keys, apiKeyQuery+" WHERE users.id = $1 ORDER BY service_account_keys.added", uid); err != nil {
		return nil, err
	}

	return keys, nil
}

func (s *Storage) NewKey(ctx context.Context, userID, tenantID uuid.UUID) (*APIKey, error) {
	q := `
		WITH k AS (
		  INSERT INTO
		    service_account_keys (membership_id)
		  SELECT
		    membership.id
		  FROM
		    membership
		    INNER JOIN users ON membership.user_id = users.id
		    AND users.account_type = 'service'
		  WHERE
		    users.id = $1
		    AND tenant_id = $2 RETURNING *
		)
		SELECT
		  k.id,
		  k.membership_id,
		  membership.user_id,
		  membership.tenant_id,
		  k.added
		FROM
		  k
		  INNER JOIN membership ON k.membership_id = membership.id`

	var key APIKey
	if err := s.DB.GetContext(ctx, &key, q, userID, tenantID); err != nil {
		if err == sql.ErrNoRows {
			err = errors.ErrMembershipNotFound
		}

		return nil, err
	}

	return &key, nil
}

func (s *Storage) DeleteKey(ctx context.Context, userID, keyID uuid.UUID) error {
	q := `
        DELETE FROM
          service_account_keys USING membership
        WHERE
          service_account_keys.membership_id = membership.id
          AND service_account_keys.id = $1
          AND membership.user_id = $2`

	res, err := s.DB.ExecContext(ctx, q, keyID, userID)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return errors.ErrKeyNotFound
	}

	return nil
}
