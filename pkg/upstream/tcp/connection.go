package tcp

import (
	"fmt"
	"io"
	"net"

	"github.com/rs/zerolog/log"
)

type Connection struct {
	conn *net.TCPConn
}

func (c *Connection) Pipe(reader io.Reader, writer io.Writer) error {
	log.Debug().Msg("Start piping")

	done := make(chan struct{})
	err := make(chan error)

	go func(doneCh chan<- struct{}, errCh chan<- error) {
		if _, err := io.Copy(c.conn, reader); err != nil {
			errCh <- err
		}

		log.Debug().Msg("downstream -> upstream: completed")
		doneCh <- struct{}{}
	}(done, err)

	go func(doneCh chan<- struct{}, errCh chan<- error) {
		if _, err := io.Copy(writer, c.conn); err != nil {
			errCh <- err
		}

		log.Debug().Msg("upstream -> downstream: completed")
		doneCh <- struct{}{}
	}(done, err)

	select {
	case <-done:
		log.Debug().Msg("Successfully done piping")

		return nil

	case e := <-err:
		return e
	}
}

func (c *Connection) Close() error {
	if err := c.conn.Close(); err != nil {
		return fmt.Errorf("failed to close TCP connection: %w", err)
	}

	log.Debug().Msg("Upstream connection closed")

	return nil
}
