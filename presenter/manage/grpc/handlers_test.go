package grpc

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	v1pb "github.com/teran/archived/presenter/manage/grpc/proto/v1"
	"github.com/teran/archived/service"
)

func (s *manageHandlersTestSuite) TestCreateContainer() {
	s.svcMock.On("CreateContainer", "test-container").Return(nil).Once()

	_, err := s.client.CreateContainer(s.ctx, &v1pb.CreateContainerRequest{
		Name: "test-container",
	})
	s.Require().NoError(err)
}

func (s *manageHandlersTestSuite) TestListContainers() {
	s.svcMock.On("ListContainers").Return([]string{"test-container1", "test-container2"}, nil).Once()

	resp, err := s.client.ListContainers(s.ctx, &v1pb.ListContainersRequest{})
	s.Require().NoError(err)
	s.Require().Equal([]string{
		"test-container1", "test-container2",
	}, resp.GetName())
}

func (s *manageHandlersTestSuite) TestCreateVersion() {
	s.svcMock.On("CreateVersion", "test-container").Return("20240102030405", nil).Once()

	resp, err := s.client.CreateVersion(s.ctx, &v1pb.CreateVersionRequest{
		Container: "test-container",
	})
	s.Require().NoError(err)
	s.Require().Equal("20240102030405", resp.GetVersion())
}

func (s *manageHandlersTestSuite) TestPublishVersion() {
	s.svcMock.On("PublishVersion", "test-container", "20240102030405").Return(nil).Once()

	_, err := s.client.PublishVersion(s.ctx, &v1pb.PublishVersionRequest{
		Container: "test-container",
		Version:   "20240102030405",
	})
	s.Require().NoError(err)
}

func (s *manageHandlersTestSuite) TestCreateObject() {
	s.svcMock.On("EnsureBLOBPresenceOrGetUploadURL", "checksum", int64(1234)).Return("https://example.com/url", nil).Once()
	s.svcMock.On("AddObject", "test-container", "version", "key", "checksum").Return(nil).Once()

	resp, err := s.client.CreateObject(s.ctx, &v1pb.CreateObjectRequest{
		Container: "test-container",
		Version:   "version",
		Key:       "key",
		Checksum:  "checksum",
		Size:      1234,
	})
	s.Require().NoError(err)
	s.Require().Equal("https://example.com/url", resp.GetUploadUrl())
}

func (s *manageHandlersTestSuite) TestListObjects() {
	s.svcMock.On("ListObjects", "container", "version").Return([]string{"obj1", "obj2", "obj3"}, nil).Once()

	resp, err := s.client.ListObjects(s.ctx, &v1pb.ListObjectsRequest{
		Container: "container",
		Version:   "version",
	})
	s.Require().NoError(err)
	s.Require().Equal([]string{
		"obj1", "obj2", "obj3",
	}, resp.GetObjects())
}

// Definitions ...
type manageHandlersTestSuite struct {
	suite.Suite

	ctx    context.Context
	cancel context.CancelFunc

	svcMock  *service.Mock
	handlers ManageServerInterface
	srv      testServer

	client v1pb.ManageServiceClient
}

func (s *manageHandlersTestSuite) SetupTest() {
	s.svcMock = service.NewMock()
	s.handlers = New(s.svcMock)
	s.srv = newTestServer()
	s.handlers.Register(s.srv.Server())

	err := s.srv.Run()
	s.Require().NoError(err)

	s.ctx, s.cancel = context.WithTimeout(context.Background(), 10*time.Second)

	dial, err := s.srv.DialContext(s.ctx)
	s.Require().NoError(err)

	s.client = v1pb.NewManageServiceClient(dial)
}

func (s *manageHandlersTestSuite) TearDownTest() {
	s.svcMock.AssertExpectations(s.T())
	s.srv.Close()
	s.cancel()
}

func TestManageHandlersTestSuite(t *testing.T) {
	suite.Run(t, &manageHandlersTestSuite{})
}
