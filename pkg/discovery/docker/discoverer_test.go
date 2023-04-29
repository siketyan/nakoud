package docker_test

import (
	"context"
	"github.com/siketyan/nakoud/pkg/discovery/docker"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/stretchr/testify/require"
)

func TestDiscoverer_Discover(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	d, err := client.NewClientWithOpts(client.FromEnv)
	require.NoError(t, err)

	network, err := d.NetworkCreate(ctx, "nakoud-network-test", types.NetworkCreate{ //nolint:exhaustruct
		Labels: map[string]string{
			"jp.s6n.nakoud.group": "local.test",
		},
	})
	require.NoError(t, err)

	defer func() {
		_ = d.NetworkRemove(ctx, network.ID)
	}()

	(&docker.Discoverer{}).Discover(ctx, []string{"local.test.hoge"})
}
