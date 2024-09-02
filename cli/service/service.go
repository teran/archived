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
	CreateNamespace(namespaceName string) func(ctx context.Context) error
	RenameNamespace(oldName, newName string) func(ctx context.Context) error
	ListNamespaces() func(ctx context.Context) error
	DeleteNamespace(namespaceName string) func(ctx context.Context) error

	CreateContainer(namespaceName, containerName string) func(ctx context.Context) error
	MoveContainer(namespaceName, containerName, destinationNamespace string) func(ctx context.Context) error
	RenameContainer(namespaceName, oldName, newName string) func(ctx context.Context) error
	ListContainers(namespaceName string) func(ctx context.Context) error
	DeleteContainer(namespaceName, containerName string) func(ctx context.Context) error

	CreateVersion(namespaceName, containerName string, shouldPublish bool, fromDir, fromYumRepo, rpmGPGKey, rpmGPGKeyChecksum *string) func(ctx context.Context) error
	DeleteVersion(namespaceName, containerName, versionID string) func(ctx context.Context) error
	ListVersions(namespaceName, containerName string) func(ctx context.Context) error
	PublishVersion(namespaceName, containerName, versionID string) func(ctx context.Context) error

	CreateObject(namespaceName, containerName, versionID, directoryPath string) func(ctx context.Context) error
	ListObjects(namespaceName, containerName, versionID string) func(ctx context.Context) error
	GetObjectURL(namespaceName, containerName, versionID, objectKey string) func(ctx context.Context) error
	DeleteObject(namespaceName, containerName, versionID, objectKey string) func(ctx context.Context) error
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

func (s *service) CreateNamespace(namespaceName string) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		_, err := s.cli.CreateNamespace(ctx, &v1proto.CreateNamespaceRequest{
			Name: namespaceName,
		})
		if err != nil {
			return errors.Wrap(err, "error creating namespace")
		}
		fmt.Printf("namespace `%s` created\n", namespaceName)
		return nil
	}
}

func (s *service) RenameNamespace(oldName, newName string) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		_, err := s.cli.RenameNamespace(ctx, &v1proto.RenameNamespaceRequest{
			OldName: oldName,
			NewName: newName,
		})
		if err != nil {
			return errors.Wrap(err, "error renaming namespace")
		}
		fmt.Printf("namespace `%s` renamed to `%s`\n", oldName, newName)
		return nil
	}
}

func (s *service) ListNamespaces() func(ctx context.Context) error {
	return func(ctx context.Context) error {
		resp, err := s.cli.ListNamespaces(ctx, &v1proto.ListNamespacesRequest{})
		if err != nil {
			return errors.Wrap(err, "error listing namespaces")
		}

		for _, namespace := range resp.GetName() {
			fmt.Println(namespace)
		}
		return nil
	}
}

func (s *service) DeleteNamespace(namespaceName string) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		_, err := s.cli.DeleteNamespace(ctx, &v1proto.DeleteNamespaceRequest{
			Name: namespaceName,
		})
		if err != nil {
			return errors.Wrap(err, "error deleting namespace")
		}

		fmt.Printf("namespace `%s` has been deleted\n", namespaceName)
		return nil
	}
}

func (s *service) CreateContainer(namespaceName, containerName string) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		_, err := s.cli.CreateContainer(ctx, &v1proto.CreateContainerRequest{
			Namespace: namespaceName,
			Name:      containerName,
		})
		if err != nil {
			return errors.Wrap(err, "error creating container")
		}
		fmt.Printf("container `%s` created\n", containerName)
		return nil
	}
}

func (s *service) MoveContainer(namespaceName, containerName, destinationNamespace string) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		_, err := s.cli.MoveContainer(ctx, &v1proto.MoveContainerRequest{
			Namespace:            namespaceName,
			ContainerName:        containerName,
			DestinationNamespace: destinationNamespace,
		})
		if err != nil {
			return errors.Wrap(err, "error moving container")
		}
		fmt.Printf("container `%s` just moved from `%s` to `%s`\n", containerName, namespaceName, destinationNamespace)
		return nil
	}
}

func (s *service) RenameContainer(namespaceName, oldName, newName string) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		_, err := s.cli.RenameContainer(ctx, &v1proto.RenameContainerRequest{
			Namespace: namespaceName,
			OldName:   oldName,
			NewName:   newName,
		})
		if err != nil {
			return errors.Wrap(err, "error renaming container")
		}
		fmt.Printf("container `%s` renamed to `%s`\n", oldName, newName)
		return nil
	}
}

func (s *service) ListContainers(namespaceName string) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		resp, err := s.cli.ListContainers(ctx, &v1proto.ListContainersRequest{
			Namespace: namespaceName,
		})
		if err != nil {
			return errors.Wrap(err, "error listing containers")
		}

		for _, container := range resp.GetName() {
			fmt.Println(container)
		}
		return nil
	}
}

func (s *service) DeleteContainer(namespaceName, containerName string) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		_, err := s.cli.DeleteContainer(ctx, &v1proto.DeleteContainerRequest{
			Namespace: namespaceName,
			Name:      containerName,
		})
		if err != nil {
			return errors.Wrap(err, "error deleting container")
		}

		fmt.Printf("container `%s` has been deleted\n", containerName)
		return nil
	}
}

