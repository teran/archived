package migrations

import (
	"database/sql"

	migrate "github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func newMigrator(dsn, migrationsPath string) (*migrate.Migrate, error) {
	log.WithFields(log.Fields{
		"migrations_path": migrationsPath,
	}).Tracef("initializing migrator with migration path")

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, errors.Wrap(err, "error opening database connection")
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return nil, errors.Wrap(err, "error pinging database")
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return nil, errors.Wrap(err, "error creating database instance")
	}

	m, err := migrate.NewWithDatabaseInstance(
		migrationsPath,
		"postgres", driver)
	if err != nil {
		return nil, errors.Wrap(err, "error creating migrator instance")
	}

	return m, nil
}

func migrateUpWithMigrationsPath(dsn, migrationsPath string) error {
	log.WithFields(log.Fields{
		"migrations_path": migrationsPath,
	}).Debug("running up migrations with migrations path")

	m, err := newMigrator(dsn, migrationsPath)
	if err != nil {
		return err
	}

	if err = m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return errors.Wrap(err, "error migrating database")
	}

	return nil
}

func migrateDownWithMigrationsPath(dsn, migrationsPath string) error {
	log.WithFields(log.Fields{
		"migrations_path": migrationsPath,
	}).Debug("running down migrations with migrations path")

	m, err := newMigrator(dsn, migrationsPath)
	if err != nil {
		return err
	}

	if err = m.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return errors.Wrap(err, "error migrating database")
	}

	return nil
}

func MigrateUp(dsn string) error {
	return migrateUpWithMigrationsPath(dsn, "file://migrations/sql")
}

func MigrateDown(dsn string) error {
	return migrateDownWithMigrationsPath(dsn, "file://migrations/sql")
}
