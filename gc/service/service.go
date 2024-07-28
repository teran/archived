package service

import (
	"context"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/teran/archived/repositories/metadata"
)

const defaultLimit uint64 = 1000

type Service interface {
	Run(ctx context.Context) error
}

type service struct {
	dryRun bool
	repo   metadata.Repository
}

func New(repo metadata.Repository, dryRun bool) Service {
	log.Infof("initializing garbage collection service with dryRun=%v ...", dryRun)
	return &service{
		dryRun: dryRun,
		repo:   repo,
	}
}

func (s *service) Run(ctx context.Context) error {
	log.Info("running garbage collection ...")
	containers, err := s.repo.ListContainers(ctx)
	if err != nil {
		return errors.Wrap(err, "error listing containers")
	}

	log.Infof("found %d containers", len(containers))

	for _, container := range containers {
		log.WithFields(log.Fields{
			"container": container,
		}).Debugf("listing unpublished versions ...")

		versions, err := s.repo.ListUnpublishedVersionsByContainer(ctx, container)
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

				total, objects, err = s.repo.ListObjects(ctx, container, version, offset, defaultLimit)
				if err != nil {
					return errors.Wrapf(err, "error listing objects for container `%s`; version `%s`", container, version)
				}

				for _, object := range objects {
					log.WithFields(log.Fields{
						"container": container,
						"version":   version,
						"object":    object,
					}).Debug("deleting object ...")

					if !s.dryRun {
						log.WithFields(log.Fields{
							"container": container,
							"version":   version,
							"object":    object,
						}).Info("Performing actual metadata deletion: object")

						err = s.repo.DeleteObject(ctx, container, version, object)
						if err != nil {
							return errors.Wrapf(err, "error removing object `%s/%s/%s`", container, version, object)
						}
					}
				}

				if total == 0 {
					break
				}
			}

			log.WithFields(log.Fields{
				"container": container,
				"version":   version,
			}).Debug("deleting version ...")

			if !s.dryRun {
				log.WithFields(log.Fields{
					"container": container,
					"version":   version,
				}).Info("Performing actual metadata deletion: version")

				err = s.repo.DeleteVersion(ctx, container, version)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}
