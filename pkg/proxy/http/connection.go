package http

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"

	"github.com/rs/zerolog/log"

	"github.com/siketyan/nakoud/pkg/discovery"
	"github.com/siketyan/nakoud/pkg/upstream"
	"github.com/siketyan/nakoud/pkg/upstream/tcp"
)

var (
	ErrMalformedRequest    = errors.New("malformed HTTP request")
	ErrNoUpstreamAvailable = errors.New("no upstream is available for the FQDN")
)

func writeHeadResponse(writer io.Writer, statusCode int) error {
	//nolint:exhaustruct
	response := &http.Response{
		StatusCode: statusCode,
		ProtoMajor: 1,
		ProtoMinor: 1,
	}

	if err := response.Write(writer); err != nil {
		return fmt.Errorf("failed to send a head response: %w", err)
	}

	return nil
}

type httpError struct {
	statusCode int
	err        error
}

func newErrorResponse(statusCode int, err error) *httpError {
	return &httpError{
		statusCode: statusCode,
		err:        err,
	}
}

func (e *httpError) Error() string {
	return e.err.Error()
}

type Connection struct {
	inner      *net.TCPConn
	discoverer discovery.Discoverer
}

func NewConnection(inner *net.TCPConn, discoverer discovery.Discoverer) *Connection {
	return &Connection{
		inner:      inner,
		discoverer: discoverer,
	}
}

func (c *Connection) connectUpstream(ctx context.Context, request *http.Request) (upstream.Connection, *httpError) {
	if request.Method != http.MethodConnect {
		return nil, newErrorResponse(http.StatusMethodNotAllowed, ErrMalformedRequest)
	}

	log.Info().Any("Request", request).Msg("Read a HTTP request")

	fqdn, portString, err := net.SplitHostPort(request.Host)
	if err != nil {
		return nil, newErrorResponse(http.StatusBadRequest, fmt.Errorf("malformed host: %w", err))
	}

	port, err := strconv.Atoi(portString)
	if err != nil {
		return nil, newErrorResponse(http.StatusBadRequest, fmt.Errorf("malformed TCP port: %w", err))
	}

	address, err := c.discoverer.Discover(ctx, fqdn)
	if err != nil {
		return nil, newErrorResponse(http.StatusBadGateway, fmt.Errorf("failed to discover upstream: %w", err))
	}

	if address == nil {
		return nil, newErrorResponse(http.StatusNotFound, ErrNoUpstreamAvailable)
	}

	connection, err := tcp.NewClient(address.AsTCP(uint16(port))).Connect()
	if err != nil {
		return nil, newErrorResponse(
			http.StatusBadGateway,
			fmt.Errorf("failed to connect to the upstream: %w", err),
		)
	}

	return connection, nil
}

func (c *Connection) Handle(ctx context.Context) error {
	reader := bufio.NewReader(c.inner)

	request, err := http.ReadRequest(reader)
	if err != nil {
		return fmt.Errorf("failed to read a HTTP request: %w", err)
	}

	connection, httpErr := c.connectUpstream(ctx, request)
	if err != nil {
		_ = writeHeadResponse(c.inner, httpErr.statusCode)

		return fmt.Errorf("failed to connect to the upstream: %w", httpErr)
	}

	defer func() {
		_ = connection.Close()
	}()

	if err := writeHeadResponse(c.inner, http.StatusOK); err != nil {
		return err
	}

	if err := connection.Pipe(c.inner, c.inner); err != nil {
		return fmt.Errorf("failed to pipe between upstream and downstream: %w", err)
	}

	return nil
}
