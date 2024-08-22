package memcache

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	memcacheCli "github.com/bradfitz/gomemcache/memcache"
	log "github.com/sirupsen/logrus"

	emodels "github.com/teran/archived/exporter/models"
	"github.com/teran/archived/models"
	"github.com/teran/archived/repositories/metadata"
)

var _ metadata.Repository = (*memcache)(nil)

type memcache struct {
	cli       *memcacheCli.Client
	keyPrefix string
	repo      metadata.Repository
	ttl       time.Duration
}

func New(cli *memcacheCli.Client, repo metadata.Repository, ttl time.Duration, keyPrefix string) metadata.Repository {
	if keyPrefix == "" {
		keyPrefix = "_"
	}

	return &memcache{
		cli:       cli,
		keyPrefix: keyPrefix,
		repo:      repo,
		ttl:       ttl,
	}
}

func (m *memcache) CreateNamespace(ctx context.Context, name string) error {
	return m.repo.CreateNamespace(ctx, name)
}

func (m *memcache) RenameNamespace(ctx context.Context, oldName, newName string) error {
	return m.repo.RenameNamespace(ctx, oldName, newName)
}

func (m *memcache) ListNamespaces(ctx context.Context) ([]string, error) {
	cacheKey := strings.Join([]string{
		m.keyPrefix,
		"ListNamespaces",
	}, ":")

	item, err := m.cli.Get(cacheKey)
	if err != nil {
		if err == memcacheCli.ErrCacheMiss {
			log.WithFields(log.Fields{
				"key": cacheKey,
			}).Tracef("cache miss")

			namespaces, err := m.repo.ListNamespaces(ctx)
			if err != nil {
				return nil, err
			}

			if err := store(m, cacheKey, namespaces); err != nil {
				return nil, err
			}

			return namespaces, nil
		}

		return nil, err
	}
	log.WithFields(log.Fields{
		"key": cacheKey,
	}).Tracef("cache hit")

	var retrievedValue []string
	err = json.Unmarshal(item.Value, &retrievedValue)
	if err != nil {
		return nil, err
	}

	return retrievedValue, nil
}

func (m *memcache) DeleteNamespace(ctx context.Context, name string) error {
	return m.repo.DeleteNamespace(ctx, name)
}

func (m *memcache) CreateContainer(ctx context.Context, namespace, name string) error {
	return m.repo.CreateContainer(ctx, namespace, name)
}

func (m *memcache) RenameContainer(ctx context.Context, namespace, oldName, newName string) error {
	return m.repo.RenameContainer(ctx, namespace, oldName, newName)
}

