package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/pkg/errors"
	rpmutils "github.com/sassoftware/go-rpmutils"
	log "github.com/sirupsen/logrus"

	"github.com/teran/archived/cli/lazyblob"
	cache "github.com/teran/archived/cli/service/stat_cache"
	"github.com/teran/archived/cli/yum"
	v1proto "github.com/teran/archived/manager/presenter/grpc/proto/v1"
)

type Service interface {
	CreateContainer(containerName string) func(ctx context.Context) error
	RenameContainer(oldName, newName string) func(ctx context.Context) error
	ListContainers() func(ctx context.Context) error
	DeleteContainer(containerName string) func(ctx context.Context) error

	CreateVersion(containerName string, shouldPublish bool, fromDir, fromYumRepo, rpmGPGKey *string) func(ctx context.Context) error
	DeleteVersion(containerName, versionID string) func(ctx context.Context) error
	ListVersions(containerName string) func(ctx context.Context) error
	PublishVersion(containerName, versionID string) func(ctx context.Context) error

	CreateObject(containerName, versionID, directoryPath string) func(ctx context.Context) error
	ListObjects(containerName, versionID string) func(ctx context.Context) error
	GetObjectURL(containerName, versionID, objectKey string) func(ctx context.Context) error
	DeleteObject(containerName, versionID, objectKey string) func(ctx context.Context) error
}

const processStatusInterval int = 100

type service struct {
	cache cache.CacheRepository
	cli   v1proto.ManageServiceClient
}

func New(cli v1proto.ManageServiceClient, cacheRepo cache.CacheRepository) Service {
	return &service{
		cache: cacheRepo,
		cli:   cli,
	}
}

func (s *service) CreateContainer(containerName string) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		_, err := s.cli.CreateContainer(ctx, &v1proto.CreateContainerRequest{
			Name: containerName,
		})
		if err != nil {
			return errors.Wrap(err, "error creating container")
		}
		fmt.Printf("container `%s` created\n", containerName)
		return nil
	}
}

func (s *service) RenameContainer(oldName, newName string) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		_, err := s.cli.RenameContainer(ctx, &v1proto.RenameContainerRequest{
			OldName: oldName,
			NewName: newName,
		})
		if err != nil {
			return errors.Wrap(err, "error renaming container")
		}
		fmt.Printf("container `%s` renamed to `%s`\n", oldName, newName)
		return nil
	}
}

func (s *service) ListContainers() func(ctx context.Context) error {
	return func(ctx context.Context) error {
		resp, err := s.cli.ListContainers(ctx, &v1proto.ListContainersRequest{})
		if err != nil {
			return errors.Wrap(err, "error listing containers")
		}

		for _, container := range resp.GetName() {
			fmt.Println(container)
		}
		return nil
	}
}

func (s *service) DeleteContainer(containerName string) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		_, err := s.cli.DeleteContainer(ctx, &v1proto.DeleteContainerRequest{
			Name: containerName,
		})
		if err != nil {
			return errors.Wrap(err, "error deleting container")
		}

		fmt.Printf("container `%s` has been deleted\n", containerName)
		return nil
	}
}

func (s *service) CreateVersion(containerName string, shouldPublish bool, fromDir, fromYumRepo, rpmGPGKey *string) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		resp, err := s.cli.CreateVersion(ctx, &v1proto.CreateVersionRequest{
			Container: containerName,
		})
		if err != nil {
			return errors.Wrap(err, "error creating version")
		}

		versionID := resp.GetVersion()

		if fromDir != nil && *fromDir != "" {
			log.Tracef("--from-dir is requested with `%s`", *fromDir)
			err = s.CreateObject(containerName, versionID, *fromDir)(ctx)
			if err != nil {
				return errors.Wrap(err, "error creating objects")
			}
		} else if fromYumRepo != nil && *fromYumRepo != "" {
			log.Tracef("--from-yum-repo is requested with `%s`", *fromYumRepo)
			var gpgKeyring openpgp.EntityList = nil
			if *rpmGPGKey != "" {
				gpgKeyring, err = getGPGKey(ctx, *rpmGPGKey)
				if err != nil {
					return err
				}
			}

			err := s.createVersionFromYUMRepository(ctx, containerName, versionID, *fromYumRepo, gpgKeyring)
			if err != nil {
				return errors.Wrap(err, "error creating objects")
			}
		}

		if shouldPublish {
			_, err = s.cli.PublishVersion(ctx, &v1proto.PublishVersionRequest{
				Container: containerName,
				Version:   versionID,
			})
			if err != nil {
				return errors.Wrap(err, "error publishing version")
			}

			fmt.Printf("version `%s` created and published\n", versionID)
		} else {
			fmt.Printf("version `%s` created unpublished\n", versionID)
		}

		return nil
	}
}

