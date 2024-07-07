package postgresql

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/suite"
	postgresApp "github.com/teran/go-docker-testsuite/applications/postgres"
	gtm "github.com/teran/go-time"

	"github.com/teran/archived/repositories/metadata"
	"github.com/teran/archived/repositories/metadata/postgresql/migrations"
)

// Definitions ...
type postgreSQLRepositoryTestSuite struct {
	suite.Suite

	ctx           context.Context
	postgresDBApp postgresApp.PostgreSQL
	db            *sql.DB
	repo          metadata.Repository
	tp            *gtm.TimeNowMock
}

func (s *postgreSQLRepositoryTestSuite) SetupSuite() {
	s.ctx = context.Background()

	app, err := postgresApp.New(s.ctx)
	s.Require().NoError(err)

	s.postgresDBApp = app
}

func (s *postgreSQLRepositoryTestSuite) SetupTest() {
	err := s.postgresDBApp.CreateDB(s.ctx, "test_db")
	s.Require().NoError(err)

	dsn, err := s.postgresDBApp.DSN("test_db")
	s.Require().NoError(err)

	err = migrations.MigrateUp(dsn)
	s.Require().NoError(err)

	db, err := sql.Open("postgres", dsn)
	s.Require().NoError(err)

	s.db = db

	s.tp = gtm.NewTimeNowMock()

	s.repo = newWithTimeProvider(s.db, s.tp.Now)
}

func (s *postgreSQLRepositoryTestSuite) TearDownTest() {
	s.repo = nil

	_, err := s.db.ExecContext(
		s.ctx,
		`SELECT
			pg_terminate_backend(pg_stat_activity.pid)
		FROM
			pg_stat_activity
		WHERE
			pg_stat_activity.datname = 'test_db'
		AND
			pid != pg_backend_pid();
	`)
	s.Require().NoError(err)

	err = s.db.Close()
	s.Require().NoError(err)

	err = s.postgresDBApp.DropDB(s.ctx, "test_db")
	s.Require().NoError(err)

	s.tp.AssertExpectations(s.T())
}

func (s *postgreSQLRepositoryTestSuite) TearDownSuite() {
	err := s.postgresDBApp.Close(s.ctx)
	s.Require().NoError(err)
}

func TestPostgreSQLRepositoryTestSuite(t *testing.T) {
	suite.Run(t, &postgreSQLRepositoryTestSuite{})
}
