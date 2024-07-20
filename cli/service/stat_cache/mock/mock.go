package mock

import (
	"context"
	"io/fs"

	cache "github.com/teran/archived/cli/service/stat_cache"
	"github.com/teran/archived/repositories/blob/mock"
)

var _ cache.CacheRepository = (*Mock)(nil)

type Mock struct {
	mock.Mock
}

func New() *Mock {
	return &Mock{}
}

func (m *Mock) Put(_ context.Context, filename string, info fs.FileInfo, value string) error {
	args := m.Called(filename, value)
	return args.Error(0)
}

func (m *Mock) Get(_ context.Context, filename string, info fs.FileInfo) (string, error) {
	args := m.Called(filename)
	return args.String(0), args.Error(1)
}
