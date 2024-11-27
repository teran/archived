package grpc

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	grpctest "github.com/teran/go-grpctest"
	ptr "github.com/teran/go-ptr"

	v1pb "github.com/teran/archived/manager/presenter/grpc/proto/v1"
	"github.com/teran/archived/models"
	"github.com/teran/archived/service"
)

const defaultNamespace = "default"

func (s *manageHandlersTestSuite) TestCreateNamespace() {
	s.svcMock.On("CreateNamespace", "test-namespace").Return(nil).Once()

	_, err := s.client.CreateNamespace(s.ctx, &v1pb.CreateNamespaceRequest{
		Name: "test-namespace",
	})
	s.Require().NoError(err)
}

func (s *manageHandlersTestSuite) TestListNamespaces() {
	s.svcMock.On("ListNamespaces").Return([]string{"test-namespace1", "test-namespace2"}, nil).Once()

	resp, err := s.client.ListNamespaces(s.ctx, &v1pb.ListNamespacesRequest{})
	s.Require().NoError(err)
	s.Require().Equal([]string{
		"test-namespace1", "test-namespace2",
	}, resp.GetName())
}

func (s *manageHandlersTestSuite) TestRenameNamespaces() {
	s.svcMock.On("RenameNamespace", "old-name", "new-name").Return(nil).Once()

	_, err := s.client.RenameNamespace(s.ctx, &v1pb.RenameNamespaceRequest{
		OldName: "old-name",
		NewName: "new-name",
	})
	s.Require().NoError(err)
}

func (s *manageHandlersTestSuite) TestDeleteNamespace() {
	s.svcMock.On("DeleteNamespace", "test-namespace").Return(nil).Once()

	_, err := s.client.DeleteNamespace(s.ctx, &v1pb.DeleteNamespaceRequest{
		Name: "test-namespace",
	})
	s.Require().NoError(err)
}

func (s *manageHandlersTestSuite) TestCreateContainer() {
	s.svcMock.On("CreateContainer", defaultNamespace, "test-container", 364*time.Second).Return(nil).Once()

	_, err := s.client.CreateContainer(s.ctx, &v1pb.CreateContainerRequest{
		Namespace:  defaultNamespace,
		Name:       "test-container",
		TtlSeconds: ptr.Int64(364),
	})
	s.Require().NoError(err)
}

func (s *manageHandlersTestSuite) TestMoveContainer() {
	s.svcMock.On("MoveContainer", defaultNamespace, "test-container", "new-namespace").Return(nil).Once()

	_, err := s.client.MoveContainer(s.ctx, &v1pb.MoveContainerRequest{
		Namespace:            defaultNamespace,
		ContainerName:        "test-container",
		DestinationNamespace: "new-namespace",
	})
	s.Require().NoError(err)
}

func (s *manageHandlersTestSuite) TestCreateContainerNotFound() {
	s.svcMock.On("CreateContainer", defaultNamespace, "test-container", 3*time.Second).Return(service.ErrNotFound).Once()

	_, err := s.client.CreateContainer(s.ctx, &v1pb.CreateContainerRequest{
		Namespace:  defaultNamespace,
		Name:       "test-container",
		TtlSeconds: ptr.Int64(3),
	})
	s.Require().Error(err)
	s.Require().Equal("rpc error: code = NotFound desc = entity not found", err.Error())
}

func (s *manageHandlersTestSuite) TestRenameContainer() {
	s.svcMock.On("RenameContainer", defaultNamespace, "old-name", "new-name").Return(nil).Once()

	_, err := s.client.RenameContainer(s.ctx, &v1pb.RenameContainerRequest{
		Namespace: defaultNamespace,
		OldName:   "old-name",
		NewName:   "new-name",
	})
	s.Require().NoError(err)
}

