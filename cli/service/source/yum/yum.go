package yum

import (
	"context"

	"github.com/teran/archived/cli/service/source"
)

var _ source.Source = (*yum)(nil)

type yum struct {
	repoURL         string
	rpmGPGKeyURL    string
	rpmGPGKeySHA256 string
}

func New(repoURL, rpmGPGKeyURL, rpmGPGKeySHA256 string) source.Source {
	return &yum{
		repoURL:         repoURL,
		rpmGPGKeyURL:    rpmGPGKeyURL,
		rpmGPGKeySHA256: rpmGPGKeySHA256,
	}
}

func (y *yum) Process(ctx context.Context, handler func(ctx context.Context, obj source.Object) error) error {
	// FIXME
	return nil
}
