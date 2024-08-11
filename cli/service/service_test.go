package service

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	echo "github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
	ptr "github.com/teran/go-ptr"

	cacheMock "github.com/teran/archived/cli/service/stat_cache/mock"
)

func init() {
	log.SetLevel(log.TraceLevel)
}

func (s *serviceTestSuite) TestCreateContainer() {
	s.cliMock.On("CreateContainer", "test-container").Return(nil).Once()

	fn := s.svc.CreateContainer("test-container")
	s.Require().NoError(fn(s.ctx))
}

func (s *serviceTestSuite) TestListContainers() {
	s.cliMock.On("ListContainers").Return([]string{"container1", "container2"}, nil).Once()

	fn := s.svc.ListContainers()
	s.Require().NoError(fn(s.ctx))
}

func (s *serviceTestSuite) TestDeleteContainer() {
	s.cliMock.On("DeleteContainer", "test-container1").Return(nil).Once()

	fn := s.svc.DeleteContainer("test-container1")
	s.Require().NoError(fn(s.ctx))
}

func (s *serviceTestSuite) TestCreateVersion() {
	s.cliMock.On("CreateVersion", "container1").Return("version_id", nil).Once()

	fn := s.svc.CreateVersion("container1", false, nil, nil)
	s.Require().NoError(fn(s.ctx))
}

func (s *serviceTestSuite) TestCreateVersionAndPublish() {
	s.cliMock.On("CreateVersion", "container1").Return("version_id", nil).Once()
	s.cliMock.On("PublishVersion", "container1", "version_id").Return(nil).Once()

	fn := s.svc.CreateVersion("container1", true, nil, nil)
	s.Require().NoError(fn(s.ctx))
}

func (s *serviceTestSuite) TestCreateVersionAndPublishWithEmptyPath() {
	s.cliMock.On("CreateVersion", "container1").Return("version_id", nil).Once()
	s.cliMock.On("PublishVersion", "container1", "version_id").Return(nil).Once()

	fn := s.svc.CreateVersion("container1", true, ptr.String(""), nil)
	s.Require().NoError(fn(s.ctx))
}