func (s *manageHandlersTestSuite) TestRenameContainerNotFound() {
	s.svcMock.On("RenameContainer", defaultNamespace, "old-name", "new-name").Return(service.ErrNotFound).Once()

	_, err := s.client.RenameContainer(s.ctx, &v1pb.RenameContainerRequest{
		Namespace: defaultNamespace,
		OldName:   "old-name",
		NewName:   "new-name",
	})
	s.Require().Error(err)
	s.Require().Equal("rpc error: code = NotFound desc = entity not found", err.Error())
}

func (s *manageHandlersTestSuite) TestSetContainerParameters() {
	s.svcMock.On("SetContainerParameters", defaultNamespace, "test-container", 1*time.Hour).Return(nil).Once()

	_, err := s.client.SetContainerParameters(s.ctx, &v1pb.SetContainerParametersRequest{
		Namespace:  defaultNamespace,
		Name:       "test-container",
		TtlSeconds: ptr.Int64(3600),
	})
	s.Require().NoError(err)
}

func (s *manageHandlersTestSuite) TestListContainers() {
	s.svcMock.On("ListContainers", defaultNamespace).Return([]models.Container{
		{Name: "test-container1"},
		{Name: "test-container2"},
	}, nil).Once()

	resp, err := s.client.ListContainers(s.ctx, &v1pb.ListContainersRequest{
		Namespace: defaultNamespace,
	})
	s.Require().NoError(err)
	s.Require().Equal([]string{
		"test-container1", "test-container2",
	}, resp.GetName())
}

func (s *manageHandlersTestSuite) TestListContainersNotFound() {
	s.svcMock.On("ListContainers", defaultNamespace).Return([]models.Container{}, service.ErrNotFound).Once()

	_, err := s.client.ListContainers(s.ctx, &v1pb.ListContainersRequest{
		Namespace: defaultNamespace,
	})
	s.Require().Error(err)
	s.Require().Equal("rpc error: code = NotFound desc = entity not found", err.Error())
}

func (s *manageHandlersTestSuite) TestCreateVersion() {
	s.svcMock.On("CreateVersion", defaultNamespace, "test-container").Return("20240102030405", nil).Once()

	resp, err := s.client.CreateVersion(s.ctx, &v1pb.CreateVersionRequest{
		Namespace: defaultNamespace,
		Container: "test-container",
	})
	s.Require().NoError(err)
	s.Require().Equal("20240102030405", resp.GetVersion())
}

func (s *manageHandlersTestSuite) TestCreateVersionNotFound() {
	s.svcMock.On("CreateVersion", defaultNamespace, "test-container").Return("", service.ErrNotFound).Once()

	_, err := s.client.CreateVersion(s.ctx, &v1pb.CreateVersionRequest{
		Namespace: defaultNamespace,
		Container: "test-container",
	})
	s.Require().Error(err)
	s.Require().Equal("rpc error: code = NotFound desc = entity not found", err.Error())
}

func (s *manageHandlersTestSuite) TestDeleteContainer() {
	s.svcMock.On("DeleteContainer", defaultNamespace, "test-container").Return(nil).Once()

	_, err := s.client.DeleteContainer(s.ctx, &v1pb.DeleteContainerRequest{
		Namespace: defaultNamespace,
		Name:      "test-container",
	})
	s.Require().NoError(err)
}

func (s *manageHandlersTestSuite) TestDeleteContainerNotFound() {
	s.svcMock.On("DeleteContainer", defaultNamespace, "test-container").Return(service.ErrNotFound).Once()

	_, err := s.client.DeleteContainer(s.ctx, &v1pb.DeleteContainerRequest{
		Namespace: defaultNamespace,
		Name:      "test-container",
	})
	s.Require().Error(err)
	s.Require().Equal("rpc error: code = NotFound desc = entity not found", err.Error())
}

func (s *manageHandlersTestSuite) TestListVersions() {
	s.svcMock.On("ListAllVersions", defaultNamespace, "test-container").Return([]models.Version{
		{
			Name: "version1",
		},
		{
			Name: "version2",
		},
	}, nil).Once()

	resp, err := s.client.ListVersions(s.ctx, &v1pb.ListVersionsRequest{
		Namespace: defaultNamespace,
		Container: "test-container",
	})
	s.Require().NoError(err)
	s.Require().Equal([]string{
		"version1", "version2",
	}, resp.GetVersions())
}

