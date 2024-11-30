package mock

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"

	emodels "github.com/teran/archived/exporter/models"
	"github.com/teran/archived/models"
	"github.com/teran/archived/repositories/metadata"
)

var _ metadata.Repository = (*Mock)(nil)

type Mock struct {
	mock.Mock
}

func New() *Mock {
	return &Mock{}
}

func (m *Mock) CreateNamespace(ctx context.Context, name string) error {
	args := m.Called(name)
	return args.Error(0)
}

func (m *Mock) RenameNamespace(ctx context.Context, oldName, newName string) error {
	args := m.Called(oldName, newName)
	return args.Error(0)
}

func (m *Mock) ListNamespaces(ctx context.Context) ([]string, error) {
	args := m.Called()
	return args.Get(0).([]string), args.Error(1)
}

func (m *Mock) DeleteNamespace(ctx context.Context, name string) error {
	args := m.Called(name)
	return args.Error(0)
}

func (m *Mock) CreateContainer(_ context.Context, namespace, name string, ttl time.Duration) error {
	args := m.Called(namespace, name, ttl)
	return args.Error(0)
}

func (m *Mock) RenameContainer(_ context.Context, namespace, oldName, newNamespace, newName string) error {
	args := m.Called(namespace, oldName, newNamespace, newName)
	return args.Error(0)
}

func (m *Mock) SetContainerParameters(_ context.Context, namespace, name string, ttl time.Duration) error {
	args := m.Called(namespace, name, ttl)
	return args.Error(0)
}

func (m *Mock) ListContainers(_ context.Context, namespace string) ([]models.Container, error) {
	args := m.Called(namespace)
	return args.Get(0).([]models.Container), args.Error(1)
}

func (m *Mock) ListContainersByPage(_ context.Context, namespace string, offset, limit uint64) (uint64, []models.Container, error) {
	args := m.Called(namespace, offset, limit)
	return args.Get(0).(uint64), args.Get(1).([]models.Container), args.Error(2)
}

func (m *Mock) DeleteContainer(_ context.Context, namespace, name string) error {
	args := m.Called(namespace, name)
	return args.Error(0)
}

func (m *Mock) CreateVersion(_ context.Context, namespace, container string) (string, error) {
	args := m.Called(namespace, container)
	return args.String(0), args.Error(1)
}

func (m *Mock) GetLatestPublishedVersionByContainer(_ context.Context, namespace, container string) (string, error) {
	args := m.Called(namespace, container)
	return args.String(0), args.Error(1)
}

func (m *Mock) ListAllVersionsByContainer(_ context.Context, namespace, container string) ([]models.Version, error) {
	args := m.Called(namespace, container)
	return args.Get(0).([]models.Version), args.Error(1)
}

func (m *Mock) ListPublishedVersionsByContainer(_ context.Context, namespace, container string) ([]models.Version, error) {
	args := m.Called(namespace, container)
	return args.Get(0).([]models.Version), args.Error(1)
}

func (m *Mock) ListPublishedVersionsByContainerAndPage(_ context.Context, namespace, container string, offset, limit uint64) (uint64, []models.Version, error) {
	args := m.Called(namespace, container, offset, limit)
	return args.Get(0).(uint64), args.Get(1).([]models.Version), args.Error(2)
}

func (m *Mock) ListUnpublishedVersionsByContainer(_ context.Context, namespace, container string) ([]models.Version, error) {
	args := m.Called(namespace, container)
	return args.Get(0).([]models.Version), args.Error(1)
}

func (m *Mock) MarkVersionPublished(_ context.Context, namespace, container, version string) error {
	args := m.Called(namespace, container, version)
	return args.Error(0)
}

func (m *Mock) DeleteVersion(ctx context.Context, namespace, container, version string) error {
	args := m.Called(namespace, container, version)
	return args.Error(0)
}

func (m *Mock) DeleteExpiredVersionsWithObjects(ctx context.Context, isPublished *bool) error {
	args := m.Called(isPublished)
	return args.Error(0)
}

func (m *Mock) CreateObject(_ context.Context, namespace, container, version, key, casKey string) error {
	args := m.Called(namespace, container, version, key, casKey)
	return args.Error(0)
}

func (m *Mock) ListObjects(_ context.Context, namespace, container, version string, offset, limit uint64) (uint64, []string, error) {
	args := m.Called(namespace, container, version, offset, limit)
	return args.Get(0).(uint64), args.Get(1).([]string), args.Error(2)
}

func (m *Mock) DeleteObject(_ context.Context, namespace, container, version string, key ...string) error {
	args := m.Called(namespace, container, version, key)
	return args.Error(0)
}

func (m *Mock) RemapObject(_ context.Context, namespace, container, version, key, newCASKey string) error {
	args := m.Called(namespace, container, version, key, newCASKey)
	return args.Error(0)
}

func (m *Mock) CreateBLOB(_ context.Context, checksum string, size uint64, mimeType string) error {
	args := m.Called(checksum, size, mimeType)
	return args.Error(0)
}

func (m *Mock) GetBlobKeyByObject(_ context.Context, namespace, container, version, key string) (string, error) {
	args := m.Called(namespace, container, version, key)
	return args.String(0), args.Error(1)
}

func (m *Mock) GetBlobByObject(_ context.Context, namespace, container, version, key string) (models.Blob, error) {
	args := m.Called(namespace, container, version, key)
	return args.Get(0).(models.Blob), args.Error(1)
}

func (m *Mock) EnsureBlobKey(_ context.Context, key string, size uint64) error {
	args := m.Called(key, size)
	return args.Error(0)
}

func (m *Mock) CountStats(ctx context.Context) (*emodels.Stats, error) {
	args := m.Called()
	return args.Get(0).(*emodels.Stats), args.Error(1)
}
