package http

import (
	"fmt"
	"io"
	"net/http"

	"github.com/rs/zerolog/log"
)

type Connection struct {
	response *http.Response
}

func NewConnection(response *http.Response) *Connection {
	return &Connection{
		response: response,
	}
}

func (c *Connection) Pipe(_ io.Reader, writer io.Writer) error {
	log.Debug().Msg("Start piping")

	if err := c.response.Write(writer); err != nil {
		return fmt.Errorf("failed to pipe HTTP messages: %w", err)
	}

	log.Debug().Msg("Successfully finished piping")

	return nil
}

func (c *Connection) Close() error {
	if err := c.response.Body.Close(); err != nil {
		return fmt.Errorf("failed to close the HTTP connection: %w", err)
	}

	log.Debug().Msg("Upstream connection closed")

	return nil
}
