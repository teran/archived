package mock

import (
	"context"
	"io"

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

func (m *Mock) PutBlob(_ context.Context, key string, rd io.ReadSeeker) error {
	data, err := io.ReadAll(rd)
	if err != nil {
		return err
	}

	args := m.Called(key, data)
	return args.Error(0)
}

func (m *Mock) GetBlobURL(_ context.Context, key string) (string, error) {
	args := m.Called(key)
	return args.String(0), args.Error(1)
}
