package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	kingpin "github.com/alecthomas/kingpin/v2"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/teran/archived/cli/router"
	v1proto "github.com/teran/archived/presenter/manage/grpc/proto/v1"
)

var (
	appVersion     = "n/a (dev build)"
	buildTimestamp = "undefined"

	app = kingpin.New("archived-cli", "CLI interface for archived")

	debug = app.
		Flag("debug", "Enable debug mode").
		Short('d').
		Envar("ARCHIVED_CLI_DEBUG").
		Bool()

	trace = app.
		Flag("trace", "Enable trace mode (debug mode on steroids)").
		Short('t').
		Envar("ARCHIVED_CLI_TRACE").
		Bool()

	manageEndpoint = app.
			Flag("endpoint", "Manage API endpoint address").
			Short('s').
			Envar("ARCHIVED_CLI_ENDPOINT").
			Required().
			String()

	insecureFlag = app.Flag("insecure", "Do not use TLS for gRPC connection").
			Envar("ARCHIVED_CLI_INSECURE").
			Default("false").
			Bool()

	container           = app.Command("container", "container operations")
	containerCreate     = container.Command("create", "create new container")
	containerCreateName = containerCreate.Arg("name", "name of the container to create").Required().String()

	containerDelete     = container.Command("delete", "delete the given container")
	containerDeleteName = containerDelete.Arg("name", "name of the container to delete").Required().String()

	containerList = container.Command("list", "list containers")

	version                = app.Command("version", "version operations")
	versionCreate          = version.Command("create", "create new version for given container")
	versionCreateContainer = versionCreate.Arg("container", "name of the container to create version for").Required().String()

	versionDelete          = version.Command("delete", "delete the given version")
	versionDeleteContainer = versionDelete.Arg("container", "name of the container to delete version of").Required().String()
	versionDeleteVersion   = versionDelete.Arg("version", "version to delete").Required().String()

	versionList          = version.Command("list", "list versions for the given container")
	versionListContainer = versionList.Arg("container", "name of the container to list versions for").Required().String()

	versionPublish          = version.Command("publish", "publish the given version")
	versionPublishContainer = versionPublish.Arg("container", "name of the container to publish version for").Required().String()
	versionPublishVersion   = versionPublish.Arg("version", "version to publish").Required().String()

	object              = app.Command("object", "object operations")
	objectList          = object.Command("list", "list objects in the given container and version")
	objectListContainer = objectList.Arg("container", "name of the container to list objects from").Required().String()
	objectListVersion   = objectList.Arg("version", "version to list objects from").Required().String()

	objectCreate          = object.Command("create", "create object(s) from location")
	objectCreateContainer = objectCreate.Arg("container", "name of the container to publish object from").Required().String()
	objectCreateVersion   = objectCreate.Arg("version", "version to publish object from").Required().String()
	objectCreatePath      = objectCreate.Arg("path", "local path of the object to create").Required().String()

	objectURL          = object.Command("url", "get URL for the object")
	objectURLContainer = objectURL.Arg("container", "name of the container to publish object from").Required().String()
	objectURLVersion   = objectURL.Arg("version", "version to publish object from").Required().String()
	objectURLKey       = objectURL.Arg("key", "key of the object to publish").Required().String()

	deleteObject          = object.Command("delete", "delete object")
	deleteObjectContainer = deleteObject.Arg("container", "name of the container to delete objects from").Required().String()
	deleteObjectVersion   = deleteObject.Arg("version", "version to delete object from").Required().String()
	deleteObjectKey       = deleteObject.Arg("key", "key of the object to delete").Required().String()
)

