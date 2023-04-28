package tcp

import (
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/rs/zerolog/log"
)

type Connection struct {
	conn *net.TCPConn
}

func (c *Connection) Pipe(reader io.Reader, writer io.Writer) error {
	log.Debug().Msg("Start piping")

	waitGroup := sync.WaitGroup{}
	done := make(chan struct{})
	err := make(chan error)

	waitGroup.Add(2) //nolint:gomnd

	go func(doneCh chan<- struct{}) {
		waitGroup.Wait()
		doneCh <- struct{}{}
	}(done)

	go func(errCh chan<- error) {
		if _, err := io.Copy(c.conn, reader); err != nil {
			errCh <- err
		}

		log.Debug().Msg("downstream -> upstream: completed")
		waitGroup.Done()
	}(err)

	go func(errCh chan<- error) {
		if _, err := io.Copy(writer, c.conn); err != nil {
			errCh <- err
		}

		log.Debug().Msg("upstream -> downstream: completed")
		waitGroup.Done()
	}(err)

	select {
	case <-done:
		log.Debug().Msg("Successfully done piping")

		return nil

	case e := <-err:
		return e
	}
}

func (c *Connection) Close() error {
	log.Debug().Msg("Upstream connection closed")

	if err := c.conn.Close(); err != nil {
		return fmt.Errorf("failed to close TCP connection: %w", err)
	}

	return nil
}
