package main

import (
	"fmt"
	"os"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/siketyan/nakoud/pkg/discovery"
	"github.com/siketyan/nakoud/pkg/discovery/docker"
	"github.com/siketyan/nakoud/pkg/proxy"
	"github.com/siketyan/nakoud/pkg/proxy/http"
)

//nolint:exhaustruct, gochecknoglobals
var command = &cobra.Command{
	Use:   "nakoud-proxy",
	Short: "Access your Docker containers easily without port forwarding",
	RunE: func(cmd *cobra.Command, args []string) error {
		return run()
	},
}

func run() error {
	dockerDiscoverer, err := docker.NewDiscoverer()
	if err != nil {
		return fmt.Errorf("could not initiate Docker discoverer: %w", err)
	}

	discoverer := discovery.
		NewMux().
		With(dockerDiscoverer)

	httpProxy, err := http.NewProxy("0.0.0.0:8080", discoverer)
	if err != nil {
		return fmt.Errorf("failed to create HTTP proxy: %w", err)
	}

	proxyMux := proxy.
		NewMux().
		With(httpProxy)

	if err := proxyMux.Run(); err != nil {
		return fmt.Errorf("proxy stopped with an error: %w", err)
	}

	return nil
}

func main() {
	cobra.OnInitialize()

	flags := command.PersistentFlags()
	flags.String("bind", "127.0.0.1:8080", "where to bind the proxy")

	if err := command.Execute(); err != nil {
		log.Fatal().Err(err).Msg("Error occurred while booting")
		os.Exit(1)
	}
}
