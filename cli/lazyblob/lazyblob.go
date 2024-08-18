package lazyblob

import (
	"context"
	"io"
	"net/http"
	"os"
	"sync"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const userAgent = "Mozilla/5.0 (compatible; archived-cli/lazyblob; +https://github.com/teran/archived)"

type LazyBLOB interface {
	URL() string
	File(ctx context.Context) (*os.File, error)
	Filename(ctx context.Context) (string, error)
	Close() error
}

type lazyblob struct {
	url          string
	tempDir      string
	length       int64
	mutex        *sync.RWMutex
	tempFilename string
}

func New(url, tempDir string, length int64) LazyBLOB {
	log.WithFields(log.Fields{
		"url":     url,
		"length":  length,
		"tempdir": tempDir,
	}).Debug("lazyblob initialized")

	return &lazyblob{
		url:     url,
		tempDir: tempDir,
		length:  length,
		mutex:   &sync.RWMutex{},
	}
}

func (l *lazyblob) download(ctx context.Context) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if l.tempFilename != "" {
		return nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, l.url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", userAgent)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err := os.MkdirAll(l.tempDir, 0o700); err != nil {
		return err
	}

	tempFile, err := os.CreateTemp(l.tempDir, "package_*.rpm.tmp")
	if err != nil {
		return err
	}
	defer tempFile.Close()

	n, err := io.Copy(tempFile, resp.Body)
	if err != nil {
		return err
	}

	if n != l.length {
		return io.ErrShortWrite
	}

	l.tempFilename = tempFile.Name()

	return nil
}

func (l *lazyblob) newReadCloser() (*os.File, error) {
	l.mutex.RLock()
	defer l.mutex.RUnlock()

	if l.tempFilename == "" {
		return nil, errors.New("file is not downloaded yet")
	}

	return os.Open(l.tempFilename)
}

func (l *lazyblob) File(ctx context.Context) (*os.File, error) {
	if l.tempFilename == "" {
		if err := l.download(ctx); err != nil {
			return nil, err
		}
	}

	return l.newReadCloser()
}

func (l *lazyblob) Filename(ctx context.Context) (string, error) {
	if l.tempFilename == "" {
		if err := l.download(ctx); err != nil {
			return "", err
		}
	}

	return l.tempFilename, nil
}

func (l *lazyblob) URL() string {
	return l.url
}

func (l *lazyblob) Close() error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if err := os.RemoveAll(l.tempDir); err != nil {
		return err
	}

	l.tempFilename = ""
	return nil
}
