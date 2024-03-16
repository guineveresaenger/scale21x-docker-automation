package main

import (
	"bufio"
	"context"
	_ "encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	buildx "github.com/docker/buildx/build"
	"github.com/docker/buildx/builder"
	"github.com/docker/buildx/controller/pb"
	"github.com/docker/buildx/util/dockerutil"
	"github.com/docker/buildx/util/progress"
	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/flags"
	_ "github.com/docker/docker/api/types"
	_ "github.com/docker/docker/api/types/registry"
	_ "github.com/docker/docker/client"
	"github.com/moby/buildkit/util/progress/progressui"
	"io"
	"path/filepath"

	_ "github.com/docker/buildx/driver/docker-container"
	_ "github.com/docker/buildx/driver/kubernetes"
	_ "github.com/docker/buildx/driver/remote"
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
	}

	builder, err := builder.New(cli,
		builder.WithName(pbOpts.Builder),
		builder.WithContextPathHash(pbOpts.ContextPath),
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

	// Ensure we load the resulting image into our local image store
	// This represents the `--load` option on `buildx build`.
	defaultExport := pb.ExportEntry{
		Type: "docker",
	}
	exp := []*pb.ExportEntry{&defaultExport}
	outputs, err := pb.CreateExports(exp)

	payload["default"] = buildx.Options{
		Inputs: buildx.Inputs{
			ContextPath:    "app",
			DockerfilePath: "app/Dockerfile",
		},
		Exports: outputs,

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
	for key, val := range results {
		fmt.Println(key, ": ", val)
	}
	fmt.Println(results)

}

type ErrorLine struct {
	Error       string      `json:"error"`
	ErrorDetail ErrorDetail `json:"errorDetail"`
}

type ErrorDetail struct {
	Message string `json:"message"`
}

func printOutput(rd io.Reader) error {
	var lastLine string

	scanner := bufio.NewScanner(rd)
	for scanner.Scan() {
		lastLine = scanner.Text()
		fmt.Println(scanner.Text())
	}

	errLine := &ErrorLine{}
	json.Unmarshal([]byte(lastLine), errLine)
	if errLine.Error != "" {
		return errors.New(errLine.Error)
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}
