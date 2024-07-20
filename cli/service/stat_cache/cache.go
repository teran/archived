package cache

import (
	"context"
	"io/fs"
)

type CacheRepository interface {
	Put(ctx context.Context, filename string, info fs.FileInfo, value string) error
	Get(ctx context.Context, filename string, info fs.FileInfo) (string, error)
}
