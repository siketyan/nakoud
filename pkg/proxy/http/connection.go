package http

import (
	"bufio"
	"net"
	"net/http"

	"github.com/rs/zerolog/log"

	"github.com/siketyan/nakoud/pkg/upstream/tcp"
)

type Connection struct {
	conn *net.TCPConn
}

func (c *Connection) Handle() error {
	reader := bufio.NewReader(c.conn)
	request, err := http.ReadRequest(reader)
	if err != nil {
		return err
	}

	if request.Method != http.MethodConnect {
		_ = (&http.Response{StatusCode: http.StatusMethodNotAllowed}).Write(c.conn)

		return nil
	}

	upstream, err := tcp.NewClient("localhost:8888").Connect()
	if err != nil {
		_ = (&http.Response{StatusCode: http.StatusBadGateway}).Write(c.conn)

		return err
	}

	defer func() {
		_ = upstream.Close()
	}()

	if err := (&http.Response{StatusCode: http.StatusOK}).Write(c.conn); err != nil {
		return err
	}

	if err := upstream.Pipe(c.conn, c.conn); err != nil {
		return err
	}

	log.Info().Msg("Downstream completed")

	return nil
}
