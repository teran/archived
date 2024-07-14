package main

import (
	"context"
	"crypto/tls"
	"os"

	kingpin "github.com/alecthomas/kingpin/v2"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/teran/archived/cli/router"
	"github.com/teran/archived/cli/service"
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
	cliSvc := service.New(cli)

	r := router.New(ctx)
	r.Register(containerCreate.FullCommand(), cliSvc.CreateContainer(*containerCreateName))
	r.Register(containerList.FullCommand(), cliSvc.ListContainers())
	r.Register(versionList.FullCommand(), cliSvc.ListVersions(*versionListContainer))
	r.Register(versionCreate.FullCommand(), cliSvc.CreateVersion(*versionCreateContainer))
	r.Register(versionPublish.FullCommand(), cliSvc.PublishVersion(*versionPublishContainer, *versionPublishVersion))
	r.Register(objectCreate.FullCommand(), cliSvc.CreateObject(*objectCreateContainer, *objectCreateVersion, *objectCreatePath))
	r.Register(objectList.FullCommand(), cliSvc.ListObjects(*objectListContainer, *objectListVersion))
	r.Register(objectURL.FullCommand(), cliSvc.GetObjectURL(*objectURLContainer, *objectURLVersion, *objectURLKey))

	if err := r.Call(appCmd); err != nil {
		panic(err)
	}
}
