package migrations

import (
	"context"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/suite"
	postgresApp "github.com/teran/go-docker-testsuite/applications/postgres"
)

func (s *migrateTestSuite) TestMigrateTwice() {
	err := migrateWithMigrationsPath(s.postgresDBApp.MustDSN("test_db"), "file://sql")
	s.Require().NoError(err)

	err = migrateWithMigrationsPath(s.postgresDBApp.MustDSN("test_db"), "file://sql")
	s.Require().NoError(err)
}

// Definitions ...
type migrateTestSuite struct {
	suite.Suite

	ctx           context.Context
	postgresDBApp postgresApp.PostgreSQL
}

func (s *migrateTestSuite) SetupTest() {
	s.ctx = context.Background()

	app, err := postgresApp.New(s.ctx)
	s.Require().NoError(err)

	s.postgresDBApp = app

	err = s.postgresDBApp.CreateDB(s.ctx, "test_db")
	s.Require().NoError(err)
}

func (s *migrateTestSuite) TearDownTest() {
	err := s.postgresDBApp.Close(s.ctx)
	s.Require().NoError(err)
}

func TestMigrateTestSuite(t *testing.T) {
	suite.Run(t, &migrateTestSuite{})
}
