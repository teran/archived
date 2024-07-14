package migrations

import (
	"database/sql"

	migrate "github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
)

func migrateWithMigrationsPath(dsn, migrationsPath string) error {
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
		migrationsPath,
		"postgres", driver)
	if err != nil {
		return errors.Wrap(err, "error creating migrator instance")
	}

	if err = m.Up(); err != nil && err != migrate.ErrNoChange {
		return errors.Wrap(err, "error migrating database")
	}

	return nil
}

func MigrateUp(dsn string) error {
	return migrateWithMigrationsPath(dsn, "file://migrations/sql")
}
