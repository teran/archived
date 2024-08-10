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

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	cache "github.com/teran/archived/cli/service/stat_cache"
	v1proto "github.com/teran/archived/presenter/manager/grpc/proto/v1"
)

type Service interface {
	CreateContainer(containerName string) func(ctx context.Context) error
	ListContainers() func(ctx context.Context) error
	DeleteContainer(containerName string) func(ctx context.Context) error

	CreateVersion(containerName string, shouldPublish bool, fromDir *string) func(ctx context.Context) error
	DeleteVersion(containerName, versionID string) func(ctx context.Context) error
	ListVersions(containerName string) func(ctx context.Context) error
	PublishVersion(containerName, versionID string) func(ctx context.Context) error

	CreateObject(containerName, versionID, directoryPath string) func(ctx context.Context) error
	ListObjects(containerName, versionID string) func(ctx context.Context) error
	GetObjectURL(containerName, versionID, objectKey string) func(ctx context.Context) error
	DeleteObject(containerName, versionID, objectKey string) func(ctx context.Context) error
}

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

func (s *service) CreateVersion(containerName string, shouldPublish bool, fromDir *string) func(ctx context.Context) error {
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
		log.Debugf("running filepath.Walk on `%s`", directoryPath)
		return filepath.Walk(directoryPath, func(path string, info fs.FileInfo, err error) error {
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

				buf := bytes.NewBuffer(nil)
				if _, err := io.Copy(buf, fp); err != nil {
					return errors.Wrap(err, "error on data copy")
				}

				req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, buf)
				if err != nil {
					return errors.Wrap(err, "error constructing request")
				}

				req.Header.Set("Content-Type", "multipart/form-data")

				c := &http.Client{}
				uploadResp, err := c.Do(req)
				if err != nil {
					return errors.Wrap(err, "error uploading file")
				}
				log.Debugf("upload HTTP response code: %s", uploadResp.Status)
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
