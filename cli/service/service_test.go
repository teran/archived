package service

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
	ptr "github.com/teran/go-ptr"

	sourceMock "github.com/teran/archived/cli/service/source/mock"
	cacheMock "github.com/teran/archived/cli/service/stat_cache/mock"
)

const (
	defaultNamespace = "default"

	mimeTypeMultipartFormData = "multipart/form-data"
)

func init() {
	log.SetLevel(log.TraceLevel)
}

func (s *serviceTestSuite) TestCreateNamespace() {
	s.cliMock.On("CreateNamespace", "test-namespace").Return(nil).Once()

	fn := s.svc.CreateNamespace("test-namespace")
	s.Require().NoError(fn(s.ctx))
}

func (s *serviceTestSuite) TestRenameNamespace() {
	s.cliMock.On("RenameNamespace", "old-name", "new-name").Return(nil).Once()

	fn := s.svc.RenameNamespace("old-name", "new-name")
	s.Require().NoError(fn(s.ctx))
}

func (s *serviceTestSuite) TestListNamespaces() {
	s.cliMock.On("ListNamespaces").Return([]string{"namespace1", "namespace2"}, nil).Once()

	fn := s.svc.ListNamespaces()
	s.Require().NoError(fn(s.ctx))
}

func (s *serviceTestSuite) TestDeleteNamespace() {
	s.cliMock.On("DeleteNamespace", "test-namespace").Return(nil).Once()

	fn := s.svc.DeleteNamespace("test-namespace")
	s.Require().NoError(fn(s.ctx))
}

func (s *serviceTestSuite) TestCreateContainer() {
	s.cliMock.On("CreateContainer", defaultNamespace, "test-container").Return(nil).Once()

	fn := s.svc.CreateContainer(defaultNamespace, "test-container")
	s.Require().NoError(fn(s.ctx))
}

func (s *serviceTestSuite) TestMoveContainer() {
	s.cliMock.On("MoveContainer", defaultNamespace, "test-container", "new-namespace").Return(nil).Once()

	fn := s.svc.MoveContainer(defaultNamespace, "test-container", "new-namespace")
	s.Require().NoError(fn(s.ctx))
}

func (s *serviceTestSuite) TestRenameContainer() {
	s.cliMock.On("RenameContainer", defaultNamespace, "old-name", "new-name").Return(nil).Once()

	fn := s.svc.RenameContainer(defaultNamespace, "old-name", "new-name")
	s.Require().NoError(fn(s.ctx))
}

func (s *serviceTestSuite) TestListContainers() {
	s.cliMock.On("ListContainers", defaultNamespace).Return([]string{"container1", "container2"}, nil).Once()

	fn := s.svc.ListContainers(defaultNamespace)
	s.Require().NoError(fn(s.ctx))
}

func (s *serviceTestSuite) TestDeleteContainer() {
	s.cliMock.On("DeleteContainer", defaultNamespace, "test-container1").Return(nil).Once()

	fn := s.svc.DeleteContainer(defaultNamespace, "test-container1")
	s.Require().NoError(fn(s.ctx))
}

func (s *serviceTestSuite) TestCreateVersion() {
	s.cliMock.On("CreateVersion", defaultNamespace, "container1").Return("version_id", nil).Once()

	s.sourceMock.On("Process").Return(nil).Once()

	fn := s.svc.CreateVersion(defaultNamespace, "container1", false, s.sourceMock)
	s.Require().NoError(fn(s.ctx))
}

func (s *serviceTestSuite) TestCreateVersionAndPublish() {
	s.cliMock.On("CreateVersion", defaultNamespace, "container1").Return("version_id", nil).Once()
	s.cliMock.On("PublishVersion", defaultNamespace, "container1", "version_id").Return(nil).Once()

	s.sourceMock.On("Process").Return(nil).Once()

	fn := s.svc.CreateVersion(defaultNamespace, "container1", true, s.sourceMock)
	s.Require().NoError(fn(s.ctx))
}

