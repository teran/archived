package blob

import (
	"context"
)

type Repository interface {
	PutBlobURL(ctx context.Context, key string) (string, error)
	GetBlobURL(ctx context.Context, key string) (string, error)
}
