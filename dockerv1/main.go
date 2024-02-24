package main

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/registry"
	"os"

	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"io"
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
		Remove:     true,
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

	// Log into the registry

	var authConfig = registry.AuthConfig{
		Username:      "gsaenger",
		Password:      os.Getenv("DOCKER_PASS"),
		ServerAddress: "https://index.docker.io/v1/",
	}

	authConfigBytes, _ := json.Marshal(authConfig)
	authConfigEncoded := base64.URLEncoding.EncodeToString(authConfigBytes)

	// Push image

	tag := "gsaenger/hello-go"
	pushOpts := types.ImagePushOptions{RegistryAuth: authConfigEncoded}
	rd, err := dockerClient.ImagePush(context.Background(), tag, pushOpts)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	defer rd.Close()

	err = printOutput(rd)
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
