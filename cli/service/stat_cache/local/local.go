package local

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"strings"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	cache "github.com/teran/archived/cli/service/stat_cache"
)

var _ cache.CacheRepository = (*local)(nil)

type local struct {
	cacheDir string
}

func New(cacheDir string) (cache.CacheRepository, error) {
	if err := os.MkdirAll(cacheDir, 0o750); err != nil {
		return nil, errors.Wrap(err, "error creating cache directory")
	}

	return &local{
		cacheDir: cacheDir,
	}, nil
}

func (l *local) Put(ctx context.Context, filename string, info fs.FileInfo, value string) error {
	cacheKey := cacheKeyFunc(filename, info)

	fullCacheObjectPath := path.Join(l.cacheDir, cacheKey)

	if err := os.MkdirAll(path.Dir(fullCacheObjectPath), 0o755); err != nil {
		return errors.Wrap(err, "error preparing cache path")
	}

	fp, err := os.Create(fullCacheObjectPath)
	if err != nil {
		return errors.Wrap(err, "error creating cache file")
	}
	defer func() { _ = fp.Close() }()

	n, err := fp.Write([]byte(value))
	if err != nil {
		return errors.Wrap(err, "error writing value into cache file")
	}

	if n != len(value) {
		return io.ErrShortWrite
	}

	return nil
}

func (l *local) Get(ctx context.Context, filename string, info fs.FileInfo) (string, error) {
	cacheKey := cacheKeyFunc(filename, info)

	data, err := os.ReadFile(path.Join(l.cacheDir, cacheKey))
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", errors.Wrap(err, "error reading cache file")
	}

	return string(data), nil
}

func cacheKeyFunc(filename string, info fs.FileInfo) string {
	h := sha256.New()

	cacheKey := fmt.Sprintf(
		"%s:%s:%d:%d",
		filename,
		info.ModTime().Format(time.RFC3339Nano),
		info.Size(),
		info.Mode(),
	)

	log.WithFields(log.Fields{
		"key":      cacheKey,
		"filename": filename,
	}).Tracef("generating cache key")

	_, err := h.Write([]byte(cacheKey))
	if err != nil {
		panic(errors.Wrap(err, "error writing into hasher buffer. memory corruption?"))
	}

	key := hex.EncodeToString(h.Sum(nil))

	return path.Join(strings.Split(key, "")...)
}
