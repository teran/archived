package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	kingpin "github.com/alecthomas/kingpin/v2"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/teran/archived/cli/router"
	"github.com/teran/archived/cli/service"
	"github.com/teran/archived/cli/service/stat_cache/local"
	v1proto "github.com/teran/archived/manager/presenter/grpc/proto/v1"
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
			Flag("endpoint", "Manager API endpoint address").
			Short('s').
			Envar("ARCHIVED_CLI_ENDPOINT").
			Required().
			String()

	insecureFlag = app.Flag("insecure", "Do not use TLS for gRPC connection").
			Default("false").
			Bool()
	insecureSkipVerify = app.Flag("insecure-skip-verify", "Do not perform TLS certificate verification for gRPC connection").
				Default("false").
				Bool()

	cacheDir = app.Flag("cache-dir", "Stat-cache directory for objects").
			Default("~/.cache/archived/cli/objects").
			Envar("ARCHIVED_CLI_STAT_CACHE_DIR").
			String()

	container           = app.Command("container", "container operations")
	containerCreate     = container.Command("create", "create new container")
	containerCreateName = containerCreate.Arg("name", "name of the container to create").Required().String()

	containerDelete     = container.Command("delete", "delete the given container")
	containerDeleteName = containerDelete.Arg("name", "name of the container to delete").Required().String()

	containerList = container.Command("list", "list containers")

	version                = app.Command("version", "version operations")
	versionCreate          = version.Command("create", "create new version for given container")
	versionCreateContainer = versionCreate.Arg("container", "name of the container to create version for").Required().String()
	versionCreatePublish   = versionCreate.Flag("publish", "publish version right after creating").
				Default("false").
				Bool()
	versionCreateFromDir = versionCreate.Flag("from-dir", "create version right from directory").
				String()
	versionCreateFromYumRepo = versionCreate.Flag("from-yum-repo", "create version right from yum repository").
					String()
	versionCreateFromYumRepoGPGKey = versionCreate.Flag("rpm-gpg-key-path", "path to the GPG key for RPM packages verification").
					String()

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

	statCache         = app.Command("stat-cache", "stat cache operations")
	statCacheShowPath = statCache.Command("show-path", "print actual cache path")
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
	} else {
		log.SetLevel(log.InfoLevel)
		log.SetFormatter(&log.TextFormatter{
			FullTimestamp: true,
		})
	}

	log.Debugf("Initializing gRPC client ...")

	grpcOpts := []grpc.DialOption{
		grpc.WithUserAgent("archived-cli/0.1"),
	}
	if *insecureFlag {
		log.Warn("insecure flag is specified which means no TLS is in use!")
		grpcOpts = append(grpcOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		if *insecureSkipVerify {
			log.Warn("insecure-skip-verify flag in specified which means high risk of man-in-the-middle attack!")
		}
		grpcOpts = append(grpcOpts, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
			InsecureSkipVerify: *insecureSkipVerify,
		})))
	}

	dial, err := grpc.NewClient(*manageEndpoint, grpcOpts...)
	if err != nil {
		panic(err)
	}

	log.Debugf("Initializing manage service client ...")
	cli := v1proto.NewManageServiceClient(dial)

	log.Debugf("Initializing cache directory at `%s`", *cacheDir)

	dir := normalizeHomeDir(*cacheDir)
	log.Tracef("normalized cache directory: %s", dir)

	cacheRepo, err := local.New(dir)
	if err != nil {
		panic(err)
	}

	cliSvc := service.New(cli, cacheRepo)

	r := router.New(ctx)
	r.Register(containerCreate.FullCommand(), cliSvc.CreateContainer(*containerCreateName))
	r.Register(containerList.FullCommand(), cliSvc.ListContainers())
	r.Register(containerDelete.FullCommand(), cliSvc.DeleteContainer(*containerDeleteName))

	r.Register(versionList.FullCommand(), cliSvc.ListVersions(*versionListContainer))
	r.Register(versionCreate.FullCommand(), cliSvc.CreateVersion(
		*versionCreateContainer, *versionCreatePublish, versionCreateFromDir, versionCreateFromYumRepo, versionCreateFromYumRepoGPGKey))
	r.Register(versionDelete.FullCommand(), cliSvc.DeleteVersion(*versionDeleteContainer, *versionDeleteVersion))
	r.Register(versionPublish.FullCommand(), cliSvc.PublishVersion(*versionPublishContainer, *versionPublishVersion))

	r.Register(objectCreate.FullCommand(), cliSvc.CreateObject(*objectCreateContainer, *objectCreateVersion, *objectCreatePath))
	r.Register(objectList.FullCommand(), cliSvc.ListObjects(*objectListContainer, *objectListVersion))
	r.Register(objectURL.FullCommand(), cliSvc.GetObjectURL(*objectURLContainer, *objectURLVersion, *objectURLKey))

	r.Register(statCacheShowPath.FullCommand(), func(ctx context.Context) error {
		fmt.Println(*cacheDir)
		return nil
	})

	if err := r.Call(appCmd); err != nil {
		panic(err)
	}
}

func normalizeHomeDir(in string) (out string) {
	usr, _ := user.Current()
	dir := usr.HomeDir
	out = in
	if strings.HasPrefix(in, "~/") {
		out = filepath.Join(dir, in[2:])
	}
	return out
}