func (m *memcache) ListContainers(ctx context.Context, namespace string) ([]string, error) {
	cacheKey := strings.Join([]string{
		m.keyPrefix,
		"ListContainers",
		namespace,
	}, ":")

	item, err := m.cli.Get(cacheKey)
	if err != nil {
		if err == memcacheCli.ErrCacheMiss {
			log.WithFields(log.Fields{
				"key": cacheKey,
			}).Tracef("cache miss")

			containers, err := m.repo.ListContainers(ctx, namespace)
			if err != nil {
				return nil, err
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
		return nil, err
	}

	return retrievedValue, nil
}

func (m *memcache) DeleteContainer(ctx context.Context, namespace, name string) error {
	return m.repo.DeleteContainer(ctx, namespace, name)
}

func (m *memcache) CreateVersion(ctx context.Context, namespace, container string) (string, error) {
	return m.repo.CreateVersion(ctx, namespace, container)
}

func (m *memcache) GetLatestPublishedVersionByContainer(ctx context.Context, namespace, container string) (string, error) {
	cacheKey := strings.Join([]string{
		m.keyPrefix,
		"GetLatestPublishedVersionByContainer",
		namespace,
		container,
	}, ":")

	item, err := m.cli.Get(cacheKey)
	if err != nil {
		if err == memcacheCli.ErrCacheMiss {
			log.WithFields(log.Fields{
				"key": cacheKey,
			}).Tracef("cache miss")

			version, err := m.repo.GetLatestPublishedVersionByContainer(ctx, namespace, container)
			if err != nil {
				return "", err
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
		return "", err
	}

	return retrievedValue, nil
}

func (m *memcache) ListAllVersionsByContainer(ctx context.Context, namespace, container string) ([]models.Version, error) {
	cacheKey := strings.Join([]string{
		m.keyPrefix,
		"ListAllVersionsByContainer",
		namespace,
		container,
	}, ":")

	item, err := m.cli.Get(cacheKey)
	if err != nil {
		if err == memcacheCli.ErrCacheMiss {
			log.WithFields(log.Fields{
				"key": cacheKey,
			}).Tracef("cache miss")

			versions, err := m.repo.ListAllVersionsByContainer(ctx, namespace, container)
			if err != nil {
				return nil, err
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
		return nil, err
	}

	return retrievedValue, nil
}

func (m *memcache) ListPublishedVersionsByContainer(ctx context.Context, namespace, container string) ([]models.Version, error) {
	cacheKey := strings.Join([]string{
		m.keyPrefix,
		"ListPublishedVersionsByContainer",
		namespace,
		container,
	}, ":")

	item, err := m.cli.Get(cacheKey)
	if err != nil {
		if err == memcacheCli.ErrCacheMiss {
			log.WithFields(log.Fields{
				"key": cacheKey,
			}).Tracef("cache miss")

			versions, err := m.repo.ListPublishedVersionsByContainer(ctx, namespace, container)
			if err != nil {
				return nil, err
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
		return nil, err
	}

	return retrievedValue, nil
}

func (m *memcache) ListPublishedVersionsByContainerAndPage(ctx context.Context, namespace, container string, offset, limit uint64) (uint64, []models.Version, error) {
	type proxy struct {
		Total    uint64
		Versions []models.Version
	}

	cacheKey := strings.Join([]string{
		m.keyPrefix,
		"ListPublishedVersionsByContainerAndPage",
		namespace,
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

			n, versions, err := m.repo.ListPublishedVersionsByContainerAndPage(ctx, namespace, container, offset, limit)
			if err != nil {
				return 0, nil, err
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
		return 0, nil, err
	}

	return retrievedValue.Total, retrievedValue.Versions, nil
}

func (m *memcache) ListUnpublishedVersionsByContainer(ctx context.Context, namespace, container string) ([]models.Version, error) {
	return m.repo.ListUnpublishedVersionsByContainer(ctx, namespace, container)
}

func (m *memcache) MarkVersionPublished(ctx context.Context, namespace, container, version string) error {
	return m.repo.MarkVersionPublished(ctx, namespace, container, version)
}

func (m *memcache) DeleteVersion(ctx context.Context, namespace, container, version string) error {
	return m.repo.DeleteVersion(ctx, namespace, container, version)
}

func (m *memcache) CreateObject(ctx context.Context, namespace, container, version, key, casKey string) error {
	return m.repo.CreateObject(ctx, namespace, container, version, key, casKey)
}

func (m *memcache) ListObjects(ctx context.Context, namespace, container, version string, offset, limit uint64) (uint64, []string, error) {
	type proxy struct {
		Total   uint64
		Objects []string
	}

	cacheKey := strings.Join([]string{
		m.keyPrefix,
		"ListObjects",
		namespace,
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

			n, objects, err := m.repo.ListObjects(ctx, namespace, container, version, offset, limit)
			if err != nil {
				return 0, nil, err
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
		return 0, nil, err
	}

	return retrievedValue.Total, retrievedValue.Objects, nil
}

func (m *memcache) DeleteObject(ctx context.Context, namespace, container, version string, key ...string) error {
	return m.repo.DeleteObject(ctx, namespace, container, version, key...)
}

func (m *memcache) RemapObject(ctx context.Context, namespace, container, version, key, newCASKey string) error {
	return m.repo.RemapObject(ctx, namespace, container, version, key, newCASKey)
}

func (m *memcache) CreateBLOB(ctx context.Context, checksum string, size uint64, mimeType string) error {
	return m.repo.CreateBLOB(ctx, checksum, size, mimeType)
}

func (m *memcache) GetBlobKeyByObject(ctx context.Context, namespace, container, version, key string) (string, error) {
	cacheKey := strings.Join([]string{
		m.keyPrefix,
		"GetBlobKeyByObject",
		namespace,
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

			key, err := m.repo.GetBlobKeyByObject(ctx, namespace, container, version, key)
			if err != nil {
				return "", err
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
		return "", err
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
