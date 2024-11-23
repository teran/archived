package yum

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"strings"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/pkg/errors"
	"github.com/sassoftware/go-rpmutils"
	log "github.com/sirupsen/logrus"

	"github.com/teran/archived/cli/lazyblob"
	"github.com/teran/archived/cli/service/source"
	"github.com/teran/archived/cli/yum"
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

func (r *repository) Process(ctx context.Context, handler func(ctx context.Context, obj source.Object) error) error {
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

		log.Tracef("handler(%s, %s, %d)", k, checksum, size)
		if err := handler(ctx, source.Object{
			Path:     k,
			Contents: bytes.NewReader(v),
			SHA256:   checksum,
			Size:     uint64(size),
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

			fp, err := lb.Reader(ctx)
			if err != nil {
				return errors.Wrap(err, "error getting object reader")
			}

			_, sigs, err := rpmutils.Verify(fp, gpgKeyring)
			if err != nil {
				return errors.Wrapf(err, "error verifying package signature: %s", name)
			}

			if len(sigs) == 0 {
				log.Warnf("package `%s` does not contain signature", name)
			}

			fp, err = lb.Reader(ctx)
			if err != nil {
				return errors.Wrap(err, "error getting object reader")
			}

			if err := handler(ctx, source.Object{
				Path:     name,
				Contents: fp,
				SHA256:   checksum,
				Size:     size,
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
