package service

import (
	"context"
	"io"

	"github.com/stretchr/testify/mock"
)

var _ ManageService = (*Mock)(nil)

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
	return args.Get(0).(string), args.Error(1)
}

func (m *Mock) ListVersions(_ context.Context, container string) ([]string, error) {
	args := m.Called(container)
	return args.Get(0).([]string), args.Error(1)
}

func (m *Mock) PublishVersion(_ context.Context, container, id string) error {
	args := m.Called(container, id)
	return args.Error(0)
}

func (m *Mock) DeleteVersion(_ context.Context, container, id string) error {
	args := m.Called(container, id)
	return args.Error(0)
}

func (m *Mock) AddObject(_ context.Context, container, versionID, key string, objReader io.Reader) error {
	data, err := io.ReadAll(objReader)
	if err != nil {
		return err
	}

	args := m.Called(container, versionID, key, data)
	return args.Error(0)
}

func (m *Mock) ListObjects(_ context.Context, container, versionID string) ([]string, error) {
	args := m.Called(container, versionID)
	return args.Get(0).([]string), args.Error(1)
}

func (m *Mock) GetObjectURL(ctx context.Context, container, versionID, key string) (string, error) {
	args := m.Called(container, versionID, key)
	return args.String(0), args.Error(1)
}

func (m *Mock) DeleteObject(_ context.Context, container, versionID, key string) error {
	args := m.Called(container)
	return args.Error(0)
}