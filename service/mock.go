package service

import (
	"context"

	"github.com/stretchr/testify/mock"
	"github.com/teran/archived/models"
)

var (
	_ Manager   = (*Mock)(nil)
	_ Publisher = (*Mock)(nil)
)

type Mock struct {
	mock.Mock
}

func NewMock() *Mock {
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

func (m *Mock) CreateVersion(_ context.Context, container string) (id string, err error) {
	args := m.Called(container)
	return args.String(0), args.Error(1)
}

func (m *Mock) ListAllVersions(_ context.Context, container string) ([]models.Version, error) {
	args := m.Called(container)
	return args.Get(0).([]models.Version), args.Error(1)
}

func (m *Mock) ListPublishedVersions(_ context.Context, container string) ([]models.Version, error) {
	args := m.Called(container)
	return args.Get(0).([]models.Version), args.Error(1)
}

func (m *Mock) ListPublishedVersionsByPage(_ context.Context, container string, pageNum uint64) (uint64, []models.Version, error) {
	args := m.Called(container, pageNum)
	return args.Get(0).(uint64), args.Get(1).([]models.Version), args.Error(2)
}

func (m *Mock) PublishVersion(_ context.Context, container, id string) error {
	args := m.Called(container, id)
	return args.Error(0)
}

func (m *Mock) DeleteVersion(_ context.Context, container, id string) error {
	args := m.Called(container, id)
	return args.Error(0)
}

func (m *Mock) AddObject(_ context.Context, container, versionID, key, casKey string) error {
	args := m.Called(container, versionID, key, casKey)
	return args.Error(0)
}

func (m *Mock) EnsureBLOBPresenceOrGetUploadURL(ctx context.Context, checksum string, size int64) (string, error) {
	args := m.Called(checksum, size)
	return args.String(0), args.Error(1)
}

func (m *Mock) ListObjects(_ context.Context, container, versionID string) ([]string, error) {
	args := m.Called(container, versionID)
	return args.Get(0).([]string), args.Error(1)
}

func (m *Mock) ListObjectsByPage(_ context.Context, container, versionID string, pageNum uint64) (uint64, []string, error) {
	args := m.Called(container, versionID, pageNum)
	return args.Get(0).(uint64), args.Get(1).([]string), args.Error(2)
}

func (m *Mock) GetObjectURL(ctx context.Context, container, versionID, key string) (string, error) {
	args := m.Called(container, versionID, key)
	return args.String(0), args.Error(1)
}

func (m *Mock) DeleteObject(_ context.Context, container, versionID, key string) error {
	args := m.Called(container)
	return args.Error(0)
}