func (s *serviceTestSuite) TestDeleteVersion() {
	s.cliMock.On("DeleteVersion", defaultNamespace, "container1", "version1").Return(nil).Once()

	fn := s.svc.DeleteVersion(defaultNamespace, "container1", "version1")
	s.Require().NoError(fn(s.ctx))
}

func (s *serviceTestSuite) TestListVersions() {
	s.cliMock.On("ListVersions", defaultNamespace, "container1").Return([]string{"version1", "version2", "version3"}, nil).Once()

	fn := s.svc.ListVersions(defaultNamespace, "container1")
	s.Require().NoError(fn(s.ctx))
}

func (s *serviceTestSuite) TestPublishVersion() {
	s.cliMock.On("PublishVersion", defaultNamespace, "container1", "version1").Return(nil).Once()

	fn := s.svc.PublishVersion(defaultNamespace, "container1", "version1")
	s.Require().NoError(fn(s.ctx))
}

func (s *serviceTestSuite) TestCreateObjectWithoutEndingSlashInThePath() {
	s.cacheMock.On("Get", "testdata/repo/somefile1").Return("", nil).Once()
	s.cacheMock.On("Get", "testdata/repo/somefile2").Return("", nil).Once()
	s.cacheMock.On("Put", "testdata/repo/somefile1", "a883dafc480d466ee04e0d6da986bd78eb1fdd2178d04693723da3a8f95d42f4").Return(nil).Once()
	s.cacheMock.On("Put", "testdata/repo/somefile2", "ff5a972ba33179c7ec67c73e00a362b629c489f9d7c86489644db2bcd8c62c61").Return(nil).Once()

	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version1", "somefile1", "a883dafc480d466ee04e0d6da986bd78eb1fdd2178d04693723da3a8f95d42f4", uint64(5)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version1", "somefile2", "ff5a972ba33179c7ec67c73e00a362b629c489f9d7c86489644db2bcd8c62c61", uint64(5)).
		Return(ptr.String(""), nil).
		Once()

	fn := s.svc.CreateObject(defaultNamespace, "container1", "version1", "testdata/repo")
	s.Require().NoError(fn(s.ctx))
}

func (s *serviceTestSuite) TestCreateObjectWithEndingSlashInThePath() {
	s.cacheMock.On("Get", "testdata/repo/somefile1").Return("", nil).Once()
	s.cacheMock.On("Get", "testdata/repo/somefile2").Return("", nil).Once()
	s.cacheMock.On("Put", "testdata/repo/somefile1", "a883dafc480d466ee04e0d6da986bd78eb1fdd2178d04693723da3a8f95d42f4").Return(nil).Once()
	s.cacheMock.On("Put", "testdata/repo/somefile2", "ff5a972ba33179c7ec67c73e00a362b629c489f9d7c86489644db2bcd8c62c61").Return(nil).Once()

	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version1", "somefile1", "a883dafc480d466ee04e0d6da986bd78eb1fdd2178d04693723da3a8f95d42f4", uint64(5)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version1", "somefile2", "ff5a972ba33179c7ec67c73e00a362b629c489f9d7c86489644db2bcd8c62c61", uint64(5)).
		Return(ptr.String(""), nil).
		Once()

	fn := s.svc.CreateObject(defaultNamespace, "container1", "version1", "testdata/repo/")
	s.Require().NoError(fn(s.ctx))
}

