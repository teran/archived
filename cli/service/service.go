package service

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/teran/archived/cli/service/source"
	cache "github.com/teran/archived/cli/service/stat_cache"
	v1proto "github.com/teran/archived/manager/presenter/grpc/proto/v1"
	ptr "github.com/teran/go-ptr"
)

type Service interface {
	CreateNamespace(namespaceName string) func(ctx context.Context) error
	RenameNamespace(oldName, newName string) func(ctx context.Context) error
	ListNamespaces() func(ctx context.Context) error
	DeleteNamespace(namespaceName string) func(ctx context.Context) error

	CreateContainer(namespaceName, containerName string, ttl time.Duration) func(ctx context.Context) error
	MoveContainer(namespaceName, containerName, destinationNamespace string) func(ctx context.Context) error
	RenameContainer(namespaceName, oldName, newName string) func(ctx context.Context) error
	ListContainers(namespaceName string) func(ctx context.Context) error
	DeleteContainer(namespaceName, containerName string) func(ctx context.Context) error
	SetContainerParameters(namespaceName, containerName string, ttl time.Duration) func(ctx context.Context) error

	CreateVersion(namespaceName, containerName string, shouldPublish bool, src source.Source) func(ctx context.Context) error
	DeleteVersion(namespaceName, containerName, versionID string) func(ctx context.Context) error
	ListVersions(namespaceName, containerName string) func(ctx context.Context) error
	PublishVersion(namespaceName, containerName, versionID string) func(ctx context.Context) error

	ListObjects(namespaceName, containerName, versionID string) func(ctx context.Context) error
	GetObjectURL(namespaceName, containerName, versionID, objectKey string) func(ctx context.Context) error
	DeleteObject(namespaceName, containerName, versionID, objectKey string) func(ctx context.Context) error
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

func (s *service) CreateContainer(namespaceName, containerName string, ttl time.Duration) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		_, err := s.cli.CreateContainer(ctx, &v1proto.CreateContainerRequest{
			Namespace:  namespaceName,
			Name:       containerName,
			TtlSeconds: ptr.Int64(int64(ttl.Seconds())),
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

func (s *service) SetContainerParameters(namespaceName, containerName string, ttl time.Duration) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		_, err := s.cli.SetContainerParameters(ctx, &v1proto.SetContainerParametersRequest{
			Namespace:  namespaceName,
			Name:       containerName,
			TtlSeconds: ptr.Int64(int64(ttl.Seconds())),
		})
		if err != nil {
			return errors.Wrap(err, "error setting container versions TTL")
		}

		fmt.Printf("container `%s` versions TTL set to %s\n", containerName, ttl)
		return nil
	}
}

func (s *service) CreateVersion(namespaceName, containerName string, shouldPublish bool, src source.Source) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		log.Tracef("creating version ...")
		resp, err := s.cli.CreateVersion(ctx, &v1proto.CreateVersionRequest{
			Namespace: namespaceName,
			Container: containerName,
		})
		if err != nil {
			return errors.Wrap(err, "error creating version")
		}

		versionID := resp.GetVersion()
		log.Tracef("version created: `%s`", versionID)

		if err := src.Process(ctx, func(ctx context.Context, obj source.Object) error {
			return s.createObject(ctx, namespaceName, containerName, versionID, obj)
		}); err != nil {
			return errors.Wrap(err, "error processing source")
		}

		if shouldPublish {
			log.Tracef("publishing is requested so publishing version ...")
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

func (s *service) createObject(ctx context.Context, namespaceName, containerName, versionID string, object source.Object) error {
	log.WithFields(log.Fields{
		"path":      object.Path,
		"sha256":    object.SHA256,
		"length":    object.Size,
		"mime_type": object.MimeType,
	}).Debug("creating object ...")

	resp, err := s.cli.CreateObject(ctx, &v1proto.CreateObjectRequest{
		Namespace: namespaceName,
		Container: containerName,
		Version:   versionID,
		Key:       object.Path,
		Checksum:  object.SHA256,
		Size:      object.Size,
		MimeType:  object.MimeType,
	})
	if err != nil {
		return errors.Wrap(err, "error creating object")
	}

	if url := resp.GetUploadUrl(); url != "" {
		log.WithFields(log.Fields{
			"path":   object.Path,
			"sha256": object.SHA256,
			"length": object.Size,
			"url":    url,
		}).Info("uploading BLOB ...")

		rd, err := object.Contents(ctx)
		if err != nil {
			return err
		}

		if err := uploadBlob(ctx, url, rd, object.Size); err != nil {
			return err
		}
	}
	return nil
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
		}).Tracef("Setting Content-Length ...")

		req.ContentLength = int64(size)
	}

	log.WithFields(log.Fields{
		"length": req.ContentLength,
	}).Tracef("running HTTP PUT request ...")

	uploadResp, err := http.DefaultClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "error uploading file")
	}
	defer func() { _ = uploadResp.Body.Close() }()

	log.Debugf("upload HTTP response code: %s", uploadResp.Status)

	if uploadResp.StatusCode > 299 {
		return errors.Errorf("unexpected status code on upload: %s", uploadResp.Status)
	}

	return nil
}
