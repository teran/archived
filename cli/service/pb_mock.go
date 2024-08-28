package service

import (
	"context"

	v1proto "github.com/teran/archived/manager/presenter/grpc/proto/v1"
	"github.com/teran/archived/repositories/blob/mock"
	"google.golang.org/grpc"
)

var _ v1proto.ManageServiceClient = (*protoClientMock)(nil)

type protoClientMock struct {
	mock.Mock
}

func newMock() *protoClientMock {
	return &protoClientMock{}
}

func (m *protoClientMock) CreateNamespace(ctx context.Context, in *v1proto.CreateNamespaceRequest, opts ...grpc.CallOption) (*v1proto.CreateNamespaceResponse, error) {
	args := m.Called(in.GetName())
	return &v1proto.CreateNamespaceResponse{}, args.Error(0)
}

func (m *protoClientMock) RenameNamespace(ctx context.Context, in *v1proto.RenameNamespaceRequest, opts ...grpc.CallOption) (*v1proto.RenameNamespaceResponse, error) {
	args := m.Called(in.GetOldName(), in.GetNewName())
	return &v1proto.RenameNamespaceResponse{}, args.Error(0)
}

func (m *protoClientMock) DeleteNamespace(ctx context.Context, in *v1proto.DeleteNamespaceRequest, opts ...grpc.CallOption) (*v1proto.DeleteNamespaceResponse, error) {
	args := m.Called(in.GetName())
	return &v1proto.DeleteNamespaceResponse{}, args.Error(0)
}

func (m *protoClientMock) ListNamespaces(ctx context.Context, in *v1proto.ListNamespacesRequest, opts ...grpc.CallOption) (*v1proto.ListNamespacesResponse, error) {
	args := m.Called()
	return &v1proto.ListNamespacesResponse{
		Name: args.Get(0).([]string),
	}, args.Error(1)
}

func (m *protoClientMock) CreateContainer(ctx context.Context, in *v1proto.CreateContainerRequest, opts ...grpc.CallOption) (*v1proto.CreateContainerResponse, error) {
	args := m.Called(in.GetNamespace(), in.GetName())
	return &v1proto.CreateContainerResponse{}, args.Error(0)
}

func (m *protoClientMock) MoveContainer(ctx context.Context, in *v1proto.MoveContainerRequest, opts ...grpc.CallOption) (*v1proto.MoveContainerResponse, error) {
	args := m.Called(in.GetNamespace(), in.GetContainerName(), in.GetDestinationNamespace())
	return &v1proto.MoveContainerResponse{}, args.Error(0)
}

func (m *protoClientMock) RenameContainer(ctx context.Context, in *v1proto.RenameContainerRequest, opts ...grpc.CallOption) (*v1proto.RenameContainerResponse, error) {
	args := m.Called(in.GetNamespace(), in.GetOldName(), in.GetNewName())
	return &v1proto.RenameContainerResponse{}, args.Error(0)
}

func (m *protoClientMock) DeleteContainer(ctx context.Context, in *v1proto.DeleteContainerRequest, opts ...grpc.CallOption) (*v1proto.DeleteContainerResponse, error) {
	args := m.Called(in.GetNamespace(), in.GetName())
	return &v1proto.DeleteContainerResponse{}, args.Error(0)
}

func (m *protoClientMock) ListContainers(ctx context.Context, in *v1proto.ListContainersRequest, opts ...grpc.CallOption) (*v1proto.ListContainersResponse, error) {
	args := m.Called(in.GetNamespace())
	return &v1proto.ListContainersResponse{
		Name: args.Get(0).([]string),
	}, args.Error(1)
}

func (m *protoClientMock) CreateVersion(ctx context.Context, in *v1proto.CreateVersionRequest, opts ...grpc.CallOption) (*v1proto.CreateVersionResponse, error) {
	args := m.Called(in.GetNamespace(), in.GetContainer())
	return &v1proto.CreateVersionResponse{
		Version: args.String(0),
	}, args.Error(1)
}

func (m *protoClientMock) ListVersions(ctx context.Context, in *v1proto.ListVersionsRequest, opts ...grpc.CallOption) (*v1proto.ListVersionsResponse, error) {
	args := m.Called(in.GetNamespace(), in.GetContainer())
	return &v1proto.ListVersionsResponse{
		Versions: args.Get(0).([]string),
	}, args.Error(1)
}

func (m *protoClientMock) DeleteVersion(ctx context.Context, in *v1proto.DeleteVersionRequest, opts ...grpc.CallOption) (*v1proto.DeleteVersionResponse, error) {
	args := m.Called(in.GetNamespace(), in.GetContainer(), in.GetVersion())
	return &v1proto.DeleteVersionResponse{}, args.Error(0)
}

func (m *protoClientMock) PublishVersion(ctx context.Context, in *v1proto.PublishVersionRequest, opts ...grpc.CallOption) (*v1proto.PublishVersionResponse, error) {
	args := m.Called(in.GetNamespace(), in.GetContainer(), in.GetVersion())
	return &v1proto.PublishVersionResponse{}, args.Error(0)
}

func (m *protoClientMock) CreateObject(ctx context.Context, in *v1proto.CreateObjectRequest, opts ...grpc.CallOption) (*v1proto.CreateObjectResponse, error) {
	args := m.Called(in.GetNamespace(), in.GetContainer(), in.GetVersion(), in.GetKey(), in.GetChecksum(), in.GetSize())
	return &v1proto.CreateObjectResponse{
		UploadUrl: args.Get(0).(*string),
	}, args.Error(1)
}

func (m *protoClientMock) ListObjects(ctx context.Context, in *v1proto.ListObjectsRequest, opts ...grpc.CallOption) (*v1proto.ListObjectsResponse, error) {
	args := m.Called(in.GetNamespace(), in.GetContainer(), in.GetVersion())
	return &v1proto.ListObjectsResponse{
		Objects: args.Get(0).([]string),
	}, args.Error(1)
}

func (m *protoClientMock) GetObjectURL(ctx context.Context, in *v1proto.GetObjectURLRequest, opts ...grpc.CallOption) (*v1proto.GetObjectURLResponse, error) {
	args := m.Called(in.GetNamespace(), in.GetContainer(), in.GetVersion(), in.GetKey())
	return &v1proto.GetObjectURLResponse{
		Url: args.String(0),
	}, args.Error(1)
}

func (m *protoClientMock) DeleteObject(_ context.Context, in *v1proto.DeleteObjectRequest, opts ...grpc.CallOption) (*v1proto.DeleteObjectResponse, error) {
	args := m.Called(in.GetNamespace(), in.GetContainer(), in.GetVersion(), in.GetKey())
	return &v1proto.DeleteObjectResponse{}, args.Error(0)
}