func (s *service) createVersionFromYUMRepository(ctx context.Context, containerName, versionID, url string, gpgKeyring openpgp.EntityList) error {
	log.WithFields(log.Fields{
		"repository_url": url,
	}).Info("running creating version from YUM repository ...")

	repo := yum.New(url)
	packages, err := repo.Packages(ctx)
	if err != nil {
		return errors.Wrap(err, "error getting repository data")
	}

	log.WithFields(log.Fields{
		"repository_url": url,
	}).Info("handling YUM repository metadata files ...")
	for k, v := range repo.Metadata() {
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

		log.Tracef("rpc CreateObject(%s, %s, %s, %s, %d)", containerName, versionID, k, checksum, size)
		resp, err := s.cli.CreateObject(ctx, &v1proto.CreateObjectRequest{
			Container: containerName,
			Version:   versionID,
			Key:       k,
			Checksum:  checksum,
			Size:      int64(size),
		})
		if err != nil {
			return errors.Wrap(err, "error creating object")
		}

		if uploadURL := resp.GetUploadUrl(); uploadURL != "" {
			err := uploadBlob(ctx, uploadURL, bytes.NewReader(v), int64(size))
			if err != nil {
				return err
			}
		}
	}

	log.WithFields(log.Fields{
		"repository_url": url,
		"packages_count": len(packages),
	}).Info("handling package files ...")
	for cnt, pkg := range packages {
		err := func(name, checksum, sourceURL string, size int64) error {
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

			log.Tracef("rpc CreateObject(%s, %s, %s, %s, %d)", containerName, versionID, name, checksum, size)
			resp, err := s.cli.CreateObject(ctx, &v1proto.CreateObjectRequest{
				Container: containerName,
				Version:   versionID,
				Key:       name,
				Checksum:  checksum,
				Size:      size,
			})
			if err != nil {
				return errors.Wrap(err, "error creating object")
			}

			if uploadURL := resp.GetUploadUrl(); uploadURL != "" {
				if gpgKeyring != nil {
					log.Debug("verifying RPM GPG signature ...")

					fp, err := lb.Reader(ctx)
					if err != nil {
						return errors.Wrap(err, "error opening package file")
					}

					_, sigs, err := rpmutils.Verify(fp, gpgKeyring)
					if err != nil {
						return errors.Wrapf(err, "error verifying package signature: %s", name)
					}

					if len(sigs) == 0 {
						log.Warnf("package `%s` does not contain signature", name)
					}
				}

				err := func(url, uploadURL string) error {
					log.Tracef("Upload URL: `%s`", uploadURL)

					fp, err := lb.Reader(ctx)
					if err != nil {
						return errors.Wrap(err, "error opening package file")
					}

					return uploadBlob(ctx, uploadURL, fp, size)
				}(url, uploadURL)
				if err != nil {
					return err
				}
			}

			if cnt%processStatusInterval == 0 {
				log.WithFields(log.Fields{
					"repository_url": url,
				}).Infof("%d files processed ...", cnt+1)
			}

			return nil
		}(pkg.Name, pkg.Checksum, strings.TrimSuffix(url, "/")+"/"+strings.TrimPrefix(pkg.Name, "/"), int64(pkg.Size))
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *service) DeleteVersion(containerName, versionID string) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		_, err := s.cli.DeleteVersion(ctx, &v1proto.DeleteVersionRequest{
			Container: containerName,
			Version:   versionID,
		})
		if err != nil {
			return errors.Wrap(err, "error deleting version")
		}

		fmt.Printf("version `%s` of container `%s` has been deleted\n", versionID, containerName)

		return nil
	}
}

func (s *service) ListVersions(containerName string) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		resp, err := s.cli.ListVersions(ctx, &v1proto.ListVersionsRequest{
			Container: containerName,
		})
		if err != nil {
			return errors.Wrap(err, "error listing versions")
		}

		for _, version := range resp.GetVersions() {
			fmt.Println(version)
		}

		return nil
	}
}

