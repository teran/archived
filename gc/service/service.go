package service

import (
	"context"

	"github.com/teran/archived/repositories/metadata"
)

type Service interface {
	Run(ctx context.Context) error
}

type service struct {
	repo metadata.Repository
}

func New(repo metadata.Repository) Service {
	return &service{
		repo: repo,
	}
}

func (s *service) Run(ctx context.Context) error {
	return nil
}
