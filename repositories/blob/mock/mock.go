package mock

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/teran/archived/repositories/blob"
)

var _ blob.Repository = (*Mock)(nil)

type Mock struct {
	mock.Mock
}

func New() *Mock {
	return &Mock{}
}

func (m *Mock) PutBlobURL(_ context.Context, key string) (string, error) {
	args := m.Called(key)
	return args.String(0), args.Error(1)
}

func (m *Mock) GetBlobURL(_ context.Context, key, mimeType, filename string) (string, error) {
	args := m.Called(key, mimeType, filename)
	return args.String(0), args.Error(1)
}
