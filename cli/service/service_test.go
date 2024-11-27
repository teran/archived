package service

import (
	"context"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"

	sourceMock "github.com/teran/archived/cli/service/source/mock"
	cacheMock "github.com/teran/archived/cli/service/stat_cache/mock"
)

const (
	defaultNamespace = "default"
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

	fn := s.svc.CreateContainer(defaultNamespace, "test-container", -1)
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

func (s *serviceTestSuite) TestSetContainerParameters() {
	s.cliMock.On("SetContainerParameters", defaultNamespace, "test-container1", 3600*time.Second).Return(nil).Once()

	fn := s.svc.SetContainerParameters(defaultNamespace, "test-container1", 3600*time.Second)
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
