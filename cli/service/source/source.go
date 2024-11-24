package source

import (
	"context"
	"io"
)

type Object struct {
	Path     string
	Contents io.Reader
	SHA256   string
	Size     uint64
	MimeType string
}

type Source interface {
	Process(ctx context.Context, handler func(ctx context.Context, obj Object) error) error
}
