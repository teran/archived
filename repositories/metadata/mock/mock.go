package mock

import (
	"context"

	"github.com/stretchr/testify/mock"
)

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

func (m *Mock) ListAllVersionsByContainer(_ context.Context, container string) ([]string, error) {
	args := m.Called(container)
	return args.Get(0).([]string), args.Error(1)
}

func (m *Mock) ListPublishedVersionsByContainer(_ context.Context, container string) ([]string, error) {
	args := m.Called(container)
	return args.Get(0).([]string), args.Error(1)
}

func (m *Mock) MarkVersionPublished(_ context.Context, container, version string) error {
	args := m.Called(container, version)
	return args.Error(0)
}

func (m *Mock) CreateObject(_ context.Context, container, version, key, casKey string) error {
	args := m.Called(container, version, key, casKey)
	return args.Error(0)
}

func (m *Mock) ListObjects(_ context.Context, container, version string, offset, limit uint64) ([]string, error) {
	args := m.Called(container, version, offset, limit)
	return args.Get(0).([]string), args.Error(1)
}

func (m *Mock) DeleteObject(_ context.Context, container, version, key string) error {
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
