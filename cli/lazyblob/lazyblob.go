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
	url      string
	tempDir  string
	length   int64
	tempFile *os.File

	mutex *sync.RWMutex
}

func New(url, tempDir string, length int64) LazyBLOB {
	log.WithFields(log.Fields{
		"url":     url,
		"length":  length,
		"tempdir": tempDir,
	}).Debug("lazyblob initialized")

	return &lazyblob{
		url:      url,
		tempDir:  tempDir,
		length:   length,
		tempFile: nil,
		mutex:    &sync.RWMutex{},
	}
}

func (l *lazyblob) download(ctx context.Context) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if l.tempFile != nil {
		log.WithFields(log.Fields{
			"filename": l.tempFile.Name(),
		}).Tracef("file is already downloaded. Skipping ...")
		return nil
	}

	log.WithFields(log.Fields{
		"url": l.url,
	}).Trace("downloading file ...")

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

	l.tempFile, err = os.CreateTemp(l.tempDir, "package_*.rpm.tmp")
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"filename": l.tempFile.Name(),
	}).Debug("temporary file created")

	n, err := io.Copy(l.tempFile, resp.Body)
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"filename": l.tempFile.Name(),
		"length":   n,
	}).Trace("bytes copied")

	if n != l.length {
		return io.ErrShortWrite
	}

	return nil
}

func (l *lazyblob) newReadCloser() (*os.File, error) {
	l.mutex.RLock()
	defer l.mutex.RUnlock()

	if l.tempFile == nil {
		return nil, errors.New("file is not downloaded yet")
	}

	log.WithFields(log.Fields{
		"filename": l.tempFile.Name(),
	}).Trace("file is downloaded, providing the file handler by request")

	_, err := l.tempFile.Seek(0, 0)
	return l.tempFile, err
}

func (l *lazyblob) File(ctx context.Context) (*os.File, error) {
	if l.tempFile == nil {
		if err := l.download(ctx); err != nil {
			return nil, err
		}
	}

	return l.newReadCloser()
}

func (l *lazyblob) Filename(ctx context.Context) (string, error) {
	if l.tempFile == nil {
		if err := l.download(ctx); err != nil {
			return "", err
		}
	}

	return l.tempFile.Name(), nil
}

func (l *lazyblob) URL() string {
	return l.url
}

func (l *lazyblob) Close() error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if l.tempFile != nil {
		err := l.tempFile.Close()
		if err != nil && !errors.Is(err, os.ErrClosed) {
			return err
		}
		l.tempFile = nil
		return nil
	}

	return os.ErrClosed
}