func (s *manageHandlersTestSuite) TestListVersionsNotFound() {
	s.svcMock.On("ListAllVersions", defaultNamespace, "test-container").Return([]models.Version{}, service.ErrNotFound).Once()

	_, err := s.client.ListVersions(s.ctx, &v1pb.ListVersionsRequest{
		Namespace: defaultNamespace,
		Container: "test-container",
	})
	s.Require().Error(err)
	s.Require().Equal("rpc error: code = NotFound desc = entity not found", err.Error())
}

func (s *manageHandlersTestSuite) TestDeleteVersion() {
	s.svcMock.On("DeleteVersion", defaultNamespace, "test-container", "test-version").Return(nil).Once()

	_, err := s.client.DeleteVersion(s.ctx, &v1pb.DeleteVersionRequest{
		Namespace: defaultNamespace,
		Container: "test-container",
		Version:   "test-version",
	})
	s.Require().NoError(err)
}

func (s *manageHandlersTestSuite) TestDeleteVersionNotFound() {
	s.svcMock.On("DeleteVersion", defaultNamespace, "test-container", "test-version").Return(service.ErrNotFound).Once()

	_, err := s.client.DeleteVersion(s.ctx, &v1pb.DeleteVersionRequest{
		Namespace: defaultNamespace,
		Container: "test-container",
		Version:   "test-version",
	})
	s.Require().Error(err)
	s.Require().Equal("rpc error: code = NotFound desc = entity not found", err.Error())
}

func (s *manageHandlersTestSuite) TestPublishVersion() {
	s.svcMock.On("PublishVersion", defaultNamespace, "test-container", "20240102030405").Return(nil).Once()

	_, err := s.client.PublishVersion(s.ctx, &v1pb.PublishVersionRequest{
		Namespace: defaultNamespace,
		Container: "test-container",
		Version:   "20240102030405",
	})
	s.Require().NoError(err)
}

func (s *manageHandlersTestSuite) TestPublishVersionNotFound() {
	s.svcMock.On("PublishVersion", defaultNamespace, "test-container", "20240102030405").Return(service.ErrNotFound).Once()

	_, err := s.client.PublishVersion(s.ctx, &v1pb.PublishVersionRequest{
		Namespace: defaultNamespace,
		Container: "test-container",
		Version:   "20240102030405",
	})
	s.Require().Error(err)
	s.Require().Equal("rpc error: code = NotFound desc = entity not found", err.Error())
}

func (s *manageHandlersTestSuite) TestCreateObject() {
	s.svcMock.On("EnsureBLOBPresenceOrGetUploadURL", "checksum", uint64(1234), "application/x-rpm").Return("https://example.com/url", nil).Once()
	s.svcMock.On("AddObject", defaultNamespace, "test-container", "version", "key", "checksum").Return(nil).Once()

	resp, err := s.client.CreateObject(s.ctx, &v1pb.CreateObjectRequest{
		Namespace: defaultNamespace,
		Container: "test-container",
		Version:   "version",
		Key:       "key",
		Checksum:  "checksum",
		Size:      uint64(1234),
		MimeType:  "application/x-rpm",
	})
	s.Require().NoError(err)
	s.Require().Equal("https://example.com/url", resp.GetUploadUrl())
}

func (s *manageHandlersTestSuite) TestCreateObjectNotFound() {
	s.svcMock.On("EnsureBLOBPresenceOrGetUploadURL", "checksum", uint64(1234), "application/x-rpm").Return("", service.ErrNotFound).Once()

	_, err := s.client.CreateObject(s.ctx, &v1pb.CreateObjectRequest{
		Namespace: defaultNamespace,
		Container: "test-container",
		Version:   "version",
		Key:       "key",
		Checksum:  "checksum",
		Size:      uint64(1234),
		MimeType:  "application/x-rpm",
	})
	s.Require().Error(err)
	s.Require().Equal("rpc error: code = NotFound desc = entity not found", err.Error())
}