func (s *serviceTestSuite) TestCreateObjectWithCache() {
	s.cacheMock.On("Get", "testdata/repo/somefile1").Return("a883dafc480d466ee04e0d6da986bd78eb1fdd2178d04693723da3a8f95d42f4", nil).Once()
	s.cacheMock.On("Get", "testdata/repo/somefile2").Return("ff5a972ba33179c7ec67c73e00a362b629c489f9d7c86489644db2bcd8c62c61", nil).Once()

	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version1", "somefile1", "a883dafc480d466ee04e0d6da986bd78eb1fdd2178d04693723da3a8f95d42f4", uint64(5)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version1", "somefile2", "ff5a972ba33179c7ec67c73e00a362b629c489f9d7c86489644db2bcd8c62c61", uint64(5)).
		Return(ptr.String(""), nil).
		Once()

	fn := s.svc.CreateObject(defaultNamespace, "container1", "version1", "testdata/repo/")
	s.Require().NoError(fn(s.ctx))
}

func (s *serviceTestSuite) TestCreateObjectWithUploadURL() {
	s.cacheMock.On("Get", "testdata/repo/somefile1").Return("a883dafc480d466ee04e0d6da986bd78eb1fdd2178d04693723da3a8f95d42f4", nil).Once()
	s.cacheMock.On("Get", "testdata/repo/somefile2").Return("ff5a972ba33179c7ec67c73e00a362b629c489f9d7c86489644db2bcd8c62c61", nil).Once()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, err := io.ReadAll(r.Body)
		s.NoError(err)
		defer r.Body.Close()

		s.Equal("1234\n", string(data))

		s.Equal("/test-url", r.RequestURI)
		s.Equal(mimeTypeMultipartFormData, r.Header.Get("Content-Type"))
	}))
	defer srv.Close()

	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version1", "somefile1", "a883dafc480d466ee04e0d6da986bd78eb1fdd2178d04693723da3a8f95d42f4", uint64(5)).
		Return(ptr.String(srv.URL+"/test-url"), nil).
		Once()
	s.cliMock.
		On("CreateObject", defaultNamespace, "container1", "version1", "somefile2", "ff5a972ba33179c7ec67c73e00a362b629c489f9d7c86489644db2bcd8c62c61", uint64(5)).
		Return(ptr.String(""), nil).
		Once()

	fn := s.svc.CreateObject(defaultNamespace, "container1", "version1", "testdata/repo/")
	s.Require().NoError(fn(s.ctx))
}

func (s *serviceTestSuite) TestDeleteObject() {
	s.cliMock.On("DeleteObject", defaultNamespace, "container1", "version1", "key1").Return(nil).Once()

	fn := s.svc.DeleteObject(defaultNamespace, "container1", "version1", "key1")
	s.Require().NoError(fn(s.ctx))
}

func (s *serviceTestSuite) TestListObjects() {
	s.cliMock.On("ListObjects", defaultNamespace, "container1", "version1").Return([]string{"obj1", "obj2", "obj3"}, nil).Once()

	fn := s.svc.ListObjects(defaultNamespace, "container1", "version1")
	s.Require().NoError(fn(s.ctx))
}

func (s *serviceTestSuite) TestGetObjectURL() {
	s.cliMock.On("GetObjectURL", defaultNamespace, "container1", "version1", "key1").Return("https://example.com", nil).Once()

	fn := s.svc.GetObjectURL(defaultNamespace, "container1", "version1", "key1")
	s.Require().NoError(fn(s.ctx))
}

// Definitions ...
type serviceTestSuite struct {
	suite.Suite

	ctx        context.Context
	cliMock    *protoClientMock
	cacheMock  *cacheMock.Mock
	svc        Service
	sourceMock *sourceMock.Mock
}

func (s *serviceTestSuite) SetupTest() {
	s.ctx = context.Background()

	s.cliMock = newMock()
	s.cacheMock = cacheMock.New()
	s.sourceMock = sourceMock.New()

	s.svc = New(s.cliMock, s.cacheMock)
}

func (s *serviceTestSuite) TearDownTest() {
	s.cliMock.AssertExpectations(s.T())
	s.cacheMock.AssertExpectations(s.T())
}

func TestServiceTestSuite(t *testing.T) {
	suite.Run(t, &serviceTestSuite{})
}
