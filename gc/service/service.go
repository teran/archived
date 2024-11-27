package service

import (
	"context"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/teran/archived/models"
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

	log.Trace("listing namespaces ...")
	namespaces, err := s.cfg.MdRepo.ListNamespaces(ctx)
	if err != nil {
		return errors.Wrap(err, "error listing namespaces")
	}

	log.Infof("found %d namespaces", len(namespaces))

	for _, namespace := range namespaces {
		log.WithFields(log.Fields{
			"namespace": namespace,
		}).Debug("processing namespace ...")

		containers, err := s.cfg.MdRepo.ListContainers(ctx, namespace)
		if err != nil {
			return errors.Wrap(err, "error listing containers")
		}

		log.Infof("found %d containers", len(containers))

		now := s.cfg.TimeNowFunc().UTC()

		for _, container := range containers {
			log.WithFields(log.Fields{
				"namespace":   namespace,
				"container":   container.Name,
				"ttl_seconds": int64(container.VersionsTTL.Seconds()),
			}).Debug("processing container ...")

			if container.VersionsTTL < 0 {
				log.Debug("container TTL is set to infinite (< 0): handling unpublished versions ...")

				err = s.deleteExpiredUnpublishedVersions(ctx, now, namespace, container)
				if err != nil {
					return errors.Wrapf(err, "error deleting expired unpublished versions for container `%s/%s`", namespace, container.Name)
				}
			} else {
				log.Debug("container TTL is set to positive value: handling all versions ...")

				err = s.deleteExpiredVersions(ctx, now, namespace, container)
				if err != nil {
					return errors.Wrapf(err, "error deleting expired versions for container `%s/%s`", namespace, container.Name)
				}
			}
		}
	}
	return nil
}

func (s *service) deleteExpiredVersions(ctx context.Context, now time.Time, namespace string, container models.Container) error {
	log.WithFields(log.Fields{
		"namespace": namespace,
		"container": container,
	}).Debug("listing expired versions ...")

	versions, err := s.cfg.MdRepo.ListAllVersionsByContainer(ctx, namespace, container.Name)
	if err != nil {
		return errors.Wrapf(err, "error listing versions for container `%s/%s`", namespace, container)
	}

	for _, version := range versions {
		log.WithFields(log.Fields{
			"namespace":    namespace,
			"container":    container,
			"version":      version.Name,
			"is_published": version.IsPublished,
			"created_at":   version.CreatedAt.Format(time.RFC3339),
		}).Trace("handling version ...")

		if version.CreatedAt.After(now.Add(-1 * container.VersionsTTL)) {
			log.WithFields(log.Fields{
				"namespace":    namespace,
				"container":    container,
				"version":      version.Name,
				"is_published": version.IsPublished,
				"created_at":   version.CreatedAt.Format(time.RFC3339),
			}).Trace("version is within TTL. Skipping ...")
			continue
		}

		log.WithFields(log.Fields{
			"namespace":    namespace,
			"container":    container,
			"version":      version.Name,
			"is_published": version.IsPublished,
			"created_at":   version.CreatedAt.Format(time.RFC3339),
		}).Debug("version is older then TTL. Deleting ...")

		err = s.deleteVersion(ctx, namespace, container, version)
		if err != nil {
			return errors.Wrapf(err, "error deleting version `%s` for container `%s/%s`", version.Name, namespace, container.Name)
		}
	}

	return nil
}

func (s *service) deleteExpiredUnpublishedVersions(ctx context.Context, now time.Time, namespace string, container models.Container) error {
	log.WithFields(log.Fields{
		"namespace": namespace,
		"container": container,
	}).Debug("listing expired unpublished versions ...")

	versions, err := s.cfg.MdRepo.ListUnpublishedVersionsByContainer(ctx, namespace, container.Name)
	if err != nil {
		return errors.Wrapf(err, "error listing versions for container `%s/%s`", namespace, container)
	}

	for _, version := range versions {
		log.WithFields(log.Fields{
			"namespace":    namespace,
			"container":    container,
			"version":      version.Name,
			"is_published": version.IsPublished,
			"created_at":   version.CreatedAt.Format(time.RFC3339),
		}).Trace("handling version ...")

		if version.CreatedAt.After(now.Add(-1 * s.cfg.UnpublishedVersionMaxAge)) {
			log.WithFields(log.Fields{
				"namespace":    namespace,
				"container":    container,
				"version":      version.Name,
				"is_published": version.IsPublished,
				"created_at":   version.CreatedAt.Format(time.RFC3339),
			}).Trace("unpublished version is within TTL. Skipping ...")
			continue
		}

		log.WithFields(log.Fields{
			"namespace":    namespace,
			"container":    container,
			"version":      version.Name,
			"is_published": version.IsPublished,
			"created_at":   version.CreatedAt.Format(time.RFC3339),
		}).Debug("unpublished version is older then TTL. Deleting ...")

		err = s.deleteVersion(ctx, namespace, container, version)
		if err != nil {
			return errors.Wrapf(err, "error deleting version `%s` for container `%s/%s`", version.Name, namespace, container.Name)
		}
	}

	return nil
}

func (s *service) deleteVersion(ctx context.Context, namespace string, container models.Container, version models.Version) error {
	var (
		total  uint64
		offset uint64
		err    error
	)

	for {
		log.WithFields(log.Fields{
			"namespace": namespace,
			"container": container,
			"version":   version,
			"offset":    offset,
			"limit":     defaultLimit,
		}).Tracef("list objects loop iteration ...")

		var objects []string
		total, objects, err = s.cfg.MdRepo.ListObjects(ctx, namespace, container.Name, version.Name, offset, defaultLimit)
		if err != nil {
			return errors.Wrapf(err, "error listing objects for container `%s/%s`; version `%s`", namespace, container, version.Name)
		}

		if total == 0 {
			break
		}

		if !s.cfg.DryRun {
			log.WithFields(log.Fields{
				"namespace": namespace,
				"container": container,
				"version":   version.Name,
				"amount":    len(objects),
			}).Info("Performing actual metadata deletion: objects")

			err = s.cfg.MdRepo.DeleteObject(ctx, namespace, container.Name, version.Name, objects...)
			if err != nil {
				return errors.Wrapf(err, "error removing object from `%s/%s/%s (%d objects)`", namespace, container, version.Name, len(objects))
			}
		}
	}

	log.WithFields(log.Fields{
		"namespace": namespace,
		"container": container,
		"version":   version.Name,
	}).Debug("deleting version ...")

	if !s.cfg.DryRun {
		log.WithFields(log.Fields{
			"namespace": namespace,
			"container": container,
			"version":   version.Name,
		}).Info("Performing actual metadata deletion: version")

		err = s.cfg.MdRepo.DeleteVersion(ctx, namespace, container.Name, version.Name)
		if err != nil {
			return err
		}
	}

	return nil
}
