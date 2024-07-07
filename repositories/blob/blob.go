package blob

import (
	"context"
	"io"
)

type Repository interface {
	PutBlob(ctx context.Context, key string, rd io.Reader) error
	GetBlobURL(ctx context.Context, key string) (string, error)
}
