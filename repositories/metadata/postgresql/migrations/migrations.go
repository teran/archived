package migrations

import (
	"database/sql"

	migrate "github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
)

func MigrateUp(dsn string) error {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return errors.Wrap(err, "error opening database connection")
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return errors.Wrap(err, "error pinging database")
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return errors.Wrap(err, "error creating database instance")
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations/sql",
		"postgres", driver)
	if err != nil {
		return errors.Wrap(err, "error creating migrator instance")
	}

	return errors.Wrap(m.Up(), "error migrating database")
}