func (s *serviceTestSuite) TestCreateVersionFromDirAndPublish() {
	s.cacheMock.On("Get", "testdata/somefile1").Return("", nil).Once()
	s.cacheMock.On("Get", "testdata/somefile2").Return("", nil).Once()
	s.cacheMock.On("Put", "testdata/somefile1", "a883dafc480d466ee04e0d6da986bd78eb1fdd2178d04693723da3a8f95d42f4").Return(nil).Once()
	s.cacheMock.On("Put", "testdata/somefile2", "ff5a972ba33179c7ec67c73e00a362b629c489f9d7c86489644db2bcd8c62c61").Return(nil).Once()

	s.cliMock.On("CreateVersion", "container1").Return("version_id", nil, nil).Once()
	s.cliMock.
		On("CreateObject", "container1", "version_id", "somefile1", "a883dafc480d466ee04e0d6da986bd78eb1fdd2178d04693723da3a8f95d42f4", int64(5)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", "container1", "version_id", "somefile2", "ff5a972ba33179c7ec67c73e00a362b629c489f9d7c86489644db2bcd8c62c61", int64(5)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.On("PublishVersion", "container1", "version_id").Return(nil).Once()

	fn := s.svc.CreateVersion("container1", true, ptr.String("testdata"), nil)
	s.Require().NoError(fn(s.ctx))
}

func (s *serviceTestSuite) TestCreateVersionFromYumRepoAndPublish() {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Static("/", "../yum/testdata/repo")

	srv := httptest.NewServer(e)
	defer srv.Close()

	e2 := echo.New()
	e2.Use(middleware.Logger())
	e2.Use(middleware.Recover())
	e2.PUT("/upload", func(c echo.Context) error {
		if c.Request().Header.Get("Content-Length") != "6156" {
			return c.NoContent(http.StatusConflict)
		}

		if c.Request().Header.Get("Content-Type") != "multipart/form-data" {
			return c.NoContent(http.StatusConflict)
		}
		return nil
	})

	uploadSrv := httptest.NewServer(e2)
	defer uploadSrv.Close()

	s.cliMock.On("CreateVersion", "container1").Return("version_id", nil, nil).Once()
	s.cliMock.
		On("CreateObject", "container1", "version_id", "repodata/2267234d92017b049818be743f720f37c176a3b3bb3e802ee4d5cd0090651091-primary.xml.gz", "2267234d92017b049818be743f720f37c176a3b3bb3e802ee4d5cd0090651091", int64(720)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", "container1", "version_id", "repodata/1b4aca205bffe8d65f33b066e3f9965cb4c009e3c94b3f296cce8bff166ad8ed-primary.sqlite.bz2", "1b4aca205bffe8d65f33b066e3f9965cb4c009e3c94b3f296cce8bff166ad8ed", int64(1995)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", "container1", "version_id", "repodata/314e73564000b8a68848551ce0fa9b36e11ed609698f232fa9ab5810ec531de1-filelists.xml.gz", "314e73564000b8a68848551ce0fa9b36e11ed609698f232fa9ab5810ec531de1", int64(313)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", "container1", "version_id", "repodata/2623c0a1472f574989dcba85417e8ce27b87983bba12922a6d91d574e617d2f6-filelists.sqlite.bz2", "2623c0a1472f574989dcba85417e8ce27b87983bba12922a6d91d574e617d2f6", int64(858)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", "container1", "version_id", "repodata/e3984def0f3b5ce1b174fad2f6eb3c05829633d2d5d5d8ba05c9720ad59046e7-other.xml.gz", "e3984def0f3b5ce1b174fad2f6eb3c05829633d2d5d5d8ba05c9720ad59046e7", int64(281)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", "container1", "version_id", "repodata/64f4875d92a3672f62a2d15d5f0ae6f0806451f42403bd07105214e1c9f4f0d7-other.sqlite.bz2", "64f4875d92a3672f62a2d15d5f0ae6f0806451f42403bd07105214e1c9f4f0d7", int64(749)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", "container1", "version_id", "repodata/repomd.xml", "ad1ff2a7e93b614596a9c432f85b141df86e2c010b6591a04c8b011051bd739c", int64(3069)).
		Return(ptr.String(""), nil).
		Once()

	s.cliMock.
		On("CreateObject", "container1", "version_id", "Packages/testpkg-1-1.src.rpm", "684303227d799ffe1f0b39e030a12ad249931a11ec1690e2079f981cc16d8c52", int64(6156)).
		Return(ptr.String(uploadSrv.URL+"/upload"), nil).
		Once()
	s.cliMock.
		On("CreateObject", "container1", "version_id", "Packages/testpkg-1-1.x86_64.rpm", "d9ae5e56ea38d2ac470f320cade63663dae6ab8b8e1630b2fd5a3c607f45e2ee", int64(6722)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.On("PublishVersion", "container1", "version_id").Return(nil).Once()

	fn := s.svc.CreateVersion("container1", true, ptr.String(""), ptr.String(srv.URL))
	s.Require().NoError(fn(s.ctx))
}

func (s *serviceTestSuite) TestCreateVersionFromYumRepoAndPublishSHA1() {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Static("/", "../yum/testdata/repo-sha1")

	srv := httptest.NewServer(e)
	defer srv.Close()

	e2 := echo.New()
	e2.Use(middleware.Logger())
	e2.Use(middleware.Recover())
	e2.PUT("/upload", func(c echo.Context) error {
		if c.Request().Header.Get("Content-Length") != "6156" {
			return c.NoContent(http.StatusConflict)
		}

		if c.Request().Header.Get("Content-Type") != "multipart/form-data" {
			return c.NoContent(http.StatusConflict)
		}
		return nil
	})

	uploadSrv := httptest.NewServer(e2)
	defer uploadSrv.Close()

	s.cliMock.On("CreateVersion", "container1").Return("version_id", nil, nil).Once()
	s.cliMock.
		On("CreateObject", "container1", "version_id", "repodata/80779e2ab55e25a77124d370de1d08deae8f1cc6-primary.xml.gz", "1c07f3f3f0e6d09972c1d7852d1dbc9715d6fbdceee66c50e8356d1e69502d3b", int64(688)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", "container1", "version_id", "repodata/e7a8a53e7398f6c22894718ea227fea60f2b78ba-primary.sqlite.bz2", "c9b8ce03b503e29d9ec2faa2328e4f2082f0a5f71478ca6cb2f1a3ab75e676bc", int64(1937)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", "container1", "version_id", "repodata/4a11e3eeb25d21b08f41e5578d702d2bea21a2e7-filelists.xml.gz", "b56801c0a86f9a0136953e8c8e59cd35c1f18fc41e70ba8fcdcccfee068dfc8a", int64(282)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", "container1", "version_id", "repodata/c66ce2caa41ed83879f9b3dd9f40e61c65af499e-filelists.sqlite.bz2", "59bd3edd4edacac87e5e15494698f34a7f52277691635f927c185e92a681d9ee", int64(787)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", "container1", "version_id", "repodata/fdedb6ce109127d52228d01b0239010ddca14c8f-other.xml.gz", "56e566dfc63b0a7056b21cec661717a411f68cf98747d9a719557bce3a8ac41a", int64(247)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", "container1", "version_id", "repodata/b31561a27d014d35b59b27c27859bb1c17ac573e-other.sqlite.bz2", "7eec446e0036d356d8e5694047d9fdb6af00f2fc62993b854232830cf9dbcff8", int64(669)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", "container1", "version_id", "repodata/repomd.xml", "9f18801e8532f631e308a130a347f66eb3900d054df1d66dff53a69aa5b9e7d3", int64(2601)).
		Return(ptr.String(""), nil).
		Once()

	s.cliMock.
		On("CreateObject", "container1", "version_id", "Packages/testpkg-1-1.src.rpm", "684303227d799ffe1f0b39e030a12ad249931a11ec1690e2079f981cc16d8c52", int64(6156)).
		Return(ptr.String(uploadSrv.URL+"/upload"), nil).
		Once()
	s.cliMock.
		On("CreateObject", "container1", "version_id", "Packages/testpkg-1-1.x86_64.rpm", "d9ae5e56ea38d2ac470f320cade63663dae6ab8b8e1630b2fd5a3c607f45e2ee", int64(6722)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.On("PublishVersion", "container1", "version_id").Return(nil).Once()

	fn := s.svc.CreateVersion("container1", true, ptr.String(""), ptr.String(srv.URL))
	s.Require().NoError(fn(s.ctx))
}

func (s *serviceTestSuite) TestDeleteVersion() {
	s.cliMock.On("DeleteVersion", "container1", "version1").Return(nil).Once()

	fn := s.svc.DeleteVersion("container1", "version1")
	s.Require().NoError(fn(s.ctx))
}

func (s *serviceTestSuite) TestListVersions() {
	s.cliMock.On("ListVersions", "container1").Return([]string{"version1", "version2", "version3"}, nil).Once()

	fn := s.svc.ListVersions("container1")
	s.Require().NoError(fn(s.ctx))
}

func (s *serviceTestSuite) TestPublishVersion() {
	s.cliMock.On("PublishVersion", "container1", "version1").Return(nil).Once()

	fn := s.svc.PublishVersion("container1", "version1")
	s.Require().NoError(fn(s.ctx))
}

func (s *serviceTestSuite) TestCreateObjectWithoutEndingSlashInThePath() {
	s.cacheMock.On("Get", "testdata/somefile1").Return("", nil).Once()
	s.cacheMock.On("Get", "testdata/somefile2").Return("", nil).Once()
	s.cacheMock.On("Put", "testdata/somefile1", "a883dafc480d466ee04e0d6da986bd78eb1fdd2178d04693723da3a8f95d42f4").Return(nil).Once()
	s.cacheMock.On("Put", "testdata/somefile2", "ff5a972ba33179c7ec67c73e00a362b629c489f9d7c86489644db2bcd8c62c61").Return(nil).Once()

	s.cliMock.
		On("CreateObject", "container1", "version1", "somefile1", "a883dafc480d466ee04e0d6da986bd78eb1fdd2178d04693723da3a8f95d42f4", int64(5)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", "container1", "version1", "somefile2", "ff5a972ba33179c7ec67c73e00a362b629c489f9d7c86489644db2bcd8c62c61", int64(5)).
		Return(ptr.String(""), nil).
		Once()

	fn := s.svc.CreateObject("container1", "version1", "testdata")
	s.Require().NoError(fn(s.ctx))
}

func (s *serviceTestSuite) TestCreateObjectWithEndingSlashInThePath() {
	s.cacheMock.On("Get", "testdata/somefile1").Return("", nil).Once()
	s.cacheMock.On("Get", "testdata/somefile2").Return("", nil).Once()
	s.cacheMock.On("Put", "testdata/somefile1", "a883dafc480d466ee04e0d6da986bd78eb1fdd2178d04693723da3a8f95d42f4").Return(nil).Once()
	s.cacheMock.On("Put", "testdata/somefile2", "ff5a972ba33179c7ec67c73e00a362b629c489f9d7c86489644db2bcd8c62c61").Return(nil).Once()

	s.cliMock.
		On("CreateObject", "container1", "version1", "somefile1", "a883dafc480d466ee04e0d6da986bd78eb1fdd2178d04693723da3a8f95d42f4", int64(5)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", "container1", "version1", "somefile2", "ff5a972ba33179c7ec67c73e00a362b629c489f9d7c86489644db2bcd8c62c61", int64(5)).
		Return(ptr.String(""), nil).
		Once()

	fn := s.svc.CreateObject("container1", "version1", "testdata/")
	s.Require().NoError(fn(s.ctx))
}

func (s *serviceTestSuite) TestCreateObjectWithCache() {
	s.cacheMock.On("Get", "testdata/somefile1").Return("a883dafc480d466ee04e0d6da986bd78eb1fdd2178d04693723da3a8f95d42f4", nil).Once()
	s.cacheMock.On("Get", "testdata/somefile2").Return("ff5a972ba33179c7ec67c73e00a362b629c489f9d7c86489644db2bcd8c62c61", nil).Once()

	s.cliMock.
		On("CreateObject", "container1", "version1", "somefile1", "a883dafc480d466ee04e0d6da986bd78eb1fdd2178d04693723da3a8f95d42f4", int64(5)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", "container1", "version1", "somefile2", "ff5a972ba33179c7ec67c73e00a362b629c489f9d7c86489644db2bcd8c62c61", int64(5)).
		Return(ptr.String(""), nil).
		Once()

	fn := s.svc.CreateObject("container1", "version1", "testdata")
	s.Require().NoError(fn(s.ctx))
}

func (s *serviceTestSuite) TestCreateObjectWithUploadURL() {
	s.cacheMock.On("Get", "testdata/somefile1").Return("a883dafc480d466ee04e0d6da986bd78eb1fdd2178d04693723da3a8f95d42f4", nil).Once()
	s.cacheMock.On("Get", "testdata/somefile2").Return("ff5a972ba33179c7ec67c73e00a362b629c489f9d7c86489644db2bcd8c62c61", nil).Once()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, err := io.ReadAll(r.Body)
		s.Require().NoError(err)
		defer r.Body.Close()

		s.Require().Equal("1234\n", string(data))

		s.Require().Equal("/test-url", r.RequestURI)
		s.Require().Equal("multipart/form-data", r.Header.Get("Content-Type"))
	}))
	defer srv.Close()

	s.cliMock.
		On("CreateObject", "container1", "version1", "somefile1", "a883dafc480d466ee04e0d6da986bd78eb1fdd2178d04693723da3a8f95d42f4", int64(5)).
		Return(ptr.String(srv.URL+"/test-url"), nil).
		Once()
	s.cliMock.
		On("CreateObject", "container1", "version1", "somefile2", "ff5a972ba33179c7ec67c73e00a362b629c489f9d7c86489644db2bcd8c62c61", int64(5)).
		Return(ptr.String(""), nil).
		Once()

	fn := s.svc.CreateObject("container1", "version1", "testdata")
	s.Require().NoError(fn(s.ctx))
}

func (s *serviceTestSuite) TestDeleteObject() {
	s.cliMock.On("DeleteObject", "container1", "version1", "key1").Return(nil).Once()

	fn := s.svc.DeleteObject("container1", "version1", "key1")
	s.Require().NoError(fn(s.ctx))
}

func (s *serviceTestSuite) TestListObjects() {
	s.cliMock.On("ListObjects", "container1", "version1").Return([]string{"obj1", "obj2", "obj3"}, nil).Once()

	fn := s.svc.ListObjects("container1", "version1")
	s.Require().NoError(fn(s.ctx))
}

func (s *serviceTestSuite) TestGetObjectURL() {
	s.cliMock.On("GetObjectURL", "container1", "version1", "key1").Return("https://example.com", nil).Once()

	fn := s.svc.GetObjectURL("container1", "version1", "key1")
	s.Require().NoError(fn(s.ctx))
}

// Definitions ...
type serviceTestSuite struct {
	suite.Suite

	ctx       context.Context
	cliMock   *protoClientMock
	cacheMock *cacheMock.Mock
	svc       Service
}

func (s *serviceTestSuite) SetupTest() {
	s.ctx = context.Background()

	s.cliMock = newMock()
	s.cacheMock = cacheMock.New()

	s.svc = New(s.cliMock, s.cacheMock)
}

func (s *serviceTestSuite) TearDownTest() {
	s.cliMock.AssertExpectations(s.T())
	s.cacheMock.AssertExpectations(s.T())
}

func TestServiceTestSuite(t *testing.T) {
	suite.Run(t, &serviceTestSuite{})
}
