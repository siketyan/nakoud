package main

import (
	"github.com/rs/zerolog/log"

	"github.com/siketyan/nakoud/pkg/discovery"
	"github.com/siketyan/nakoud/pkg/discovery/docker"
	"github.com/siketyan/nakoud/pkg/proxy"
	"github.com/siketyan/nakoud/pkg/proxy/http"
)

func main() {
	dockerDiscoverer, err := docker.NewDiscoverer()
	if err != nil {
		log.Error().Err(err).Msg("Could not initiate the Docker discoverer")

		return
	}

	discoverer := discovery.NewMux().With(dockerDiscoverer)

	httpProxy, err := http.NewProxy("0.0.0.0:8080", discoverer)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create HTTP proxy")

		return
	}

	if err := proxy.NewMux().With(httpProxy).Run(); err != nil {
		log.Error().Err(err).Msg("Proxy mux stopped with an error")

		return
	}
}
