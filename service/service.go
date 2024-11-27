package service

import (
	"context"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/teran/archived/models"
	"github.com/teran/archived/repositories/blob"
	"github.com/teran/archived/repositories/metadata"
)

var ErrNotFound = errors.New("entity not found")

type Manager interface {
	Publisher

	CreateNamespace(ctx context.Context, name string) error
	RenameNamespace(ctx context.Context, oldName, newName string) error
	DeleteNamespace(ctx context.Context, name string) error

	CreateContainer(ctx context.Context, namespace, name string) error
	MoveContainer(ctx context.Context, namespace, container, destNamespace string) error
	RenameContainer(ctx context.Context, namespace, oldName, newName string) error
	DeleteContainer(ctx context.Context, namespace, name string) error
	SetContainerParameters(ctx context.Context, namespace, name string, ttl time.Duration) error

	CreateVersion(ctx context.Context, namespace, container string) (id string, err error)
	ListAllVersions(ctx context.Context, namespace, container string) ([]models.Version, error)
	PublishVersion(ctx context.Context, namespace, container, id string) error
	DeleteVersion(ctx context.Context, namespace, container, id string) error

	AddObject(ctx context.Context, namespace, container, versionID, key string, casKey string) error
	ListObjects(ctx context.Context, namespace, container, versionID string) ([]string, error)
	DeleteObject(ctx context.Context, namespace, container, versionID, key string) error

	EnsureBLOBPresenceOrGetUploadURL(ctx context.Context, checksum string, size uint64, mimeType string) (string, error)
}

type Publisher interface {
	ListNamespaces(ctx context.Context) ([]string, error)

	ListContainers(ctx context.Context, namespace string) ([]models.Container, error)
	ListContainersByPage(ctx context.Context, namespace string, pageNum uint64) (uint64, []models.Container, error)

	ListPublishedVersions(ctx context.Context, namespace, container string) ([]models.Version, error)
	ListPublishedVersionsByPage(ctx context.Context, namespace, container string, pageNum uint64) (uint64, []models.Version, error)

	ListObjectsByPage(ctx context.Context, namespace, container, versionID string, pageNum uint64) (uint64, []string, error)
	GetObjectURL(ctx context.Context, namespace, container, versionID, key string) (string, error)
}

type service struct {
	mdRepo             metadata.Repository
	blobRepo           blob.Repository
	versionsPageSize   uint64
	objectsPageSize    uint64
	containersPageSize uint64
}

func NewManager(mdRepo metadata.Repository, blobRepo blob.Repository) Manager {
	return newSvc(mdRepo, blobRepo, 50, 50, 50)
}

func NewPublisher(mdRepo metadata.Repository, blobRepo blob.Repository, versionsPerPage, objectsPerPage, containersPerPage uint64) Publisher {
	return newSvc(mdRepo, blobRepo, versionsPerPage, objectsPerPage, containersPerPage)
}

func newSvc(mdRepo metadata.Repository, blobRepo blob.Repository, versionsPerPage, objectsPerPage, containersPerPage uint64) *service {
	return &service{
		mdRepo:             mdRepo,
		blobRepo:           blobRepo,
		versionsPageSize:   versionsPerPage,
		objectsPageSize:    objectsPerPage,
		containersPageSize: containersPerPage,
	}
}

func (s *service) CreateNamespace(ctx context.Context, name string) error {
	err := s.mdRepo.CreateNamespace(ctx, name)
	if err != nil {
		return mapMetadataErrors(err)
	}
	return nil
}

func (s *service) RenameNamespace(ctx context.Context, oldName, newName string) error {
	err := s.mdRepo.RenameNamespace(ctx, oldName, newName)
	if err != nil {
		return mapMetadataErrors(err)
	}
	return nil
}

func (s *service) ListNamespaces(ctx context.Context) ([]string, error) {
	containers, err := s.mdRepo.ListNamespaces(ctx)
	return containers, mapMetadataErrors(err)
}

func (s *service) DeleteNamespace(ctx context.Context, name string) error {
	err := s.mdRepo.DeleteNamespace(ctx, name)
	return mapMetadataErrors(err)
}

func (s *service) CreateContainer(ctx context.Context, namespace, name string) error {
	err := s.mdRepo.CreateContainer(ctx, namespace, name)
	if err != nil {
		return mapMetadataErrors(err)
	}
	return nil
}

func (s *service) MoveContainer(ctx context.Context, namespace, container, destNamespace string) error {
	err := s.mdRepo.RenameContainer(ctx, namespace, container, destNamespace, container)
	if err != nil {
		return mapMetadataErrors(err)
	}
	return nil
}

func (s *service) RenameContainer(ctx context.Context, namespace, oldName, newName string) error {
	err := s.mdRepo.RenameContainer(ctx, namespace, oldName, namespace, newName)
	if err != nil {
		return mapMetadataErrors(err)
	}
	return nil
}

func (s *service) SetContainerParameters(ctx context.Context, namespace, name string, ttl time.Duration) error {
	err := s.mdRepo.SetContainerParameters(ctx, namespace, name, ttl)
	if err != nil {
		return mapMetadataErrors(err)
	}
	return nil
}

func (s *service) ListContainers(ctx context.Context, namespace string) ([]models.Container, error) {
	containers, err := s.mdRepo.ListContainers(ctx, namespace)
	return containers, mapMetadataErrors(err)
}

