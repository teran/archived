package service

import (
	"context"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type Service interface {
	Run(ctx context.Context) error
}

type service struct {
	cfg *Config
}

func New(cfg *Config) (Service, error) {
	if err := cfg.Validate(); err != nil {
		return nil, errors.Wrap(err, "error validation gc service configuration")
	}

	log.Infof("initializing gc service ...")

	return &service{
		cfg: cfg,
	}, nil
}

func (s *service) Run(ctx context.Context) error {
	log.Info("running garbage collection ...")

	log.Debug("Running expired versions collection ...")
	if err := s.deleteExpiredVersions(ctx, nil); err != nil {
		return errors.Wrap(err, "error deleting expired versions")
	}

	return nil
}

func (s *service) deleteExpiredVersions(ctx context.Context, isPublished *bool) error {
	if err := s.cfg.MdRepo.DeleteExpiredVersionsWithObjects(ctx, isPublished); err != nil {
		return errors.Wrap(err, "error calling repository")
	}

	return nil
}
