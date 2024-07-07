package blob

import (
	"context"
	"io"
)

type Repository interface {
	PutBlob(ctx context.Context, key string, rd io.ReadSeeker) error
	GetBlobURL(ctx context.Context, key string) (string, error)
}
