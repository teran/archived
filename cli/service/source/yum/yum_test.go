package yum

import (
	"context"
	"net/http/httptest"
	"testing"

	echo "github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/stretchr/testify/suite"
	"github.com/teran/archived/cli/service/source"
)

func (s *yumTestSuite) TestRepo() {
	result := []source.Object{}

	repo := New(s.srv.URL+"/repo/", nil, nil)
	err := repo.Process(s.ctx, func(ctx context.Context, obj source.Object) error {
		result = append(result, obj)
		return nil
	})
	s.Require().NoError(err)
	s.Require().Len(result, 67)
}

// Definitions ...
type yumTestSuite struct {
	suite.Suite

	ctx context.Context
	srv *httptest.Server
}

func (s *yumTestSuite) SetupTest() {
	s.ctx = context.Background()

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Static("/", "testdata/")

	s.srv = httptest.NewServer(e)
}

func (s *yumTestSuite) TearDownTest() {
	s.srv.Close()
}

func TestYumTestSuite(t *testing.T) {
	suite.Run(t, &yumTestSuite{})
}
