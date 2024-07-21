package service

import (
	"context"

	"github.com/pkg/errors"

	"github.com/teran/archived/repositories/blob"
	"github.com/teran/archived/repositories/metadata"
)

const defaultVersionsPerPage = 50

var ErrNotFound = errors.New("entity not found")

type Manager interface {
	Publisher

	CreateContainer(ctx context.Context, name string) error
	DeleteContainer(ctx context.Context, name string) error

	CreateVersion(ctx context.Context, container string) (id string, err error)
	ListAllVersions(ctx context.Context, container string) ([]string, error)
	PublishVersion(ctx context.Context, container, id string) error
	DeleteVersion(ctx context.Context, container, id string) error

	AddObject(ctx context.Context, container, versionID, key string, casKey string) error
	DeleteObject(ctx context.Context, container, versionID, key string) error

	EnsureBLOBPresenceOrGetUploadURL(ctx context.Context, checksum string, size int64) (string, error)
}

type Publisher interface {
	ListContainers(ctx context.Context) ([]string, error)

	ListPublishedVersions(ctx context.Context, container string) ([]string, error)
	ListPublishedVersionsByPage(ctx context.Context, container string, pageNum uint64) (uint64, []string, error)

	ListObjects(ctx context.Context, container, versionID string) ([]string, error)
	GetObjectURL(ctx context.Context, container, versionID, key string) (string, error)
}

type service struct {
	mdRepo           metadata.Repository
	blobRepo         blob.Repository
	versionsPageSize uint64
}

func NewManager(mdRepo metadata.Repository, blobRepo blob.Repository) Manager {
	return newSvc(mdRepo, blobRepo)
}

func NewPublisher(mdRepo metadata.Repository, blobRepo blob.Repository) Publisher {
	return newSvc(mdRepo, blobRepo)
}

func newSvc(mdRepo metadata.Repository, blobRepo blob.Repository) *service {
	return &service{
		mdRepo:           mdRepo,
		blobRepo:         blobRepo,
		versionsPageSize: defaultVersionsPerPage,
	}
}

func (s *service) CreateContainer(ctx context.Context, name string) error {
	err := s.mdRepo.CreateContainer(ctx, name)
	if err != nil {
		return errors.Wrap(err, "error creating container")
	}
	return err
}

func (s *service) ListContainers(ctx context.Context) ([]string, error) {
	containers, err := s.mdRepo.ListContainers(ctx)
	return containers, mapMetadataErrors(err)
}

func (s *service) DeleteContainer(ctx context.Context, name string) error {
	err := s.mdRepo.DeleteContainer(ctx, name)
	return mapMetadataErrors(err)
}

func (s *service) CreateVersion(ctx context.Context, container string) (id string, err error) {
	version, err := s.mdRepo.CreateVersion(ctx, container)
	return version, mapMetadataErrors(err)
}

func (s *service) ListPublishedVersions(ctx context.Context, container string) ([]string, error) {
	versions, err := s.mdRepo.ListPublishedVersionsByContainer(ctx, container)
	return versions, mapMetadataErrors(err)
}

func (s *service) ListPublishedVersionsByPage(ctx context.Context, container string, pageNum uint64) (uint64, []string, error) {
	if pageNum < 1 {
		pageNum = 1
	}

	offset := (pageNum - 1) * s.versionsPageSize
	limit := s.versionsPageSize
	totalVersions, versions, err := s.mdRepo.ListPublishedVersionsByContainerAndPage(ctx, container, offset, limit)
	if err != nil {
		return 0, nil, err
	}

	totalPages := (totalVersions / s.versionsPageSize)
	if (totalVersions % s.versionsPageSize) != 0 {
		totalPages++
	}

	return totalPages, versions, mapMetadataErrors(err)
}

func (s *service) ListAllVersions(ctx context.Context, container string) ([]string, error) {
	versions, err := s.mdRepo.ListAllVersionsByContainer(ctx, container)
	return versions, mapMetadataErrors(err)
}

func (s *service) PublishVersion(ctx context.Context, container, id string) error {
	err := s.mdRepo.MarkVersionPublished(ctx, container, id)
	return mapMetadataErrors(err)
}

func (s *service) DeleteVersion(ctx context.Context, container, id string) error {
	panic("not implemented")
}

func (s *service) AddObject(ctx context.Context, container, versionID, key, casKey string) error {
	return s.mdRepo.CreateObject(ctx, container, versionID, key, casKey)
}

func (s *service) ListObjects(ctx context.Context, container, versionID string) ([]string, error) {
	_, objects, err := s.mdRepo.ListObjects(ctx, container, versionID, 0, 1000)
	return objects, mapMetadataErrors(err)
}

func (s *service) ListObjectsByPage(ctx context.Context, container, versionID string, pageNum uint64) (uint64, []string, error) {
	if pageNum < 1 {
		pageNum = 1
	}

	offset := (pageNum - 1) * s.versionsPageSize
	limit := s.versionsPageSize
	totalVersions, versions, err := s.mdRepo.ListObjects(ctx, container, versionID, offset, limit)
	if err != nil {
		return 0, nil, err
	}

	totalPages := (totalVersions / s.versionsPageSize)
	if (totalVersions % s.versionsPageSize) != 0 {
		totalPages++
	}

	return totalPages, versions, mapMetadataErrors(err)
}

func (s *service) GetObjectURL(ctx context.Context, container, versionID, key string) (string, error) {
	key, err := s.mdRepo.GetBlobKeyByObject(ctx, container, versionID, key)
	if err != nil {
		return "", mapMetadataErrors(err)
	}

	return s.blobRepo.GetBlobURL(ctx, key)
}

func (s *service) EnsureBLOBPresenceOrGetUploadURL(ctx context.Context, checksum string, size int64) (string, error) {
	err := s.mdRepo.EnsureBlobKey(ctx, checksum, uint64(size))
	if err == nil {
		return "", nil
	}

	if err == metadata.ErrNotFound {
		url, err := s.blobRepo.PutBlobURL(ctx, checksum)
		if err != nil {
			return "", err
		}
		return url, s.mdRepo.CreateBLOB(ctx, checksum, uint64(size), "application/octet-stream")
	}

	return "", err
}

func (s *service) DeleteObject(ctx context.Context, container, versionID, key string) error {
	panic("not implemented")
}

func mapMetadataErrors(err error) error {
	switch err {
	case metadata.ErrNotFound:
		return ErrNotFound
	default:
		return err
	}
}