func (s *manageHandlersTestSuite) TestListObjects() {
	s.svcMock.On("ListObjects", defaultNamespace, "container", "version").Return([]string{"obj1", "obj2", "obj3"}, nil).Once()

	resp, err := s.client.ListObjects(s.ctx, &v1pb.ListObjectsRequest{
		Namespace: defaultNamespace,
		Container: "container",
		Version:   "version",
	})
	s.Require().NoError(err)
	s.Require().Equal([]string{
		"obj1", "obj2", "obj3",
	}, resp.GetObjects())
}

func (s *manageHandlersTestSuite) TestListObjectsNotFound() {
	s.svcMock.On("ListObjects", defaultNamespace, "container", "version").Return([]string{}, service.ErrNotFound).Once()

	_, err := s.client.ListObjects(s.ctx, &v1pb.ListObjectsRequest{
		Namespace: defaultNamespace,
		Container: "container",
		Version:   "version",
	})
	s.Require().Error(err)
	s.Require().Equal("rpc error: code = NotFound desc = entity not found", err.Error())
}

func (s *manageHandlersTestSuite) TestGetObjectURL() {
	s.svcMock.On("GetObjectURL", defaultNamespace, "test-container", "test-version", "test-key").Return("test-url", nil).Once()

	resp, err := s.client.GetObjectURL(s.ctx, &v1pb.GetObjectURLRequest{
		Namespace: defaultNamespace,
		Container: "test-container",
		Version:   "test-version",
		Key:       "test-key",
	})
	s.Require().NoError(err)
	s.Require().Equal("test-url", resp.GetUrl())
}

func (s *manageHandlersTestSuite) TestGetObjectURLNotFound() {
	s.svcMock.On("GetObjectURL", defaultNamespace, "test-container", "test-version", "test-key").Return("test-url", service.ErrNotFound).Once()

	_, err := s.client.GetObjectURL(s.ctx, &v1pb.GetObjectURLRequest{
		Namespace: defaultNamespace,
		Container: "test-container",
		Version:   "test-version",
		Key:       "test-key",
	})
	s.Require().Error(err)
	s.Require().Equal("rpc error: code = NotFound desc = entity not found", err.Error())
}

func (s *manageHandlersTestSuite) TestDeleteObject() {
	s.svcMock.On("DeleteObject", defaultNamespace, "test-container", "test-version", "test-key").Return(nil).Once()

	_, err := s.client.DeleteObject(s.ctx, &v1pb.DeleteObjectRequest{
		Namespace: defaultNamespace,
		Container: "test-container",
		Version:   "test-version",
		Key:       "test-key",
	})
	s.Require().NoError(err)
}

func (s *manageHandlersTestSuite) TestDeleteObjectNotFound() {
	s.svcMock.On("DeleteObject", defaultNamespace, "test-container", "test-version", "test-key").Return(service.ErrNotFound).Once()

	_, err := s.client.DeleteObject(s.ctx, &v1pb.DeleteObjectRequest{
		Namespace: defaultNamespace,
		Container: "test-container",
		Version:   "test-version",
		Key:       "test-key",
	})
	s.Require().Error(err)
	s.Require().Equal("rpc error: code = NotFound desc = entity not found", err.Error())
}

// Definitions ...
type manageHandlersTestSuite struct {
	suite.Suite

	ctx    context.Context
	cancel context.CancelFunc

	svcMock  *service.Mock
	handlers ManageServerInterface
	srv      grpctest.Server

	client v1pb.ManageServiceClient
}

func (s *manageHandlersTestSuite) SetupTest() {
	s.svcMock = service.NewMock()
	s.handlers = New(s.svcMock)
	s.srv = grpctest.New()
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
