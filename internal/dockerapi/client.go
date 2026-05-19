package dockerapi

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

// Client is a docker API adapter used by node-agent.
type Client struct {
	docker *client.Client
}

// NewFromEnv creates docker client configured by standard docker environment variables.
func NewFromEnv() (*Client, error) {
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("create docker client: %w", err)
	}

	return &Client{docker: dockerClient}, nil
}

// DiskUsage retrieves disk usage statistics from docker daemon.
func (c *Client) DiskUsage(ctx context.Context, options types.DiskUsageOptions) (types.DiskUsage, error) {
	return c.docker.DiskUsage(ctx, options)
}

// Close closes docker client resources.
func (c *Client) Close() error {
	return c.docker.Close()
}
