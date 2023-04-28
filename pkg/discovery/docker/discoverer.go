package docker

import (
	"context"
	"fmt"
	"net"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"

	"github.com/siketyan/nakoud/pkg/proto/ip"
)

const labelContainerFqdn = "jp.s6n.nakoud.fqdn"

type Discoverer struct {
	myNetworks []*net.IPNet
	docker     *client.Client
}

func NewDiscoverer() (*Discoverer, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("failed to find network interfaces: %w", err)
	}

	networks := make([]*net.IPNet, 0, len(interfaces))

	for _, i := range interfaces {
		addresses, err := i.Addrs()
		if err != nil {
			return nil, fmt.Errorf("failed to find addresses for the interface: %w", err)
		}

		for _, address := range addresses {
			switch addr := address.(type) {
			case *net.IPAddr:
				networks = append(
					networks, &net.IPNet{
						IP:   addr.IP,
						Mask: addr.IP.DefaultMask(),
					},
				)

			case *net.IPNet:
				networks = append(networks, addr)
			}
		}
	}

	log.Info().Any("MyNetworks", networks).Msg("My networks")

	docker, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the Docker daemon: %w", err)
	}

	return &Discoverer{
		myNetworks: networks,
		docker:     docker,
	}, nil
}

func (d *Discoverer) Discover(ctx context.Context, fqdn string) (*ip.Address, error) {
	log.Debug().
		Str("FQDN", fqdn).
		Msg("Discovering a upstream on Docker")

	containers, err := d.docker.ContainerList(ctx, types.ContainerListOptions{}) //nolint:exhaustruct
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Docker containers: %w", err)
	}

	for _, container := range containers {
		if f, ok := container.Labels[labelContainerFqdn]; !ok || f != fqdn {
			continue
		}

		log.Debug().
			Str("ID", container.ID).
			Msg("A Docker container matched to the FQDN")

		for _, network := range container.NetworkSettings.Networks {
			myNetwork, ok := lo.Find(
				d.myNetworks, func(myNetwork *net.IPNet) bool {
					return myNetwork.Contains(net.ParseIP(network.IPAddress))
				},
			)
			if !ok {
				continue
			}

			address, err := ip.ParseAddress(network.IPAddress, myNetwork)
			if err != nil {
				return nil, fmt.Errorf("failed to parse the IP address: %w", err)
			}

			return address, nil
		}

		log.Warn().
			Str("ID", container.ID).
			Msg("The container has no connection for the proxy, skipping")
	}

	return nil, nil
}
