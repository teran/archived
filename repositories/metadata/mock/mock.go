package mock

import (
	"context"

	"github.com/stretchr/testify/mock"
	"github.com/teran/archived/exporter/models"
	"github.com/teran/archived/repositories/metadata"
)

var _ metadata.Repository = (*Mock)(nil)

type Mock struct {
	mock.Mock
}

func New() *Mock {
	return &Mock{}
}

func (m *Mock) CreateContainer(_ context.Context, name string) error {
	args := m.Called(name)
	return args.Error(0)
}

func (m *Mock) ListContainers(context.Context) ([]string, error) {
	args := m.Called()
	return args.Get(0).([]string), args.Error(1)
}

func (m *Mock) DeleteContainer(_ context.Context, name string) error {
	args := m.Called(name)
	return args.Error(0)
}

func (m *Mock) CreateVersion(_ context.Context, container string) (string, error) {
	args := m.Called(container)
	return args.String(0), args.Error(1)
}

func (m *Mock) GetLatestPublishedVersionByContainer(_ context.Context, container string) (string, error) {
	args := m.Called(container)
	return args.String(0), args.Error(1)
}

func (m *Mock) ListAllVersionsByContainer(_ context.Context, container string) ([]string, error) {
	args := m.Called(container)
	return args.Get(0).([]string), args.Error(1)
}

func (m *Mock) ListPublishedVersionsByContainer(_ context.Context, container string) ([]string, error) {
	args := m.Called(container)
	return args.Get(0).([]string), args.Error(1)
}

func (m *Mock) ListPublishedVersionsByContainerAndPage(_ context.Context, container string, offset, limit uint64) (uint64, []string, error) {
	args := m.Called(container, offset, limit)
	return args.Get(0).(uint64), args.Get(1).([]string), args.Error(2)
}

func (m *Mock) ListUnpublishedVersionsByContainer(_ context.Context, container string) ([]string, error) {
	args := m.Called(container)
	return args.Get(0).([]string), args.Error(1)
}

func (m *Mock) MarkVersionPublished(_ context.Context, container, version string) error {
	args := m.Called(container, version)
	return args.Error(0)
}

func (m *Mock) DeleteVersion(ctx context.Context, container, version string) error {
	args := m.Called(container, version)
	return args.Error(0)
}

func (m *Mock) CreateObject(_ context.Context, container, version, key, casKey string) error {
	args := m.Called(container, version, key, casKey)
	return args.Error(0)
}

func (m *Mock) ListObjects(_ context.Context, container, version string, offset, limit uint64) (uint64, []string, error) {
	args := m.Called(container, version, offset, limit)
	return args.Get(0).(uint64), args.Get(1).([]string), args.Error(2)
}

func (m *Mock) DeleteObject(_ context.Context, container, version string, key ...string) error {
	args := m.Called(container, version, key)
	return args.Error(0)
}

func (m *Mock) RemapObject(_ context.Context, container, version, key, newCASKey string) error {
	args := m.Called(container, version, key, newCASKey)
	return args.Error(0)
}

func (m *Mock) CreateBLOB(_ context.Context, checksum string, size uint64, mimeType string) error {
	args := m.Called(checksum, size, mimeType)
	return args.Error(0)
}

func (m *Mock) GetBlobKeyByObject(_ context.Context, container, version, key string) (string, error) {
	args := m.Called(container, version, key)
	return args.String(0), args.Error(1)
}

func (m *Mock) EnsureBlobKey(_ context.Context, key string, size uint64) error {
	args := m.Called(key, size)
	return args.Error(0)
}

func (m *Mock) CountStats(ctx context.Context) (*models.Stats, error) {
	args := m.Called()
	return args.Get(0).(*models.Stats), args.Error(1)
}
