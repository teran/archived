package yum

import (
	"context"
	"net/http/httptest"
	"testing"

	echo "github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/stretchr/testify/suite"
	"github.com/teran/archived/cli/service/source"
	ptr "github.com/teran/go-ptr"
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

func (s *yumTestSuite) TestRepoWithGPGKey() {
	result := []source.Object{}

	repo := New(s.srv.URL+"/repo-signed/", ptr.String(s.srv.URL+"/gpg/somekey.gpg"), nil)
	err := repo.Process(s.ctx, func(ctx context.Context, obj source.Object) error {
		result = append(result, obj)
		return nil
	})
	s.Require().NoError(err)
	s.Require().Len(result, 9)
}

func (s *yumTestSuite) TestRepoWithGPGKeyAndChecksum() {
	result := []source.Object{}

	repo := New(
		s.srv.URL+"/repo-signed/",
		ptr.String(s.srv.URL+"/gpg/somekey.gpg"),
		ptr.String("aa392a2005c38f10ce21034d6d1aaace5bbee1c3d98ac1ee06a42336d741473e"),
	)
	err := repo.Process(s.ctx, func(ctx context.Context, obj source.Object) error {
		result = append(result, obj)
		return nil
	})
	s.Require().NoError(err)
	s.Require().Len(result, 9)
}

func (s *yumTestSuite) TestRepoWithGPGKeyAndInvalidChecksum() {
	result := []source.Object{}

	repo := New(
		s.srv.URL+"/repo-signed/",
		ptr.String(s.srv.URL+"/gpg/somekey.gpg"),
		ptr.String("invalid"),
	)
	err := repo.Process(s.ctx, func(ctx context.Context, obj source.Object) error {
		result = append(result, obj)
		return nil
	})
	s.Require().Error(err)
	s.Require().Equal(
		"GPG Key checksum mismatch",
		err.Error(),
	)
}

func (s *yumTestSuite) TestRepoSHA1() {
	result := []source.Object{}

	repo := New(s.srv.URL+"/repo/", nil, nil)
	err := repo.Process(s.ctx, func(ctx context.Context, obj source.Object) error {
		result = append(result, obj)
		return nil
	})
	s.Require().NoError(err)
	s.Require().Len(result, 67)
}

func (s *yumTestSuite) TestRepoWithGPGKeyMissedSignature() {
	repo := New(s.srv.URL+"/repo/", ptr.String(s.srv.URL+"/gpg/somekey.gpg"), nil)
	err := repo.Process(s.ctx, func(ctx context.Context, obj source.Object) error {
		return nil
	})
	s.Require().Error(err)
	s.Require().Equal(
		"error verifying package signature: SRPMS/testpkg1-1-1.src.rpm: keyid 93a645a017898e46 not found",
		err.Error(),
	)
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
