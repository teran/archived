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

func (m *Mock) CreateNamespace(_ context.Context, name string) error {
	args := m.Called(name)
	return args.Error(0)
}

func (m *Mock) ListNamespaces(ctx context.Context) ([]string, error) {
	args := m.Called()
	return args.Get(0).([]string), args.Error(1)
}

func (m *Mock) RenameNamespace(ctx context.Context, oldName, newName string) error {
	args := m.Called(oldName, newName)
	return args.Error(0)
}

func (m *Mock) DeleteNamespace(ctx context.Context, name string) error {
	args := m.Called(name)
	return args.Error(0)
}

func (m *Mock) CreateContainer(_ context.Context, namespace, name string) error {
	args := m.Called(namespace, name)
	return args.Error(0)
}

func (m *Mock) MoveContainer(ctx context.Context, namespace, container, destNamespace string) error {
	args := m.Called(namespace, container, destNamespace)
	return args.Error(0)
}

func (m *Mock) RenameContainer(_ context.Context, namespace, oldName, newName string) error {
	args := m.Called(namespace, oldName, newName)
	return args.Error(0)
}

func (m *Mock) ListContainers(_ context.Context, namespace string) ([]string, error) {
	args := m.Called(namespace)
	return args.Get(0).([]string), args.Error(1)
}

func (m *Mock) DeleteContainer(_ context.Context, namespace, name string) error {
	args := m.Called(namespace, name)
	return args.Error(0)
}

func (m *Mock) CreateVersion(_ context.Context, namespace, container string) (id string, err error) {
	args := m.Called(namespace, container)
	return args.String(0), args.Error(1)
}

func (m *Mock) ListAllVersions(_ context.Context, namespace, container string) ([]models.Version, error) {
	args := m.Called(namespace, container)
	return args.Get(0).([]models.Version), args.Error(1)
}

func (m *Mock) ListPublishedVersions(_ context.Context, namespace, container string) ([]models.Version, error) {
	args := m.Called(namespace, container)
	return args.Get(0).([]models.Version), args.Error(1)
}

func (m *Mock) ListPublishedVersionsByPage(_ context.Context, namespace, container string, pageNum uint64) (uint64, []models.Version, error) {
	args := m.Called(namespace, container, pageNum)
	return args.Get(0).(uint64), args.Get(1).([]models.Version), args.Error(2)
}

func (m *Mock) PublishVersion(_ context.Context, namespace, container, id string) error {
	args := m.Called(namespace, container, id)
	return args.Error(0)
}

func (m *Mock) DeleteVersion(_ context.Context, namespace, container, id string) error {
	args := m.Called(namespace, container, id)
	return args.Error(0)
}

func (m *Mock) AddObject(_ context.Context, namespace, container, versionID, key, casKey string) error {
	args := m.Called(namespace, container, versionID, key, casKey)
	return args.Error(0)
}

func (m *Mock) EnsureBLOBPresenceOrGetUploadURL(ctx context.Context, checksum string, size uint64) (string, error) {
	args := m.Called(checksum, size)
	return args.String(0), args.Error(1)
}

func (m *Mock) ListObjects(_ context.Context, namespace, container, versionID string) ([]string, error) {
	args := m.Called(namespace, container, versionID)
	return args.Get(0).([]string), args.Error(1)
}

func (m *Mock) ListObjectsByPage(_ context.Context, namespace, container, versionID string, pageNum uint64) (uint64, []string, error) {
	args := m.Called(namespace, container, versionID, pageNum)
	return args.Get(0).(uint64), args.Get(1).([]string), args.Error(2)
}

func (m *Mock) GetObjectURL(ctx context.Context, namespace, container, versionID, key string) (string, error) {
	args := m.Called(namespace, container, versionID, key)
	return args.String(0), args.Error(1)
}

func (m *Mock) DeleteObject(_ context.Context, namespace, container, versionID, key string) error {
	args := m.Called(namespace, container, versionID, key)
	return args.Error(0)
}
