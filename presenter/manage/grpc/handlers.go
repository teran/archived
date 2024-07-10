package grpc

import (
	"context"

	"github.com/teran/archived/presenter/manage/grpc/proto"
	"google.golang.org/grpc"
)

var _ proto.ManageServer = (*handlers)(nil)

type ManageServerInterface interface {
	proto.ManageServer

	Register(*grpc.Server)
}

type handlers struct {
	proto.UnimplementedManageServer
}

func New() ManageServerInterface {
	return &handlers{}
}

func (h *handlers) CreateContainer(context.Context, *proto.CreateContainerRequest) (*proto.CreateContainerResponse, error) {
	panic("not implemented")
}

func (h *handlers) DeleteContainer(context.Context, *proto.DeleteContainerRequest) (*proto.DeleteContainerResponse, error) {
	panic("not implemented")
}

func (h *handlers) ListContainers(context.Context, *proto.ListContainersRequest) (*proto.ListContainersResponse, error) {
	panic("not implemented")
}

func (h *handlers) ListVersions(context.Context, *proto.ListVersionsRequest) (*proto.ListVersionsResponse, error) {
	panic("not implemented")
}

func (h *handlers) DeleteVersion(context.Context, *proto.DeleteVersionRequest) (*proto.DeleteVersionResponse, error) {
	panic("not implemented")
}

func (h *handlers) PublishVersion(context.Context, *proto.PublishVersionRequest) (*proto.PublishVersionResponse, error) {
	panic("not implemented")
}

func (h *handlers) ListObjects(context.Context, *proto.ListObjectsRequest) (*proto.ListObjectsResponse, error) {
	panic("not implemented")
}

func (h *handlers) GetObjectURL(context.Context, *proto.GetObjectURLRequest) (*proto.GetObjectURLResponse, error) {
	panic("not implemented")
}

func (h *handlers) DeleteObject(context.Context, *proto.DeleteObjectRequest) (*proto.DeleteObjectResponse, error) {
	panic("not implemented")
}

func (h *handlers) Register(gs *grpc.Server) {
	proto.RegisterManageServer(gs, h)
}
