package main

import (
	"fmt"
	"log"

	"github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types"
	"golang.org/x/net/context"
)

func main() {
	c, err := client.NewClient(
		"unix:///var/run/docker.sock",
		client.DefaultVersion,
		nil,
		map[string]string{})
	if err != nil {
		log.Panic(err)
	}

	ctx := context.Background()

	ctrs, err := c.ContainerList(ctx, types.ContainerListOptions{
		All: true,
	})
	if err != nil {
		log.Panic(err)
	}

	for _, ctr := range ctrs {
		fmt.Printf("% 20s % 20s\n", ctr.ID[:12], ctr.NetworkSettings.Networks["bridge"].IPAddress)
	}
}