func (s *service) CreateVersion(namespaceName, containerName string, shouldPublish bool, fromDir, fromYumRepo, rpmGPGKey, rpmGPGKeyChecksum *string) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		resp, err := s.cli.CreateVersion(ctx, &v1proto.CreateVersionRequest{
			Namespace: namespaceName,
			Container: containerName,
		})
		if err != nil {
			return errors.Wrap(err, "error creating version")
		}

		versionID := resp.GetVersion()

		if fromDir != nil && *fromDir != "" {
			log.Tracef("--from-dir is requested with `%s`", *fromDir)
			err = s.CreateObject(namespaceName, containerName, versionID, *fromDir)(ctx)
			if err != nil {
				return errors.Wrap(err, "error creating objects")
			}
		} else if fromYumRepo != nil && *fromYumRepo != "" {
			log.Tracef("--from-yum-repo is requested with `%s`", *fromYumRepo)
			var gpgKeyring openpgp.EntityList = nil
			if *rpmGPGKey != "" {
				gpgKeyring, err = getGPGKey(ctx, *rpmGPGKey, rpmGPGKeyChecksum)
				if err != nil {
					return err
				}
			}

			err := s.createVersionFromYUMRepository(ctx, namespaceName, containerName, versionID, *fromYumRepo, gpgKeyring)
			if err != nil {
				return errors.Wrap(err, "error creating objects")
			}
		}

		if shouldPublish {
			_, err = s.cli.PublishVersion(ctx, &v1proto.PublishVersionRequest{
				Namespace: namespaceName,
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

func (s *service) createVersionFromYUMRepository(ctx context.Context, namespaceName string, containerName, versionID, url string, gpgKeyring openpgp.EntityList) error {
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

		log.Tracef("rpc CreateObject(%s, %s, %s, %s, %s, %d)", namespaceName, containerName, versionID, k, checksum, size)
		resp, err := s.cli.CreateObject(ctx, &v1proto.CreateObjectRequest{
			Namespace: namespaceName,
			Container: containerName,
			Version:   versionID,
			Key:       k,
			Checksum:  checksum,
			Size:      uint64(size),
		})
		if err != nil {
			return errors.Wrap(err, "error creating object")
		}

		if uploadURL := resp.GetUploadUrl(); uploadURL != "" {
			err := uploadBlob(ctx, uploadURL, bytes.NewReader(v), uint64(size))
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

			log.Tracef("rpc CreateObject(%s, %s, %s, %s, %s, %d)", namespaceName, containerName, versionID, name, checksum, size)
			resp, err := s.cli.CreateObject(ctx, &v1proto.CreateObjectRequest{
				Namespace: namespaceName,
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
						return errors.Wrap(err, "error getting reader for package file")
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
		}(pkg.Name, pkg.Checksum, strings.TrimSuffix(url, "/")+"/"+strings.TrimPrefix(pkg.Name, "/"), pkg.Size)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *service) DeleteVersion(namespaceName, containerName, versionID string) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		_, err := s.cli.DeleteVersion(ctx, &v1proto.DeleteVersionRequest{
			Namespace: namespaceName,
			Container: containerName,
			Version:   versionID,
		})
		if err != nil {
			return errors.Wrap(err, "error deleting version")
		}

		fmt.Printf("version `%s` of container `%s/%s` has been deleted\n", versionID, namespaceName, containerName)

		return nil
	}
}

func (s *service) ListVersions(namespaceName, containerName string) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		resp, err := s.cli.ListVersions(ctx, &v1proto.ListVersionsRequest{
			Namespace: namespaceName,
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

func (s *service) PublishVersion(namespaceName, containerName, versionID string) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		_, err := s.cli.PublishVersion(ctx, &v1proto.PublishVersionRequest{
			Namespace: namespaceName,
			Container: containerName,
			Version:   versionID,
		})
		if err != nil {
			return errors.Wrap(err, "error publishing version")
		}

		fmt.Printf("version `%s` of container `%s/%s` is published now\n", namespaceName, containerName, versionID)
		return nil
	}
}

func (s *service) CreateObject(namespaceName, containerName, versionID, directoryPath string) func(ctx context.Context) error {
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

			log.Tracef("rpc CreateObject(%s,%s, %s, %s, %s, %d)", namespaceName, containerName, versionID, shortPath, checksum, size)
			resp, err := s.cli.CreateObject(ctx, &v1proto.CreateObjectRequest{
				Namespace: namespaceName,
				Container: containerName,
				Version:   versionID,
				Key:       shortPath,
				Checksum:  checksum,
				Size:      uint64(size),
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

				if err := uploadBlob(ctx, url, fp, uint64(size)); err != nil {
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

func (s *service) ListObjects(namespaceName, containerName, versionID string) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		resp, err := s.cli.ListObjects(ctx, &v1proto.ListObjectsRequest{
			Namespace: namespaceName,
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

func (s *service) GetObjectURL(namespaceName, containerName, versionID, objectKey string) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		url, err := s.cli.GetObjectURL(ctx, &v1proto.GetObjectURLRequest{
			Namespace: namespaceName,
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

func (s *service) DeleteObject(namespaceName, containerName, versionID, objectKey string) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		_, err := s.cli.DeleteObject(ctx, &v1proto.DeleteObjectRequest{
			Namespace: namespaceName,
			Container: containerName,
			Version:   versionID,
			Key:       objectKey,
		})
		if err != nil {
			return errors.Wrap(err, "error deleting object")
		}

		log.Printf("Object `%s` (%s/%s/%s) deleted", objectKey, namespaceName, containerName, versionID)
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

func uploadBlob(ctx context.Context, url string, rd io.Reader, size uint64) error {
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

		req.ContentLength = int64(size)
	}

	log.WithFields(log.Fields{
		"length": req.ContentLength,
	}).Tracef("running HTTP PUT request ...")

	uploadResp, err := http.DefaultClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "error uploading file")
	}
	defer uploadResp.Body.Close()

	log.Debugf("upload HTTP response code: %s", uploadResp.Status)

	if uploadResp.StatusCode > 299 {
		return errors.Errorf("unexpected status code on upload: %s", uploadResp.Status)
	}

	return nil
}
