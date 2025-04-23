package grpc

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	v1 "github.com/teran/archived/manager/presenter/grpc/proto/v1"
	"github.com/teran/archived/service"
	"github.com/teran/go-collection/types/ptr"
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

func (h *handlers) CreateNamespace(ctx context.Context, in *v1.CreateNamespaceRequest) (*v1.CreateNamespaceResponse, error) {
	err := h.svc.CreateNamespace(ctx, in.GetName())
	if err != nil {
		return nil, mapServiceError(err)
	}

	return &v1.CreateNamespaceResponse{}, nil
}

func (h *handlers) RenameNamespace(ctx context.Context, in *v1.RenameNamespaceRequest) (*v1.RenameNamespaceResponse, error) {
	err := h.svc.RenameNamespace(ctx, in.GetOldName(), in.GetNewName())
	if err != nil {
		return nil, mapServiceError(err)
	}

	return &v1.RenameNamespaceResponse{}, nil
}

func (h *handlers) DeleteNamespace(ctx context.Context, in *v1.DeleteNamespaceRequest) (*v1.DeleteNamespaceResponse, error) {
	err := h.svc.DeleteNamespace(ctx, in.GetName())
	if err != nil {
		return nil, mapServiceError(err)
	}

	return &v1.DeleteNamespaceResponse{}, nil
}

func (h *handlers) ListNamespaces(ctx context.Context, in *v1.ListNamespacesRequest) (*v1.ListNamespacesResponse, error) {
	namespaces, err := h.svc.ListNamespaces(ctx)
	if err != nil {
		return nil, mapServiceError(err)
	}

	return &v1.ListNamespacesResponse{
		Name: namespaces,
	}, nil
}

func (h *handlers) CreateContainer(ctx context.Context, in *v1.CreateContainerRequest) (*v1.CreateContainerResponse, error) {
	err := h.svc.CreateContainer(ctx, in.GetNamespace(), in.GetName(), time.Duration(in.GetTtlSeconds())*time.Second)
	if err != nil {
		return nil, mapServiceError(err)
	}

	return &v1.CreateContainerResponse{}, nil
}

func (h *handlers) MoveContainer(ctx context.Context, in *v1.MoveContainerRequest) (*v1.MoveContainerResponse, error) {
	err := h.svc.MoveContainer(ctx, in.GetNamespace(), in.GetContainerName(), in.GetDestinationNamespace())
	if err != nil {
		return nil, mapServiceError(err)
	}

	return &v1.MoveContainerResponse{}, nil
}

func (h *handlers) RenameContainer(ctx context.Context, in *v1.RenameContainerRequest) (*v1.RenameContainerResponse, error) {
	err := h.svc.RenameContainer(ctx, in.GetNamespace(), in.GetOldName(), in.GetNewName())
	if err != nil {
		return nil, mapServiceError(err)
	}

	return &v1.RenameContainerResponse{}, nil
}

func (h *handlers) DeleteContainer(ctx context.Context, in *v1.DeleteContainerRequest) (*v1.DeleteContainerResponse, error) {
	err := h.svc.DeleteContainer(ctx, in.GetNamespace(), in.GetName())
	if err != nil {
		return nil, mapServiceError(err)
	}

	return &v1.DeleteContainerResponse{}, nil
}

func (h *handlers) SetContainerParameters(ctx context.Context, in *v1.SetContainerParametersRequest) (*v1.SetContainerParametersResponse, error) {
	err := h.svc.SetContainerParameters(ctx, in.GetNamespace(), in.GetName(), time.Duration(in.GetTtlSeconds())*time.Second)
	if err != nil {
		return nil, mapServiceError(err)
	}

	return &v1.SetContainerParametersResponse{}, nil
}

