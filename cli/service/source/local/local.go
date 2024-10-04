package local

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/teran/archived/cli/service/source"
	cache "github.com/teran/archived/cli/service/stat_cache"
)

const (
	sourceType = "local"

	processStatusInterval int = 100
)

var _ source.Source = (*local)(nil)

type local struct {
	dir   string
	cache cache.CacheRepository
}

func New(dir string, c cache.CacheRepository) source.Source {
	return &local{
		dir:   dir,
		cache: c,
	}
}

func (l *local) Process(ctx context.Context, handler func(ctx context.Context, obj source.Object) error) error {
	log.WithFields(log.Fields{
		"source_type": sourceType,
		"directory":   l.dir,
	}).Info("scanning directory ...")

	var cnt int

	return filepath.Walk(l.dir, func(path string, info fs.FileInfo, err error) error {
		defer func() { cnt++ }()

		if err != nil {
			return errors.Wrap(err, "walk: internal error")
		}

		if info.IsDir() {
			return nil
		}

		shortPath := strings.TrimPrefix(path, l.dir)
		shortPath = strings.TrimPrefix(shortPath, "/")
		size := info.Size()

		log.WithFields(log.Fields{
			"filename": shortPath,
			"size":     size,
		}).Debug("file found")

		log.WithFields(log.Fields{
			"filename": shortPath,
			"size":     size,
		}).Tracef("attempting to retrieve checksum from cache")

		checksum, err := l.cache.Get(ctx, path, info)
		if err != nil {
			return errors.Wrap(err, "error retrieving checksum from cache")
		}

		if checksum == "" {
			log.WithFields(log.Fields{
				"filename": shortPath,
				"size":     size,
			}).Debug("generating checksum")
			checksum, err = checksumFile(path)
			if err != nil {
				return errors.Wrap(err, "error calculating file checksum")
			}

			err := l.cache.Put(ctx, path, info, checksum)
			if err != nil {
				log.Warnf("error putting checksum calculation result into cache: %s", err)
			}
		}

		log.WithFields(log.Fields{
			"filename": shortPath,
			"size":     size,
			"checksum": checksum,
		}).Debug("checksum")

		fp, err := os.Open(path)
		if err != nil {
			return errors.Wrap(err, "error opening file")
		}
		defer fp.Close()

		if err := handler(ctx, source.Object{
			Path:     shortPath,
			Contents: fp,
			SHA256:   checksum,
			Size:     uint64(size),
		}); err != nil {
			return err
		}

		if cnt%processStatusInterval == 0 {
			log.WithFields(log.Fields{
				"directory": l.dir,
			}).Infof("%d files processed ...", cnt+1)
		}

		return nil
	})
}

func checksumFile(filename string) (string, error) {
	info, err := os.Stat(filename)
	if err != nil {
		return "", errors.Wrap(err, "error performing stat on file")
	}
	fp, err := os.Open(filename)
	if err != nil {
		return "", errors.Wrap(err, "error opening file")
	}
	defer fp.Close()

	h := sha256.New()
	n, err := io.Copy(h, fp)
	if err != nil {
		return "", errors.Wrap(err, "error reading file")
	}

	if n != info.Size() {
		return "", errors.Errorf("file size is %d bytes while only %d was copied: early EOF", info.Size(), n)
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}
