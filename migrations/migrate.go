package migrations

import (
	"database/sql"
	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database/postgres"
	"github.com/golang-migrate/migrate/source/go-bindata"
	"net/url"
	"strconv"
)

const (
	defaultConnectTimeout = 10
)

func NewDB(db *sql.DB) (*migrate.Migrate, error) {
	drv, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return nil, err
	}

	source, err := bindata.WithInstance(bindata.Resource(AssetNames(), func(name string) ([]byte, error) {
		return Asset(name)
	}))
	if err != nil {
		return nil, err
	}

	return migrate.NewWithInstance("go-bindata", source, "postgres", drv)
}

func New(databaseUrl string) (*migrate.Migrate, error) {
	// Set connection timeout
	url, err := url.Parse(databaseUrl)
	if err != nil {
		return nil, err
	}

	q := url.Query()
	if _, ok := q["connect_timeout"]; !ok {
		q["connect_timeout"] = []string{strconv.FormatInt(defaultConnectTimeout, 10)}
	}
	url.RawQuery = q.Encode()

	source, err := bindata.WithInstance(bindata.Resource(AssetNames(), func(name string) ([]byte, error) {
		return Asset(name)
	}))
	if err != nil {
		return nil, err
	}

	return migrate.NewWithSourceInstance("go-bindata", source, url.String())
}