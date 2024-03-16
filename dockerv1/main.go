package main

import (
	"bufio"
	"context"
	_ "encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/docker/cli/cli/config"
	"github.com/docker/cli/cli/config/configfile"
	"github.com/docker/cli/cli/config/credentials"
	"github.com/docker/docker/api/types"
	_ "github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	_ "github.com/moby/buildkit/identity"
	_ "github.com/moby/buildkit/session"
	_ "github.com/moby/buildkit/session/auth/authprovider"
	"io"
	_ "net"
	_ "os"
)

func main() {
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	tar, err := archive.TarWithOptions("app/", &archive.TarOptions{})
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	opts := types.ImageBuildOptions{
		Dockerfile: "Dockerfile",
		Tags:       []string{"gsaenger/hello-go"},
		Version:    types.BuilderV1,
		Platform:   "amd64",
	}

	res, err := dockerClient.ImageBuild(context.Background(), tar, opts)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	defer res.Body.Close()

	err = printOutput(res.Body)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

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

func getDefaultDockerConfig() (*configfile.ConfigFile, error) {
	cfg, err := config.Load(config.Dir())
	if err != nil {
		return nil, err
	}
	cfg.CredentialsStore = credentials.DetectDefaultStore(cfg.CredentialsStore)
	return cfg, nil
}
