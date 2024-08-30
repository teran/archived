package metadata

import (
	"context"

	"github.com/pkg/errors"

	emodels "github.com/teran/archived/exporter/models"
	"github.com/teran/archived/models"
)

var (
	ErrNotFound = errors.New("not found")
	ErrConflict = errors.New("entity with given identifier already exists")
)

type Repository interface {
	CreateNamespace(ctx context.Context, name string) error
	RenameNamespace(ctx context.Context, oldName, newName string) error
	ListNamespaces(ctx context.Context) ([]string, error)
	DeleteNamespace(ctx context.Context, name string) error

	CreateContainer(ctx context.Context, namespace, name string) error
	RenameContainer(ctx context.Context, namespace, oldName, newNamespace, newName string) error
	ListContainers(ctx context.Context, namespace string) ([]models.Container, error)
	ListContainersByPage(ctx context.Context, namespace string, offset, limit uint64) (uint64, []models.Container, error)
	DeleteContainer(ctx context.Context, namespace, name string) error

	CreateVersion(ctx context.Context, namespace, container string) (string, error)
	GetLatestPublishedVersionByContainer(ctx context.Context, namespace, container string) (string, error)
	ListAllVersionsByContainer(ctx context.Context, namespace, container string) ([]models.Version, error)
	ListPublishedVersionsByContainer(ctx context.Context, namespace, container string) ([]models.Version, error)
	ListPublishedVersionsByContainerAndPage(ctx context.Context, namespace, container string, offset, limit uint64) (uint64, []models.Version, error)
	ListUnpublishedVersionsByContainer(ctx context.Context, namespace, container string) ([]models.Version, error)
	MarkVersionPublished(ctx context.Context, namespace, container, version string) error
	DeleteVersion(ctx context.Context, namespace, container, version string) error

	CreateObject(ctx context.Context, namespace, container, version, key, casKey string) error
	ListObjects(ctx context.Context, namespace, container, version string, offset, limit uint64) (uint64, []string, error)
	DeleteObject(ctx context.Context, namespace, container, version string, key ...string) error
	RemapObject(ctx context.Context, namespace, container, version, key, newCASKey string) error

	CreateBLOB(ctx context.Context, checksum string, size uint64, mimeType string) error
	GetBlobKeyByObject(ctx context.Context, namespace, scontainer, version, key string) (string, error)
	EnsureBlobKey(ctx context.Context, key string, size uint64) error

	CountStats(ctx context.Context) (*emodels.Stats, error)
}
