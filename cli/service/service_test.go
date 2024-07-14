package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/teran/go-ptr"
)

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

func (s *serviceTestSuite) TestCreateVersion() {
	s.cliMock.On("CreateVersion", "container1").Return("version_id", nil).Once()

	fn := s.svc.CreateVersion("container1")
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

func (s *serviceTestSuite) TestCreateObject() {
	s.cliMock.
		On("CreateObject", "container1", "version1", "/somefile1", "a883dafc480d466ee04e0d6da986bd78eb1fdd2178d04693723da3a8f95d42f4", int64(5)).
		Return(ptr.String(""), nil).
		Once()
	s.cliMock.
		On("CreateObject", "container1", "version1", "/somefile2", "ff5a972ba33179c7ec67c73e00a362b629c489f9d7c86489644db2bcd8c62c61", int64(5)).
		Return(ptr.String(""), nil).
		Once()

	fn := s.svc.CreateObject("container1", "version1", "testdata")
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

	ctx     context.Context
	cliMock *protoClientMock
	svc     Service
}

func (s *serviceTestSuite) SetupTest() {
	s.ctx = context.Background()
	s.cliMock = newMock()
	s.svc = New(s.cliMock)
}

func (s *serviceTestSuite) TearDownTest() {
	s.cliMock.AssertExpectations(s.T())
}

func TestServiceTestSuite(t *testing.T) {
	suite.Run(t, &serviceTestSuite{})
}
