package tools

import (
	"context"
	"fmt"
	"os"

	"github.com/docker/docker/client"
)

// GetDockerClient tries to create a Docker client using environment variables
func GetDockerClient() (*client.Client, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err == nil {
		_, pingErr := cli.Ping(context.Background())
		if pingErr == nil {
			return cli, nil
		}
	}

	socketPaths := []string{
		"/var/run/docker.sock",
		os.Getenv("HOME") + "/.docker/run/docker.sock",
		"/run/docker.sock",
	}

	for _, socketPath := range socketPaths {
		if _, err := os.Stat(socketPath); err == nil {
			cli, err := client.NewClientWithOpts(
				client.WithHost("unix://"+socketPath),
				client.WithAPIVersionNegotiation(),
			)
			if err == nil {
				_, pingErr := cli.Ping(context.Background())
				if pingErr == nil {
					return cli, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("could not connect to Docker daemon")
}