func (s *service) ListContainersByPage(ctx context.Context, namespace string, pageNum uint64) (uint64, []models.Container, error) {
	if pageNum < 1 {
		pageNum = 1
	}

	offset := (pageNum - 1) * s.containersPageSize
	limit := s.containersPageSize
	totalContainers, containers, err := s.mdRepo.ListContainersByPage(ctx, namespace, offset, limit)
	if err != nil {
		return 0, nil, mapMetadataErrors(err)
	}

	totalPages := (totalContainers / s.containersPageSize)
	if (totalContainers % s.containersPageSize) != 0 {
		totalPages++
	}

	return totalPages, containers, mapMetadataErrors(err)
}

func (s *service) DeleteContainer(ctx context.Context, namespace, name string) error {
	err := s.mdRepo.DeleteContainer(ctx, namespace, name)
	return mapMetadataErrors(err)
}

func (s *service) CreateVersion(ctx context.Context, namespace, container string) (id string, err error) {
	version, err := s.mdRepo.CreateVersion(ctx, namespace, container)
	return version, mapMetadataErrors(err)
}

func (s *service) ListPublishedVersions(ctx context.Context, namespace, container string) ([]models.Version, error) {
	versions, err := s.mdRepo.ListPublishedVersionsByContainer(ctx, namespace, container)
	return versions, mapMetadataErrors(err)
}

func (s *service) ListPublishedVersionsByPage(ctx context.Context, namespace, container string, pageNum uint64) (uint64, []models.Version, error) {
	if pageNum < 1 {
		pageNum = 1
	}

	offset := (pageNum - 1) * s.versionsPageSize
	limit := s.versionsPageSize
	totalVersions, versions, err := s.mdRepo.ListPublishedVersionsByContainerAndPage(ctx, namespace, container, offset, limit)
	if err != nil {
		return 0, nil, mapMetadataErrors(err)
	}

	totalPages := (totalVersions / s.versionsPageSize)
	if (totalVersions % s.versionsPageSize) != 0 {
		totalPages++
	}

	return totalPages, versions, mapMetadataErrors(err)
}

func (s *service) ListAllVersions(ctx context.Context, namespace, container string) ([]models.Version, error) {
	versions, err := s.mdRepo.ListAllVersionsByContainer(ctx, namespace, container)
	return versions, mapMetadataErrors(err)
}

func (s *service) PublishVersion(ctx context.Context, namespace, container, id string) error {
	err := s.mdRepo.MarkVersionPublished(ctx, namespace, container, id)
	return mapMetadataErrors(err)
}

func (s *service) DeleteVersion(ctx context.Context, namespace, container, id string) error {
	err := s.mdRepo.DeleteVersion(ctx, namespace, container, id)
	return mapMetadataErrors(err)
}

func (s *service) AddObject(ctx context.Context, namespace, container, versionID, key, casKey string) error {
	err := s.mdRepo.CreateObject(ctx, namespace, container, versionID, strings.TrimPrefix(key, "/"), casKey)
	return mapMetadataErrors(err)
}

func (s *service) ListObjects(ctx context.Context, namespace, container, versionID string) ([]string, error) {
	_, objects, err := s.mdRepo.ListObjects(ctx, namespace, container, versionID, 0, 1000)
	return objects, mapMetadataErrors(err)
}

func (s *service) ListObjectsByPage(ctx context.Context, namespace, container, versionID string, pageNum uint64) (uint64, []string, error) {
	var err error
	if versionID == "latest" {
		versionID, err = s.mdRepo.GetLatestPublishedVersionByContainer(ctx, namespace, container)
		if err != nil {
			return 0, nil, mapMetadataErrors(err)
		}
	}

	if pageNum < 1 {
		pageNum = 1
	}

	offset := (pageNum - 1) * s.objectsPageSize
	limit := s.objectsPageSize
	totalObjects, objects, err := s.mdRepo.ListObjects(ctx, namespace, container, versionID, offset, limit)
	if err != nil {
		return 0, nil, mapMetadataErrors(err)
	}

	totalPages := (totalObjects / s.objectsPageSize)
	if (totalObjects % s.objectsPageSize) != 0 {
		totalPages++
	}

	return totalPages, objects, mapMetadataErrors(err)
}

func (s *service) GetObjectURL(ctx context.Context, namespace, container, versionID, key string) (string, error) {
	var err error
	if versionID == "latest" {
		versionID, err = s.mdRepo.GetLatestPublishedVersionByContainer(ctx, namespace, container)
		if err != nil {
			return "", mapMetadataErrors(err)
		}
	}

	blob, err := s.mdRepo.GetBlobByObject(ctx, namespace, container, versionID, key)
	if err != nil {
		return "", mapMetadataErrors(err)
	}

	return s.blobRepo.GetBlobURL(ctx, blob.Checksum, blob.MimeType, key)
}

func (s *service) EnsureBLOBPresenceOrGetUploadURL(ctx context.Context, checksum string, size uint64, mimeType string) (string, error) {
	err := s.mdRepo.EnsureBlobKey(ctx, checksum, size)
	if err == nil {
		return "", nil
	}

	if errors.Is(err, metadata.ErrNotFound) {
		url, err := s.blobRepo.PutBlobURL(ctx, checksum)
		if err != nil {
			return "", err
		}
		return url, s.mdRepo.CreateBLOB(ctx, checksum, size, mimeType)
	}

	return "", err
}

func (s *service) DeleteObject(ctx context.Context, namespace, container, versionID, key string) error {
	err := s.mdRepo.DeleteObject(ctx, namespace, container, versionID, key)
	return mapMetadataErrors(err)
}

func mapMetadataErrors(err error) error {
	switch {
	case errors.Is(err, metadata.ErrNotFound):
		return ErrNotFound
	default:
		return err
	}
}