func (s *service) PublishVersion(containerName, versionID string) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		_, err := s.cli.PublishVersion(ctx, &v1proto.PublishVersionRequest{
			Container: containerName,
			Version:   versionID,
		})
		if err != nil {
			return errors.Wrap(err, "error publishing version")
		}

		fmt.Printf("version `%s` of container `%s` is published now\n", containerName, versionID)
		return nil
	}
}

func (s *service) CreateObject(containerName, versionID, directoryPath string) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		log.WithFields(log.Fields{
			"directory": directoryPath,
		}).Info("scanning directory ...")
		var cnt int
		return filepath.Walk(directoryPath, func(path string, info fs.FileInfo, err error) error {
			defer func() { cnt++ }()

			if err != nil {
				return errors.Wrap(err, "walk: internal error")
			}

			if info.IsDir() {
				return nil
			}

			shortPath := strings.TrimPrefix(path, directoryPath)
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
			checksum, err := s.cache.Get(ctx, path, info)
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

				err := s.cache.Put(ctx, path, info, checksum)
				if err != nil {
					log.Warnf("error putting checksum calculation result into cache: %s", err)
				}
			}
			log.WithFields(log.Fields{
				"filename": shortPath,
				"size":     size,
				"checksum": checksum,
			}).Debug("checksum")

			log.Tracef("rpc CreateObject(%s, %s, %s, %s, %d)", containerName, versionID, shortPath, checksum, size)
			resp, err := s.cli.CreateObject(ctx, &v1proto.CreateObjectRequest{
				Container: containerName,
				Version:   versionID,
				Key:       shortPath,
				Checksum:  checksum,
				Size:      size,
			})
			if err != nil {
				return errors.Wrap(err, "error creating object")
			}

			if url := resp.GetUploadUrl(); url != "" {
				log.Tracef("Upload URL: `%s`", url)

				fp, err := os.Open(path)
				if err != nil {
					return errors.Wrap(err, "error opening file")
				}
				defer fp.Close()

				if err := uploadBlob(ctx, url, fp, size); err != nil {
					return err
				}
			}

			if cnt%processStatusInterval == 0 {
				log.WithFields(log.Fields{
					"directory": directoryPath,
				}).Infof("%d files processed ...", cnt+1)
			}

			return nil
		})
	}
}

func (s *service) ListObjects(containerName, versionID string) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		resp, err := s.cli.ListObjects(ctx, &v1proto.ListObjectsRequest{
			Container: containerName,
			Version:   versionID,
		})
		if err != nil {
			return errors.Wrap(err, "error listing objects")
		}

		for _, object := range resp.GetObjects() {
			fmt.Println(object)
		}

		return nil
	}
}

func (s *service) GetObjectURL(containerName, versionID, objectKey string) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		url, err := s.cli.GetObjectURL(ctx, &v1proto.GetObjectURLRequest{
			Container: containerName,
			Version:   versionID,
			Key:       objectKey,
		})
		if err != nil {
			return errors.Wrap(err, "error getting object URL")
		}

		log.Printf("Object URL received: %s", url)
		return nil
	}
}

func (s *service) DeleteObject(containerName, versionID, objectKey string) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		_, err := s.cli.DeleteObject(ctx, &v1proto.DeleteObjectRequest{
			Container: containerName,
			Version:   versionID,
			Key:       objectKey,
		})
		if err != nil {
			return errors.Wrap(err, "error deleting object")
		}

		log.Printf("Object `%s` (%s/%s) deleted", objectKey, containerName, versionID)
		return nil
	}
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

func uploadBlob(ctx context.Context, url string, rd io.Reader, size int64) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, io.NopCloser(rd))
	if err != nil {
		return errors.Wrap(err, "error constructing request")
	}

	req.Header.Set("Content-Type", "multipart/form-data")
	if req.ContentLength == 0 || req.ContentLength == -1 {
		log.WithFields(log.Fields{
			"url":    url,
			"length": size,
		}).Tracef("size is set")

		req.ContentLength = size
	}

	log.WithFields(log.Fields{
		"length": req.ContentLength,
	}).Tracef("running HTTP PUT request ...")

	uploadResp, err := http.DefaultClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "error uploading file")
	}

	log.Debugf("upload HTTP response code: %s", uploadResp.Status)

	if uploadResp.StatusCode > 299 {
		return errors.Errorf("unexpected status code on upload: %s", uploadResp.Status)
	}

	return nil
}
