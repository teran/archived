package source

import (
	"context"
	"io"
)

type Object struct {
	Path     string
	Contents func(ctx context.Context) (io.Reader, error)
	SHA256   string
	Size     uint64
	MimeType string
}

type ObjectHandler func(ctx context.Context, obj Object) error

type Source interface {
	Process(ctx context.Context, handler ObjectHandler) error
}
