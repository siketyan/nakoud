package docker

import (
	"context"
	"fmt"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"

	"github.com/siketyan/nakoud/pkg/upstream"
	"github.com/siketyan/nakoud/pkg/upstream/tcp"
)

const (
	labelNetworkGroup = "jp.s6n.nakoud.group"
)

func sliceStartsWith[T comparable](haystack []T, needle ...T) bool {
	if len(haystack) < len(needle) {
		return false
	}

	for i, n := range needle {
		if v := haystack[i]; v != n {
			return false
		}
	}

	return true
}

type Discoverer struct{}

func (d *Discoverer) Discover(ctx context.Context, host []string) (upstream.Client, error) {
	docker, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the Docker daemon: %w", err)
	}

	networks, err := docker.NetworkList(ctx, types.NetworkListOptions{
		Filters: filters.Args{},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Docker networks: %w", err)
	}

	network, host := func() (*types.NetworkResource, []string) {
		for _, n := range networks {
			for k, v := range n.Labels {
				group := strings.Split(v, ".")
				if k == labelNetworkGroup && sliceStartsWith(host, group...) {
					return &n, host[len(group):]
				}
			}
		}

		return nil, nil
	}()
	if network == nil {
		return nil, nil
	}

	if len(host) != 1 {
		return nil, nil
	}

	for _, container := range network.Containers {
		if container.Name == host[0] {
			return tcp.NewClient(container.IPv4Address), nil
		}
	}

	return nil, nil
}
