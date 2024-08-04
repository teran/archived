package metadata

import (
	"context"

	"github.com/pkg/errors"

	emodels "github.com/teran/archived/exporter/models"
	"github.com/teran/archived/models"
)

var ErrNotFound = errors.New("not found")

type Repository interface {
	CreateContainer(ctx context.Context, name string) error
	ListContainers(ctx context.Context) ([]string, error)
	DeleteContainer(ctx context.Context, name string) error

	CreateVersion(ctx context.Context, container string) (string, error)
	GetLatestPublishedVersionByContainer(ctx context.Context, container string) (string, error)
	ListAllVersionsByContainer(ctx context.Context, container string) ([]models.Version, error)
	ListPublishedVersionsByContainer(ctx context.Context, container string) ([]models.Version, error)
	ListPublishedVersionsByContainerAndPage(ctx context.Context, container string, offset, limit uint64) (uint64, []models.Version, error)
	ListUnpublishedVersionsByContainer(ctx context.Context, container string) ([]models.Version, error)
	MarkVersionPublished(ctx context.Context, container, version string) error
	DeleteVersion(ctx context.Context, container, version string) error

	CreateObject(ctx context.Context, container, version, key, casKey string) error
	ListObjects(ctx context.Context, container, version string, offset, limit uint64) (uint64, []string, error)
	DeleteObject(ctx context.Context, container, version string, key ...string) error
	RemapObject(ctx context.Context, container, version, key, newCASKey string) error

	CreateBLOB(ctx context.Context, checksum string, size uint64, mimeType string) error
	GetBlobKeyByObject(ctx context.Context, container, version, key string) (string, error)
	EnsureBlobKey(ctx context.Context, key string, size uint64) error

	CountStats(ctx context.Context) (*emodels.Stats, error)
}
