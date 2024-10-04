package mock

import (
	"context"

	"github.com/teran/archived/cli/service/source"
	"github.com/teran/archived/cli/service/stat_cache/mock"
)

var _ source.Source = (*Mock)(nil)

type Mock struct {
	mock.Mock
}

func New() *Mock {
	return &Mock{}
}

func (m *Mock) Process(ctx context.Context, handler func(ctx context.Context, obj source.Object) error) error {
	args := m.Called()
	return args.Error(0)
}