func (h *handlers) ListContainers(ctx context.Context, in *v1.ListContainersRequest) (*v1.ListContainersResponse, error) {
	containers, err := h.svc.ListContainers(ctx, in.GetNamespace())
	if err != nil {
		return nil, mapServiceError(err)
	}

	containerNames := []string{}
	for _, v := range containers {
		containerNames = append(containerNames, v.Name)
	}

	return &v1.ListContainersResponse{
		Name: containerNames,
	}, nil
}

func (h *handlers) CreateVersion(ctx context.Context, in *v1.CreateVersionRequest) (*v1.CreateVersionResponse, error) {
	version, err := h.svc.CreateVersion(ctx, in.GetNamespace(), in.GetContainer())
	if err != nil {
		return nil, mapServiceError(err)
	}

	return &v1.CreateVersionResponse{
		Version: version,
	}, nil
}

func (h *handlers) ListVersions(ctx context.Context, in *v1.ListVersionsRequest) (*v1.ListVersionsResponse, error) {
	versions, err := h.svc.ListAllVersions(ctx, in.GetNamespace(), in.GetContainer())
	if err != nil {
		return nil, mapServiceError(err)
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
	err := h.svc.DeleteVersion(ctx, in.GetNamespace(), in.GetContainer(), in.GetVersion())
	if err != nil {
		return nil, mapServiceError(err)
	}

	return &v1.DeleteVersionResponse{}, nil
}

func (h *handlers) PublishVersion(ctx context.Context, in *v1.PublishVersionRequest) (*v1.PublishVersionResponse, error) {
	err := h.svc.PublishVersion(ctx, in.GetNamespace(), in.GetContainer(), in.GetVersion())
	if err != nil {
		return nil, mapServiceError(err)
	}

	return &v1.PublishVersionResponse{}, nil
}

func (h *handlers) CreateObject(ctx context.Context, in *v1.CreateObjectRequest) (*v1.CreateObjectResponse, error) {
	url, err := h.svc.EnsureBLOBPresenceOrGetUploadURL(ctx, in.GetChecksum(), in.GetSize(), in.GetMimeType())
	if err != nil && url == "" {
		return nil, mapServiceError(err)
	}

	err = h.svc.AddObject(ctx, in.GetNamespace(), in.GetContainer(), in.GetVersion(), in.GetKey(), in.GetChecksum())
	if err != nil {
		return nil, mapServiceError(err)
	}

	return &v1.CreateObjectResponse{
		UploadUrl: ptr.String(url),
	}, nil
}

func (h *handlers) ListObjects(ctx context.Context, in *v1.ListObjectsRequest) (*v1.ListObjectsResponse, error) {
	objects, err := h.svc.ListObjects(ctx, in.GetNamespace(), in.GetContainer(), in.GetVersion())
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, mapServiceError(err)
	}

	return &v1.ListObjectsResponse{
		Objects: objects,
	}, nil
}

func (h *handlers) GetObjectURL(ctx context.Context, in *v1.GetObjectURLRequest) (*v1.GetObjectURLResponse, error) {
	url, err := h.svc.GetObjectURL(ctx, in.GetNamespace(), in.GetContainer(), in.GetVersion(), in.GetKey())
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, mapServiceError(err)
	}

	return &v1.GetObjectURLResponse{
		Url: url,
	}, nil
}

func (h *handlers) DeleteObject(ctx context.Context, in *v1.DeleteObjectRequest) (*v1.DeleteObjectResponse, error) {
	err := h.svc.DeleteObject(ctx, in.GetNamespace(), in.GetContainer(), in.GetVersion(), in.GetKey())
	if err != nil {
		return nil, mapServiceError(err)
	}

	return &v1.DeleteObjectResponse{}, nil
}

func (h *handlers) Register(gs *grpc.Server) {
	v1.RegisterManageServiceServer(gs, h)
}

func mapServiceError(err error) error {
	if errors.Is(err, service.ErrNotFound) {
		return status.Error(codes.NotFound, err.Error())
	}
	return status.Error(codes.Internal, err.Error())
}
