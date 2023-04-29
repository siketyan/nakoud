package http

import (
	"fmt"
	"net/http"

	"github.com/rs/zerolog/log"

	"github.com/siketyan/nakoud/pkg/upstream"
)

type Client struct {
	transport http.RoundTripper
	request   *http.Request
}

func NewClient(transport http.RoundTripper, request *http.Request) *Client {
	return &Client{
		transport: transport,
		request:   request,
	}
}

func NewClientWithDefaultTransport(request *http.Request) *Client {
	return NewClient(http.DefaultTransport, request)
}

func (c *Client) Connect() (upstream.Connection, error) {
	log.Info().Any("URL", c.request.URL).Msg("Connecting to a HTTP upstream")

	response, err := http.DefaultClient.Do(c.request) //nolint:bodyclose
	if err != nil {
		return nil, fmt.Errorf("failed to connect to HTTP upstream: %w", err)
	}

	return NewConnection(response), nil
}
