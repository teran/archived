package router

import (
	"context"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestRouter(t *testing.T) {
	r := require.New(t)

	var (
		test1Called = 0
		test2Called = 0
		errTest1    = errors.New("blah")
	)

	rt := New(context.Background())
	rt.Register("test1", func(ctx context.Context) error {
		test1Called++
		return errTest1
	})
	rt.Register("test2", func(ctx context.Context) error {
		test2Called++
		return nil
	})

	err := rt.Call("test1")
	r.Error(err)
	r.Equal(errTest1, err)
	r.Equal(1, test1Called)

	err = rt.Call("test2")
	r.NoError(err)
	r.Equal(1, test2Called)
}
