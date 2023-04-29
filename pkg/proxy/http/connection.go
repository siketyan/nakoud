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

	"github.com/siketyan/nakoud/pkg/discovery"
	"github.com/siketyan/nakoud/pkg/upstream"
	httpUpstream "github.com/siketyan/nakoud/pkg/upstream/http"
	tcpUpstream "github.com/siketyan/nakoud/pkg/upstream/tcp"
)

var ErrNoUpstreamAvailable = errors.New("no upstream is available for the FQDN")

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

func (c *Connection) connectTCPUpstream(ctx context.Context, request *http.Request) (upstream.Connection, *httpError) {
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

	connection, err := tcpUpstream.NewClient(address.AsTCP(uint16(port))).Connect()
	if err != nil {
		return nil, newErrorResponse(
			http.StatusBadGateway,
			fmt.Errorf("failed to connect to the upstream: %w", err),
		)
	}

	return connection, nil
}

func (c *Connection) connectHTTPUpstream(ctx context.Context, request *http.Request) (upstream.Connection, *httpError) {
	url := request.URL

	address, err := c.discoverer.Discover(ctx, url.Hostname())
	if err != nil {
		return nil, newErrorResponse(http.StatusBadGateway, fmt.Errorf("failed to discover upstream: %w", err))
	}

	if address == nil {
		return nil, newErrorResponse(http.StatusNotFound, ErrNoUpstreamAvailable)
	}

	// HACK: Clear the request URI to treat this as a client request
	request.RequestURI = ""

	if url.Port() == "" {
		url.Host = address.String()
	} else {
		url.Host = fmt.Sprintf("%s:%s", address.String(), url.Port())
	}

	connection, err := httpUpstream.NewClientWithDefaultTransport(request).Connect()
	if err != nil {
		return nil, newErrorResponse(
			http.StatusBadGateway,
			fmt.Errorf("failed to connect to the upstream: %w", err),
		)
	}

	return connection, nil
}

func (c *Connection) connectUpstream(ctx context.Context, request *http.Request) (upstream.Connection, error) {
	if request.Method == http.MethodConnect {
		connection, err := c.connectTCPUpstream(ctx, request)
		if err != nil {
			_ = writeHeadResponse(c.inner, err.statusCode)

			return nil, fmt.Errorf("failed to connect to the upstream: %w", err)
		}

		if err := writeHeadResponse(c.inner, http.StatusOK); err != nil {
			return nil, err
		}

		return connection, nil
	}

	connection, err := c.connectHTTPUpstream(ctx, request)
	if err != nil {
		_ = writeHeadResponse(c.inner, err.statusCode)

		return nil, fmt.Errorf("failed to connect to the upstream: %w", err)
	}

	return connection, nil
}

func (c *Connection) Handle(ctx context.Context) error {
	reader := bufio.NewReader(c.inner)

	for {
		request, err := http.ReadRequest(reader)
		if err != nil {
			// Handles the client disconnects from the proxy.
			if errors.Is(err, io.EOF) {
				break
			}

			return fmt.Errorf("failed to read a HTTP request: %w", err)
		}

		connection, err := c.connectUpstream(ctx, request)
		if err != nil {
			return err
		}

		if err := connection.Pipe(c.inner, c.inner); err != nil {
			return fmt.Errorf("failed to pipe between upstream and downstream: %w", err)
		}

		_ = connection.Close()

		// To reuse the proxy connection for another usages, keeps alive the HTTP connection if needed.
		if request.Header.Get("Proxy-Connection") != "keep-alive" {
			break
		}
	}

	return nil
}

func (c *Connection) Close() error {
	if err := c.inner.Close(); err != nil {
		return fmt.Errorf("failed to close the TCP connection: %w", err)
	}

	return nil
}
