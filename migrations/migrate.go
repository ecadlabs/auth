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

func MigrateDB(db *sql.DB) error {
	source, err := bindata.WithInstance(bindata.Resource(AssetNames(), func(name string) ([]byte, error) {
		return Asset(name)
	}))
	if err != nil {
		return err
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithInstance("go-bindata", source, "postgres", driver)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil {
		return err
	}

	return nil
}

func Migrate(databaseUrl string) error {
	// Set connection timeout
	url, err := url.Parse(databaseUrl)
	if err != nil {
		return err
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
		return err
	}

	m, err := migrate.NewWithSourceInstance("go-bindata", source, url.String())
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	if _, err := m.Close(); err != nil {
		return err
	}

	return nil
}