func main() {
	ctx := context.Background()
	appCmd := kingpin.MustParse(app.Parse(os.Args[1:]))

	if *trace {
		log.SetLevel(log.TraceLevel)
		log.SetFormatter(&log.TextFormatter{
			FullTimestamp: true,
		})
		log.Trace("Trace mode is enabled. Beware of verbosity!")
	} else if *debug {
		log.SetLevel(log.DebugLevel)
		log.SetFormatter(&log.TextFormatter{
			FullTimestamp: true,
		})
		log.Debug("Debug mode is enabled.")
	}

	log.Debugf("Initializing gRPC client ...")

	grpcOpts := []grpc.DialOption{}
	if *insecureFlag {
		grpcOpts = append(grpcOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		grpcOpts = append(grpcOpts, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})))
	}

	dial, err := grpc.NewClient(*manageEndpoint, grpcOpts...)
	if err != nil {
		panic(err)
	}

	log.Debugf("Initializing manage service client ...")
	cli := v1proto.NewManageServiceClient(dial)

	r := router.New(ctx)
	r.Register(containerCreate.FullCommand(), func(ctx context.Context) error {
		_, err := cli.CreateContainer(ctx, &v1proto.CreateContainerRequest{
			Name: *containerCreateName,
		})
		if err != nil {
			return err
		}
		fmt.Printf("container `%s` created\n", *containerCreateName)
		return nil
	})
	r.Register(containerList.FullCommand(), func(ctx context.Context) error {
		resp, err := cli.ListContainers(ctx, &v1proto.ListContainersRequest{})
		if err != nil {
			return err
		}

		for _, container := range resp.GetName() {
			fmt.Println(container)
		}
		return nil
	})
	r.Register(versionList.FullCommand(), func(ctx context.Context) error {
		resp, err := cli.ListVersions(ctx, &v1proto.ListVersionsRequest{
			Container: *versionListContainer,
		})
		if err != nil {
			return err
		}

		for _, version := range resp.GetVersions() {
			fmt.Println(version)
		}

		return nil
	})
	r.Register(versionCreate.FullCommand(), func(ctx context.Context) error {
		resp, err := cli.CreateVersion(ctx, &v1proto.CreateVersionRequest{
			Container: *versionCreateContainer,
		})
		if err != nil {
			return err
		}

		fmt.Printf("version `%s` created unpublished\n", resp.GetVersion())
		return nil
	})
	r.Register(versionPublish.FullCommand(), func(ctx context.Context) error {
		_, err := cli.PublishVersion(ctx, &v1proto.PublishVersionRequest{
			Container: *versionPublishContainer,
			Version:   *versionPublishVersion,
		})
		if err != nil {
			return err
		}

		fmt.Printf("version `%s` if container `%s` is published now\n", *versionPublishVersion, *versionPublishContainer)
		return nil
	})
	r.Register(objectCreate.FullCommand(), func(ctx context.Context) error {
		return filepath.Walk(*objectCreatePath, func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			shortPath := strings.TrimPrefix(path, *objectCreatePath)
			log.Debugf("Found: %s\n", shortPath)

			size := info.Size()
			checksum, err := checksumFile(path)
			if err != nil {
				return err
			}

			resp, err := cli.CreateObject(ctx, &v1proto.CreateObjectRequest{
				Container: *objectCreateContainer,
				Version:   *objectCreateVersion,
				Key:       shortPath,
				Checksum:  checksum,
				Size:      size,
			})
			if err != nil {
				return err
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
					return err
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
	})
	r.Register(objectList.FullCommand(), func(ctx context.Context) error {
		resp, err := cli.ListObjects(ctx, &v1proto.ListObjectsRequest{
			Container: *objectListContainer,
			Version:   *objectListVersion,
		})
		if err != nil {
			return err
		}

		for _, object := range resp.GetObjects() {
			fmt.Println(object)
		}

		return nil
	})
	r.Register(objectURL.FullCommand(), func(ctx context.Context) error {
		url, err := cli.GetObjectURL(ctx, &v1proto.GetObjectURLRequest{
			Container: *objectURLContainer,
			Version:   *objectURLVersion,
			Key:       *objectURLKey,
		})
		if err != nil {
			return err
		}

		log.Printf("Object URL received: %s", url)
		return nil
	})

	if err := r.Call(appCmd); err != nil {
		panic(err)
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
