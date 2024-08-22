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
	Reader(ctx context.Context) (io.Reader, error)
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
		return errors.Wrap(err, "error creating request object")
	}

	req.Header.Set("User-Agent", userAgent)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "error performing HTTP request")
	}
	defer resp.Body.Close()

	if resp.StatusCode > 299 {
		return errors.Errorf("%s: unexpected HTTP response status: %s", l.url, resp.Status)
	}

	if err := os.MkdirAll(l.tempDir, 0o700); err != nil {
		return errors.Wrap(err, "error creating directory structure")
	}

	l.tempFile, err = os.CreateTemp(l.tempDir, "package_*.rpm.tmp")
	if err != nil {
		return errors.Wrap(err, "error creating temporary file")
	}

	log.WithFields(log.Fields{
		"filename": l.tempFile.Name(),
	}).Debug("temporary file created")

	n, err := io.Copy(l.tempFile, resp.Body)
	if err != nil {
		return errors.Wrap(err, "error writing data")
	}

	log.WithFields(log.Fields{
		"filename": l.tempFile.Name(),
		"length":   n,
	}).Trace("bytes copied")

	if n != l.length {
		return errors.Wrap(io.ErrShortWrite, "data length copied mismatch")
	}

	return nil
}

func (l *lazyblob) newReader() (io.Reader, error) {
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

func (l *lazyblob) Reader(ctx context.Context) (io.Reader, error) {
	if l.tempFile == nil {
		if err := l.download(ctx); err != nil {
			return nil, errors.Wrap(err, "error downloading file")
		}
	}

	return l.newReader()
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
		if err := os.Remove(l.tempFile.Name()); err != nil {
			log.WithFields(log.Fields{
				"filename": l.tempFile.Name(),
			}).Debug("error removing temporary file")
		}

		log.Tracef("tempfile is suppose to be alive: closing ...")
		if err := l.tempFile.Close(); err != nil && !errors.Is(err, os.ErrClosed) {
			return err
		}

		l.tempFile = nil
	}

	return nil
}
