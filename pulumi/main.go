package main

import (
	"github.com/pulumi/pulumi-docker/sdk/v4/go/docker"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		demoImage, err := docker.NewImage(ctx, "demo-image", &docker.ImageArgs{
			Build: &docker.DockerBuildArgs{
				Args: pulumi.StringMap{
					"platform": pulumi.String("linux/amd64"),
				},
				Context:    pulumi.String("app"),
				Dockerfile: pulumi.String("app/Dockerfile"),
			},
			ImageName: pulumi.String("username/image:tag1"),
			SkipPush:  pulumi.Bool(true),
		})
		if err != nil {
			return err
		}
		//_, err = docker.NewContainer(ctx, "demoContainer", &docker.ContainerArgs{
		//	Image: demoImage.ID(),
		//})
		ctx.Export("imageName", demoImage.ID())
		return nil
	})
}
