package main

import (
	"github.com/rs/zerolog/log"

	"github.com/siketyan/nakoud/pkg/proxy"
	"github.com/siketyan/nakoud/pkg/proxy/http"
)

func main() {
	httpProxy, err := http.NewProxy("127.0.0.1:8080")
	if err != nil {
		log.Error().Err(err).Msg("Failed to create HTTP proxy")

		return
	}

	if err := proxy.NewMux().With(httpProxy).Run(); err != nil {
		log.Error().Err(err).Msg("Proxy mux stopped with an error")

		return
	}
}
