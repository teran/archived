package service

import (
	"context"
	"io"

	"github.com/pkg/errors"

	"github.com/teran/archived/repositories/blob"
	"github.com/teran/archived/repositories/metadata"
)

var ErrNotFound = errors.New("entity not found")

type ManageService interface {
	AccessService

	CreateContainer(ctx context.Context, name string) error
	DeleteContainer(ctx context.Context, name string) error

	CreateVersion(ctx context.Context, container string) (id string, err error)
	PublishVersion(ctx context.Context, container, id string) error
	DeleteVersion(ctx context.Context, container, id string) error

	AddObject(ctx context.Context, container, versionID, key string, objReader io.Reader) error
	DeleteObject(ctx context.Context, container, versionID, key string) error
}

type AccessService interface {
	ListContainers(ctx context.Context) ([]string, error)

	ListVersions(ctx context.Context, container string) ([]string, error)

	ListObjects(ctx context.Context, container, versionID string) ([]string, error)
	GetObjectURL(ctx context.Context, container, versionID, key string) (string, error)
}

type service struct {
	mdRepo   metadata.Repository
	blobRepo blob.Repository
}

func NewManageService(mdRepo metadata.Repository, blobRepo blob.Repository) ManageService {
	return &service{}
}

func NewAccessService(mdRepo metadata.Repository, blobRepo blob.Repository) AccessService {
	return &service{}
}

func (s *service) CreateContainer(ctx context.Context, name string) error {
	return s.mdRepo.CreateContainer(ctx, name)
}

func (s *service) ListContainers(ctx context.Context) ([]string, error) {
	return s.mdRepo.ListContainers(ctx)
}

func (s *service) DeleteContainer(ctx context.Context, name string) error {
	return s.mdRepo.DeleteContainer(ctx, name)
}

func (s *service) CreateVersion(ctx context.Context, container string) (id string, err error) {
	panic("not implemented")
}

func (s *service) ListVersions(ctx context.Context, container string) ([]string, error) {
	panic("not implemented")
}

func (s *service) PublishVersion(ctx context.Context, container, id string) error {
	panic("not implemented")
}

func (s *service) DeleteVersion(ctx context.Context, container, id string) error {
	panic("not implemented")
}

func (s *service) AddObject(ctx context.Context, container, versionID, key string, objReader io.Reader) error {
	panic("not implemented")
}

func (s *service) ListObjects(ctx context.Context, container, versionID string) ([]string, error) {
	panic("not implemented")
}

func (s *service) GetObjectURL(ctx context.Context, container, versionID, key string) (string, error) {
	panic("not implemented")
}

func (s *service) DeleteObject(ctx context.Context, container, versionID, key string) error {
	panic("not implemented")
}
