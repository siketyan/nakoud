package http

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"

	"github.com/rs/zerolog/log"

	"github.com/siketyan/nakoud/pkg/upstream/tcp"
)

func writeHeadResponse(writer io.Writer, statusCode int) error {
	response := &http.Response{ //nolint:exhaustruct
		StatusCode: statusCode,
		ProtoMajor: 1,
		ProtoMinor: 1,
	}

	if err := response.Write(writer); err != nil {
		return fmt.Errorf("failed to send a head response: %w", err)
	}

	return nil
}

type Connection struct {
	conn *net.TCPConn
}

func (c *Connection) Handle() error {
	reader := bufio.NewReader(c.conn)

	request, err := http.ReadRequest(reader)
	if err != nil {
		return fmt.Errorf("failed to read a HTTP request: %w", err)
	}

	if request.Method != http.MethodConnect {
		_ = writeHeadResponse(c.conn, http.StatusMethodNotAllowed)

		return nil
	}

	upstream, err := tcp.NewClient("localhost:8888").Connect()
	if err != nil {
		_ = writeHeadResponse(c.conn, http.StatusBadGateway)

		return fmt.Errorf("failed to connect to the upstream: %w", err)
	}

	defer func() {
		_ = upstream.Close()
	}()

	if err := writeHeadResponse(c.conn, http.StatusOK); err != nil {
		return err
	}

	if err := upstream.Pipe(c.conn, c.conn); err != nil {
		return fmt.Errorf("failed to pipe between upstream and downstream: %w", err)
	}

	log.Info().Msg("Downstream completed")

	return nil
}
