package yum

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/pkg/errors"
	"github.com/sassoftware/go-rpmutils"
	log "github.com/sirupsen/logrus"

	"github.com/teran/archived/cli/lazyblob"
	"github.com/teran/archived/cli/service/source"
	yum "github.com/teran/archived/cli/service/source/yum/yum_repo"
)

const processStatusInterval = 100

var (
	_ source.Source = (*repository)(nil)

	ErrFileNotFound = errors.New("file not found")
)

type repository struct {
	repo            yum.YumRepo
	repoURL         string
	rpmGPGKeyURL    *string
	rpmGPGKeySHA256 *string
}

func New(repoURL string, rpmGPGKeyURL, rpmGPGKeySHA256 *string) source.Source {
	log.WithFields(log.Fields{
		"url":            repoURL,
		"gpg_key_url":    rpmGPGKeyURL,
		"gpg_key_sha256": rpmGPGKeySHA256,
	}).Trace("initializing YUM source ...")

	return &repository{
		repo:            yum.New(repoURL),
		repoURL:         repoURL,
		rpmGPGKeyURL:    rpmGPGKeyURL,
		rpmGPGKeySHA256: rpmGPGKeySHA256,
	}
}

func (r *repository) Process(ctx context.Context, handler source.ObjectHandler) error {
	log.WithFields(log.Fields{
		"repository_url": r.repoURL,
	}).Info("running creating version from YUM repository ...")

	packages, err := r.repo.Packages(ctx)
	if err != nil {
		return errors.Wrap(err, "error getting repository data")
	}

	var gpgKeyring openpgp.EntityList = nil

	if r.rpmGPGKeyURL != nil && *r.rpmGPGKeyURL != "" {
		log.Tracef("RPM GPG Key was passed so initialing GPG keyring ...")
		gpgKeyring, err = getGPGKey(ctx, *r.rpmGPGKeyURL, r.rpmGPGKeySHA256)
		if err != nil {
			return err
		}
	}

	log.WithFields(log.Fields{
		"repository_url": r.repoURL,
	}).Info("handling YUM repository metadata files ...")
	for k, v := range r.repo.Metadata() {
		size := len(v)

		hasher := sha256.New()
		n, err := hasher.Write(v)
		if err != nil {
			return err
		}

		if n != size {
			return io.ErrShortWrite
		}

		checksum := hex.EncodeToString(hasher.Sum(nil))
		mimeType := http.DetectContentType(v)

		log.Tracef("handler(%s, %s, %d)", k, checksum, size)
		if err := handler(ctx, source.Object{
			Path: k,
			Contents: func(ctx context.Context) (io.Reader, error) {
				return bytes.NewReader(v), nil
			},
			SHA256:   checksum,
			Size:     uint64(size),
			MimeType: mimeType,
		}); err != nil {
			return errors.Wrap(err, "error calling object handler")
		}
	}

	log.WithFields(log.Fields{
		"repository_url": r.repoURL,
		"packages_count": len(packages),
	}).Info("handling package files ...")

	for cnt, pkg := range packages {
		log.WithFields(log.Fields{
			"name":     pkg.Name,
			"checksum": pkg.Checksum,
			"path":     strings.TrimSuffix(r.repoURL, "/") + "/" + strings.TrimPrefix(pkg.Name, "/"),
			"size":     pkg.Size,
		}).Trace("processing package ...")

		err := func(name, checksum, sourceURL string, size uint64) error {
			lb := lazyblob.New(sourceURL, os.TempDir(), size)
			defer func() {
				if err := lb.Close(); err != nil {
					log.Warnf("error removing scratch data: %s", err)
				}
			}()

			if pkg.ChecksumType != "sha256" {
				filename, err := lb.Filename(ctx)
				if err != nil {
					return errors.Wrap(err, "error getting package filename")
				}

				checksum, err = checksumFile(filename)
				if err != nil {
					return errors.Wrap(err, "error calculating checksum")
				}
			}

			if len(gpgKeyring) > 0 {
				fp, err := lb.Reader(ctx)
				if err != nil {
					return errors.Wrap(err, "error getting object reader")
				}

				_, sigs, err := rpmutils.Verify(fp, gpgKeyring)
				if err != nil {
					return errors.Wrapf(err, "error verifying package signature: %s", name)
				}

				if len(sigs) == 0 {
					log.Warnf("package `%s` does not contain signature but verification is requested", name)
				}
			}

			if err := handler(ctx, source.Object{
				Path:     name,
				Contents: lb.Reader,
				SHA256:   checksum,
				Size:     size,
				MimeType: detectMimeType(name),
			}); err != nil {
				return errors.Wrap(err, "error calling object handler")
			}

			if cnt%processStatusInterval == 0 {
				log.WithFields(log.Fields{
					"repository_url": r.repoURL,
				}).Infof("%d files processed ...", cnt+1)
			}

			return nil
		}(pkg.Name, pkg.Checksum, strings.TrimSuffix(r.repoURL, "/")+"/"+strings.TrimPrefix(pkg.Name, "/"), pkg.Size)
		if err != nil {
			return err
		}
	}
	return nil
}

func detectMimeType(filename string) string {
	switch path.Ext(filename) {
	case ".gz":
		return "application/gzip"
	case ".xz":
		return "application/x-xz"
	case ".xml":
		return "application/xml"
	case ".rpm":
		return "application/x-rpm"
	}
	return "application/octet-stream"
}
