package main

import (
	"context"
	"fmt"
	"os"

	kingpin "github.com/alecthomas/kingpin/v2"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	v1pb "github.com/teran/archived/presenter/manage/grpc/proto/v1"
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
			Flag("endpoint", "manage API endpoint address").
			Short('s').
			Envar("ARCHIVED_CLI_ENDPOINT").
			Required().
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

	objectURL           = object.Command("url", "get URL for the object")
	objectURLtContainer = objectURL.Arg("container", "name of the container to publish object from").Required().String()
	objectURLVersion    = objectURL.Arg("version", "version to publish object from").Required().String()
	objectURLKey        = objectURL.Arg("key", "key of the object to publish").Required().String()

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

	dial, err := grpc.NewClient(*manageEndpoint, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	cli := v1pb.NewManageServiceClient(dial)

	switch appCmd {
	case containerCreate.FullCommand():
		_, err := cli.CreateContainer(ctx, &v1pb.CreateContainerRequest{
			Name: *containerCreateName,
		})
		if err != nil {
			panic(err)
		}
		fmt.Printf("container `%s` created", *containerCreateName)

	case version.FullCommand():
		fmt.Printf(
			"%s %s built @ %s\n",
			os.Args[0], appVersion, buildTimestamp,
		)
		os.Exit(1)
	}
}
