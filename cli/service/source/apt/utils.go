package apt

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"path/filepath"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/ulikunitz/xz"
	debian "pault.ag/go/debian/control"
)

var errFileNotFound = errors.New("file not found")

func fetchMetadata[T any](ctx context.Context, url string, v T) ([]byte, error) {
	rawData, err := getFile(ctx, url)
	if err != nil {
		return nil, errors.Wrap(err, "error getting file")
	}

	log.WithFields(log.Fields{
		"length": len(rawData),
	}).Trace("control structure received into buffer")

	var rd io.Reader
	switch filepath.Ext(url) {
	case ".gz":
		rd, err = gzip.NewReader(bytes.NewReader(rawData))
		if err != nil {
			return nil, errors.Wrap(err, "error constructing gzip reader")
		}
		defer func() { _ = rd.(*gzip.Reader).Close() }()
	case ".xz":
		rd, err = xz.NewReader(bytes.NewReader(rawData))
		if err != nil {
			return nil, errors.Wrap(err, "error constructing xz reader")
		}
	default:
		rd = bytes.NewReader(rawData)
	}

	if err := debian.Unmarshal(v, rd); err != nil {
		return nil, errors.Wrap(err, "error unmarshaling control structure")
	}

	return rawData, nil
}

func sha256FromBytes(in []byte) (string, error) {
	hasher := sha256.New()
	n, err := hasher.Write(in)
	if err != nil {
		return "", err
	}

	if n != len(in) {
		return "", io.ErrShortWrite
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func detectMimeTypeByFilename(path string) string {
	switch filepath.Ext(path) {
	case ".deb":
		return "application/vnd.debian.binary-package"
	case ".gz":
		return "application/x-gzip"
	case ".xz":
		return "application/x-xz"
	default:
		return "application/octet-stream"
	}
}

func getFile(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	c := &http.Client{}
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	switch resp.StatusCode {
	case http.StatusNotFound:
		return nil, errFileNotFound
	}

	return io.ReadAll(resp.Body)
}
