package service

import (
	"context"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const defaultLimit uint64 = 1000

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
	containers, err := s.cfg.MdRepo.ListContainers(ctx)
	if err != nil {
		return errors.Wrap(err, "error listing containers")
	}

	log.Infof("found %d containers", len(containers))

	for _, container := range containers {
		log.WithFields(log.Fields{
			"container": container,
		}).Debugf("listing unpublished versions ...")

		versions, err := s.cfg.MdRepo.ListUnpublishedVersionsByContainer(ctx, container)
		if err != nil {
			return errors.Wrapf(err, "error listing versions for container `%s`", container)
		}

		for _, version := range versions {
			log.WithFields(log.Fields{
				"container": container,
				"version":   version,
			}).Debugf("listing objects ...")

			var (
				total  uint64 = 0
				offset uint64 = 0
			)

			for {
				log.WithFields(log.Fields{
					"container": container,
					"version":   version,
					"offset":    offset,
					"limit":     defaultLimit,
				}).Tracef("list objects loop iteration ...")

				var objects []string = []string{}

				total, objects, err = s.cfg.MdRepo.ListObjects(ctx, container, version, offset, defaultLimit)
				if err != nil {
					return errors.Wrapf(err, "error listing objects for container `%s`; version `%s`", container, version)
				}

				if total == 0 {
					break
				}

				if !s.cfg.DryRun {
					log.WithFields(log.Fields{
						"container": container,
						"version":   version,
						"amount":    len(objects),
					}).Info("Performing actual metadata deletion: objects")

					err = s.cfg.MdRepo.DeleteObject(ctx, container, version, objects...)
					if err != nil {
						return errors.Wrapf(err, "error removing object from `%s/%s (%d objects)`", container, version, len(objects))
					}
				}
			}

			log.WithFields(log.Fields{
				"container": container,
				"version":   version,
			}).Debug("deleting version ...")

			if !s.cfg.DryRun {
				log.WithFields(log.Fields{
					"container": container,
					"version":   version,
				}).Info("Performing actual metadata deletion: version")

				err = s.cfg.MdRepo.DeleteVersion(ctx, container, version)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}
