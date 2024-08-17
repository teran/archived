package grpc

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	v1 "github.com/teran/archived/manager/presenter/grpc/proto/v1"
	"github.com/teran/archived/service"
	ptr "github.com/teran/go-ptr"
)

var _ v1.ManageServiceServer = (*handlers)(nil)

type ManageServerInterface interface {
	v1.ManageServiceServer

	Register(*grpc.Server)
}

type handlers struct {
	v1.UnimplementedManageServiceServer

	svc service.Manager
}

func New(svc service.Manager) ManageServerInterface {
	return &handlers{
		svc: svc,
	}
}

func (h *handlers) CreateContainer(ctx context.Context, in *v1.CreateContainerRequest) (*v1.CreateContainerResponse, error) {
	err := h.svc.CreateContainer(ctx, in.GetName())
	if err != nil {
		return nil, err
	}

	return &v1.CreateContainerResponse{}, nil
}

func (h *handlers) DeleteContainer(ctx context.Context, in *v1.DeleteContainerRequest) (*v1.DeleteContainerResponse, error) {
	err := h.svc.DeleteContainer(ctx, in.GetName())
	if err != nil {
		return nil, err
	}

	return &v1.DeleteContainerResponse{}, nil
}

func (h *handlers) ListContainers(ctx context.Context, _ *v1.ListContainersRequest) (*v1.ListContainersResponse, error) {
	containers, err := h.svc.ListContainers(ctx)
	if err != nil {
		return nil, err
	}

	return &v1.ListContainersResponse{
		Name: containers,
	}, nil
}

func (h *handlers) CreateVersion(ctx context.Context, in *v1.CreateVersionRequest) (*v1.CreateVersionResponse, error) {
	version, err := h.svc.CreateVersion(ctx, in.GetContainer())
	if err != nil {
		return nil, err
	}

	return &v1.CreateVersionResponse{
		Version: version,
	}, nil
}

func (h *handlers) ListVersions(ctx context.Context, in *v1.ListVersionsRequest) (*v1.ListVersionsResponse, error) {
	versions, err := h.svc.ListAllVersions(ctx, in.GetContainer())
	if err != nil {
		if err == service.ErrNotFound {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, err
	}

	versionNames := []string{}
	for _, v := range versions {
		versionNames = append(versionNames, v.Name)
	}

	return &v1.ListVersionsResponse{
		Versions: versionNames,
	}, nil
}

func (h *handlers) DeleteVersion(ctx context.Context, in *v1.DeleteVersionRequest) (*v1.DeleteVersionResponse, error) {
	err := h.svc.DeleteVersion(ctx, in.GetContainer(), in.GetVersion())
	if err != nil {
		return nil, err
	}

	return &v1.DeleteVersionResponse{}, nil
}

func (h *handlers) PublishVersion(ctx context.Context, in *v1.PublishVersionRequest) (*v1.PublishVersionResponse, error) {
	err := h.svc.PublishVersion(ctx, in.GetContainer(), in.GetVersion())
	if err != nil {
		return nil, err
	}

	return &v1.PublishVersionResponse{}, nil
}

func (h *handlers) CreateObject(ctx context.Context, in *v1.CreateObjectRequest) (*v1.CreateObjectResponse, error) {
	url, err := h.svc.EnsureBLOBPresenceOrGetUploadURL(ctx, in.GetChecksum(), in.GetSize())
	if err != nil && url == "" {
		return nil, err
	}

	err = h.svc.AddObject(ctx, in.GetContainer(), in.GetVersion(), in.GetKey(), in.GetChecksum())
	if err != nil {
		return nil, err
	}

	return &v1.CreateObjectResponse{
		UploadUrl: ptr.String(url),
	}, nil
}

func (h *handlers) ListObjects(ctx context.Context, in *v1.ListObjectsRequest) (*v1.ListObjectsResponse, error) {
	objects, err := h.svc.ListObjects(ctx, in.GetContainer(), in.GetVersion())
	if err != nil {
		if err == service.ErrNotFound {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, err
	}

	return &v1.ListObjectsResponse{
		Objects: objects,
	}, nil
}

func (h *handlers) GetObjectURL(ctx context.Context, in *v1.GetObjectURLRequest) (*v1.GetObjectURLResponse, error) {
	url, err := h.svc.GetObjectURL(ctx, in.GetContainer(), in.GetVersion(), in.GetKey())
	if err != nil {
		if err == service.ErrNotFound {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, err
	}

	return &v1.GetObjectURLResponse{
		Url: url,
	}, nil
}

func (h *handlers) DeleteObject(ctx context.Context, in *v1.DeleteObjectRequest) (*v1.DeleteObjectResponse, error) {
	err := h.svc.DeleteObject(ctx, in.GetContainer(), in.GetVersion(), in.GetKey())
	if err != nil {
		return nil, err
	}

	return &v1.DeleteObjectResponse{}, nil
}

func (h *handlers) Register(gs *grpc.Server) {
	v1.RegisterManageServiceServer(gs, h)
}
