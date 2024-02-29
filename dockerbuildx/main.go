package main

import (
	"context"
	"fmt"
	buildx "github.com/docker/buildx/build"
	"github.com/docker/buildx/builder"
	"github.com/docker/buildx/controller/pb"
	"github.com/docker/buildx/util/dockerutil"
	"github.com/docker/buildx/util/progress"
	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/flags"
	"github.com/moby/buildkit/util/progress/progressui"
	"path/filepath"

	_ "github.com/docker/buildx/driver/docker-container"
	_ "github.com/docker/buildx/driver/kubernetes"
	_ "github.com/docker/buildx/driver/remote"
	//cfgtypes "github.com/docker/cli/cli/config/types"
	"os"
)

func main() {
	cli, err := command.NewDockerCli(
		command.WithCombinedStreams(os.Stdout),
	)
	buildCtx := context.Background()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	opts := &flags.ClientOptions{}
	err = cli.Initialize(opts)

	pbOpts := pb.BuildOptions{
		ContextPath:    "./app/",
		DockerfileName: "./app/Dockerfile",
		//Builder:        "newBuilder",
	}

	builder, err := builder.New(cli,
		builder.WithName(pbOpts.Builder),
		builder.WithContextPathHash("/Users/guin/go/src/github.com/guineveresaenger/docker-talk/dockerbuildx/app"),
	)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	nodes, err := builder.LoadNodes(buildCtx)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	payload := map[string]buildx.Options{}
	payload["default"] = buildx.Options{
		Inputs: buildx.Inputs{
			ContextPath:    "app",
			DockerfilePath: "app/Dockerfile",
		},

		//Platforms: []string{"arm64"},

		Tags: []string{"gsaenger/buildx-hello-go"},
	}

	printer, err := progress.NewPrinter(buildCtx, os.Stdout,
		progressui.PlainMode,
		progress.WithDesc(
			fmt.Sprintf("building with %q instance using %s driver", builder.Name, builder.Driver),
			fmt.Sprintf("%s:%s", builder.Driver, builder.Name),
		),
	)

	results, err := buildx.Build(
		buildCtx,
		nodes,
		payload,
		dockerutil.NewClient(cli),
		filepath.Dir(cli.ConfigFile().Filename),
		printer,
	)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(results)
	fmt.Println("Tönis Tiigi is a jerk")
}

//func getCachedBuilder(opts controllerapi.BuildOptions,
//) (*cachedBuilder, error) {
//
//	contextPathHash := opts.ContextPath
//	if absContextPath, err := filepath.Abs(contextPathHash); err == nil {
//		contextPathHash = absContextPath
//	}
//	b, err := builder.New(d.cli,
//		builder.WithName(opts.Builder),
//		builder.WithContextPathHash(contextPathHash),
//	)
//	if err != nil {
//		return nil, err
//	}
//
//	nodes, err := b.LoadNodes(context.Background())
//	if err != nil {
//		return nil, err
//	}
//
//	cached := &cachedBuilder{name: b.Name, driver: b.Driver, nodes: nodes}
//	d.builders[opts.Builder] = cached
//
//	return cached, nil
//}

type cachedBuilder struct {
	name   string
	driver string
	nodes  []builder.Node
}
