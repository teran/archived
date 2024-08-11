package memcache

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	memcacheCli "github.com/bradfitz/gomemcache/memcache"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	emodels "github.com/teran/archived/exporter/models"
	"github.com/teran/archived/models"
	"github.com/teran/archived/repositories/metadata"
)

var _ metadata.Repository = (*memcache)(nil)

type memcache struct {
	cli  *memcacheCli.Client
	repo metadata.Repository
	ttl  time.Duration
}

func New(cli *memcacheCli.Client, repo metadata.Repository, ttl time.Duration) metadata.Repository {
	return &memcache{
		cli:  cli,
		repo: repo,
		ttl:  ttl,
	}
}

func (m *memcache) CreateContainer(ctx context.Context, name string) error {
	return m.repo.CreateContainer(ctx, name)
}

func (m *memcache) ListContainers(ctx context.Context) ([]string, error) {
	cacheKey := "_ListContainers"
	item, err := m.cli.Get(cacheKey)
	if err != nil {
		if err == memcacheCli.ErrCacheMiss {
			log.WithFields(log.Fields{
				"key": cacheKey,
			}).Tracef("cache miss")

			containers, err := m.repo.ListContainers(ctx)
			if err != nil {
				return nil, errors.Wrap(err, "error retrieving container list")
			}

			if err := store(m, cacheKey, containers); err != nil {
				return nil, err
			}

			return containers, nil
		}

		return nil, err
	}
	log.WithFields(log.Fields{
		"key": cacheKey,
	}).Tracef("cache hit")

	var retrievedValue []string
	err = json.Unmarshal(item.Value, &retrievedValue)
	if err != nil {
		return nil, errors.Wrap(err, "error unmarshaling cached value")
	}

	return retrievedValue, nil
}

func (m *memcache) DeleteContainer(ctx context.Context, name string) error {
	return m.repo.DeleteContainer(ctx, name)
}

func (m *memcache) CreateVersion(ctx context.Context, container string) (string, error) {
	return m.repo.CreateVersion(ctx, container)
}

func (m *memcache) GetLatestPublishedVersionByContainer(ctx context.Context, container string) (string, error) {
	cacheKey := "_GetLatestPublishedVersionByContainer:" + container
	item, err := m.cli.Get(cacheKey)
	if err != nil {
		if err == memcacheCli.ErrCacheMiss {
			log.WithFields(log.Fields{
				"key": cacheKey,
			}).Tracef("cache miss")

			version, err := m.repo.GetLatestPublishedVersionByContainer(ctx, container)
			if err != nil {
				return "", errors.Wrapf(err, "error retrieving latest version for container `%s`", container)
			}

			if err := store(m, cacheKey, version); err != nil {
				return "", err
			}

			return version, nil
		}

		return "", err
	}
	log.WithFields(log.Fields{
		"key": cacheKey,
	}).Tracef("cache hit")

	var retrievedValue string
	err = json.Unmarshal(item.Value, &retrievedValue)
	if err != nil {
		return "", errors.Wrap(err, "error unmarshaling cached value")
	}

	return retrievedValue, nil
}

func (m *memcache) ListAllVersionsByContainer(ctx context.Context, container string) ([]models.Version, error) {
	cacheKey := "_ListAllVersionsByContainer:" + container
	item, err := m.cli.Get(cacheKey)
	if err != nil {
		if err == memcacheCli.ErrCacheMiss {
			log.WithFields(log.Fields{
				"key": cacheKey,
			}).Tracef("cache miss")

			versions, err := m.repo.ListAllVersionsByContainer(ctx, container)
			if err != nil {
				return nil, errors.Wrapf(err, "error retrieving latest version for container `%s`", container)
			}

			if err := store(m, cacheKey, versions); err != nil {
				return nil, err
			}

			return versions, nil
		}

		return nil, err
	}
	log.WithFields(log.Fields{
		"key": cacheKey,
	}).Tracef("cache hit")

	var retrievedValue []models.Version
	err = json.Unmarshal(item.Value, &retrievedValue)
	if err != nil {
		return nil, errors.Wrap(err, "error unmarshaling cached value")
	}

	return retrievedValue, nil
}

func (m *memcache) ListPublishedVersionsByContainer(ctx context.Context, container string) ([]models.Version, error) {
	cacheKey := "_ListPublishedVersionsByContainer:" + container
	item, err := m.cli.Get(cacheKey)
	if err != nil {
		if err == memcacheCli.ErrCacheMiss {
			log.WithFields(log.Fields{
				"key": cacheKey,
			}).Tracef("cache miss")

			versions, err := m.repo.ListPublishedVersionsByContainer(ctx, container)
			if err != nil {
				return nil, errors.Wrapf(err, "error retrieving latest version for container `%s`", container)
			}

			if err := store(m, cacheKey, versions); err != nil {
				return nil, err
			}

			return versions, nil
		}

		return nil, err
	}
	log.WithFields(log.Fields{
		"key": cacheKey,
	}).Tracef("cache hit")

	var retrievedValue []models.Version
	err = json.Unmarshal(item.Value, &retrievedValue)
	if err != nil {
		return nil, errors.Wrap(err, "error unmarshaling cached value")
	}

	return retrievedValue, nil
}

