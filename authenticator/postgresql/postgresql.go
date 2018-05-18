package postgresql

import (
	"context"
	"database/sql"
	"git.ecadlabs.com/ecad/auth/authenticator"
	"git.ecadlabs.com/ecad/auth/authenticator/postgresql/migrations"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
	"time"
)

type User struct {
	Id           uuid.UUID `db:"id"`
	Email        string    `db:"email"`
	PasswordHash []byte    `db:"password_hash"`
	FirstName    string    `db:"first_name"`
	LastName     string    `db:"last_name"`
	Added        time.Time `db:"added"`
	Modified     time.Time `db:"modified"`
}

type Error struct {
	Err error
}

func (e Error) Rejected() bool {
	return e.Err == sql.ErrNoRows || e.Err == bcrypt.ErrMismatchedHashAndPassword
}

func (e Error) Error() string {
	return "authenticator/postgresql: " + e.Err.Error()
}

type Result struct {
	*User
	a *Authenticator
}

type Authenticator struct {
	DB        *sqlx.DB
	Namespace string
}

func (a *Authenticator) Authenticate(ctx context.Context, cred *authenticator.Credentials) (authenticator.Result, error) {
	var u User
	if err := a.DB.GetContext(ctx, &u, "SELECT * FROM users WHERE email=$1", cred.ID); err != nil {
		return nil, Error{Err: err}
	}

	if err := bcrypt.CompareHashAndPassword(u.PasswordHash, cred.Secret); err != nil {
		return nil, Error{Err: err}
	}

	return &Result{
		User: &u,
		a:    a,
	}, nil
}

func (a *Authenticator) Ping(ctx context.Context) error {
	if err := a.DB.PingContext(ctx); err != nil {
		return Error{Err: err}
	}

	return nil
}

func NewWithMigration(driverName, dataSourceName, namespace string) (*Authenticator, error) {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return nil, err
	}

	if err := migrations.Migrate(driverName); err != nil {
		return nil, err
	}

	return &Authenticator{
		DB:        sqlx.NewDb(db, driverName),
		Namespace: namespace,
	}, nil
}

func New(driverName, dataSourceName, namespace string) (*Authenticator, error) {
	db, err := sqlx.Open(driverName, dataSourceName)
	if err != nil {
		return nil, err
	}

	return &Authenticator{
		DB:        db,
		Namespace: namespace,
	}, nil
}

func (r *Result) Claims() map[string]interface{} {
	ret := map[string]interface{}{
		"sub": r.User.Id,
	}

	if r.a.Namespace != "" {
		ret[r.a.Namespace+"/email"] = r.User.Email
		ret[r.a.Namespace+"/first_name"] = r.User.FirstName
		ret[r.a.Namespace+"/last_name"] = r.User.LastName
	}

	return ret
}

var _ authenticator.Authenticator = &Authenticator{}
