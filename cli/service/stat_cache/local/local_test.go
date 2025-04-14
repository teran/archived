package local

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
	cache "github.com/teran/archived/cli/service/stat_cache"
)

func (s *localCacheRepositoryTestSuite) TestAll() {
	fi, err := os.Stat("local.go")
	s.Require().NoError(err)

	v, err := s.repo.Get(s.ctx, "local.go", fi)
	s.Require().NoError(err)
	s.Require().Empty(v)

	err = s.repo.Put(s.ctx, "local.go", fi, "deadbeef")
	s.Require().NoError(err)

	v, err = s.repo.Get(s.ctx, "local.go", fi)
	s.Require().NoError(err)
	s.Require().Equal("deadbeef", v)
}

// Definitions ...
type localCacheRepositoryTestSuite struct {
	suite.Suite

	ctx      context.Context
	cacheDir string
	repo     cache.CacheRepository
}

func (s *localCacheRepositoryTestSuite) SetupTest() {
	s.ctx = context.TODO()

	var err error
	s.cacheDir, err = os.MkdirTemp(os.TempDir(), "local-cache-repository-test-suite")
	s.Require().NoError(err)

	s.repo, err = New(s.cacheDir)
	s.Require().NoError(err)
}

func (s *localCacheRepositoryTestSuite) TearDownTest() {
	err := os.RemoveAll(s.cacheDir)
	s.Require().NoError(err)
}

func TestLocalCacheRepositoryTestSuite(t *testing.T) {
	suite.Run(t, &localCacheRepositoryTestSuite{})
}