func (m *memcache) ListPublishedVersionsByContainerAndPage(ctx context.Context, container string, offset, limit uint64) (uint64, []models.Version, error) {
	type proxy struct {
		Total    uint64
		Versions []models.Version
	}

	cacheKey := strings.Join([]string{
		"_ListPublishedVersionsByContainerAndPage",
		container,
		strconv.FormatUint(offset, 10),
		strconv.FormatUint(limit, 10),
	}, ":")

	item, err := m.cli.Get(cacheKey)
	if err != nil {
		if err == memcacheCli.ErrCacheMiss {
			log.WithFields(log.Fields{
				"key": cacheKey,
			}).Tracef("cache miss")

			n, versions, err := m.repo.ListPublishedVersionsByContainerAndPage(ctx, container, offset, limit)
			if err != nil {
				return 0, nil, errors.Wrapf(err, "error retrieving published versions for container `%s` (offset=%d; limit=%d)", container, offset, limit)
			}

			if err := store(m, cacheKey, proxy{Total: n, Versions: versions}); err != nil {
				return 0, nil, err
			}

			return n, versions, nil
		}

		return 0, nil, nil
	}
	log.WithFields(log.Fields{
		"key": cacheKey,
	}).Tracef("cache hit")

	var retrievedValue proxy
	err = json.Unmarshal(item.Value, &retrievedValue)
	if err != nil {
		return 0, nil, errors.Wrap(err, "error unmarshaling cached value")
	}

	return retrievedValue.Total, retrievedValue.Versions, nil
}

func (m *memcache) ListUnpublishedVersionsByContainer(ctx context.Context, container string) ([]models.Version, error) {
	return m.repo.ListUnpublishedVersionsByContainer(ctx, container)
}

func (m *memcache) MarkVersionPublished(ctx context.Context, container, version string) error {
	return m.repo.MarkVersionPublished(ctx, container, version)
}

func (m *memcache) DeleteVersion(ctx context.Context, container, version string) error {
	return m.repo.DeleteVersion(ctx, container, version)
}

func (m *memcache) CreateObject(ctx context.Context, container, version, key, casKey string) error {
	return m.repo.CreateObject(ctx, container, version, key, casKey)
}

func (m *memcache) ListObjects(ctx context.Context, container, version string, offset, limit uint64) (uint64, []string, error) {
	type proxy struct {
		Total   uint64
		Objects []string
	}

	cacheKey := strings.Join([]string{
		"_ListObjects",
		container,
		version,
		strconv.FormatUint(offset, 10),
		strconv.FormatUint(limit, 10),
	}, ":")

	item, err := m.cli.Get(cacheKey)
	if err != nil {
		if err == memcacheCli.ErrCacheMiss {
			log.WithFields(log.Fields{
				"key": cacheKey,
			}).Tracef("cache miss")

			n, objects, err := m.repo.ListObjects(ctx, container, version, offset, limit)
			if err != nil {
				return 0, nil, errors.Wrapf(err, "error retrieving objects for `%s/%s`, offset=%d; limit=%d", container, version, offset, limit)
			}

			if err := store(m, cacheKey, proxy{Total: n, Objects: objects}); err != nil {
				return 0, nil, err
			}

			return n, objects, err
		}

		return 0, nil, nil
	}
	log.WithFields(log.Fields{
		"key": cacheKey,
	}).Tracef("cache hit")

	var retrievedValue proxy
	err = json.Unmarshal(item.Value, &retrievedValue)
	if err != nil {
		return 0, nil, errors.Wrap(err, "error unmarshaling cached value")
	}

	return retrievedValue.Total, retrievedValue.Objects, nil
}

func (m *memcache) DeleteObject(ctx context.Context, container, version string, key ...string) error {
	return m.repo.DeleteObject(ctx, container, version, key...)
}

func (m *memcache) RemapObject(ctx context.Context, container, version, key, newCASKey string) error {
	return m.repo.RemapObject(ctx, container, version, key, newCASKey)
}

func (m *memcache) CreateBLOB(ctx context.Context, checksum string, size uint64, mimeType string) error {
	return m.repo.CreateBLOB(ctx, checksum, size, mimeType)
}

func (m *memcache) GetBlobKeyByObject(ctx context.Context, container, version, key string) (string, error) {
	cacheKey := strings.Join([]string{
		"_GetBlobKeyByObject",
		container,
		version,
		key,
	}, ":")

	item, err := m.cli.Get(cacheKey)
	if err != nil {
		if err == memcacheCli.ErrCacheMiss {
			log.WithFields(log.Fields{
				"key": cacheKey,
			}).Tracef("cache miss")

			key, err := m.repo.GetBlobKeyByObject(ctx, container, version, key)
			if err != nil {
				return "", errors.Wrapf(err, "error retrieving object key for `%s/%s/%s`", container, version, key)
			}

			if err = store(m, cacheKey, key); err != nil {
				return "", err
			}

			return key, err
		}

		return "", err
	}
	log.WithFields(log.Fields{
		"key": cacheKey,
	}).Tracef("cache hit")

	var retrievedValue string
	err = json.Unmarshal(item.Value, &retrievedValue)
	if err != nil {
		return "", errors.Wrap(err, "error unmarshaling cached value")
	}

	return retrievedValue, nil
}

func (m *memcache) EnsureBlobKey(ctx context.Context, key string, size uint64) error {
	return m.repo.EnsureBlobKey(ctx, key, size)
}

func (m *memcache) CountStats(ctx context.Context) (*emodels.Stats, error) {
	return m.repo.CountStats(ctx)
}

func store[T any](m *memcache, key string, in T) error {
	cacheValue, err := json.Marshal(in)
	if err != nil {
		return err
	}

	return m.cli.Set(&memcacheCli.Item{
		Key:        key,
		Expiration: int32(m.ttl.Seconds()),
		Value:      cacheValue,
	})
}
