package tcp

import (
	"io"
	"net"
	"sync"

	"github.com/rs/zerolog/log"
)

type Connection struct {
	conn *net.TCPConn
}

func (c *Connection) Pipe(r io.Reader, w io.Writer) error {
	log.Debug().Msg("Start piping")

	waitGroup := sync.WaitGroup{}
	done := make(chan struct{})
	err := make(chan error)

	waitGroup.Add(2)
	go func(doneCh chan<- struct{}) {
		waitGroup.Wait()
		doneCh <- struct{}{}
	}(done)

	go func(errCh chan<- error) {
		if _, err := io.Copy(c.conn, r); err != nil {
			errCh <- err
		}

		log.Debug().Msg("downstream -> upstream: completed")
		waitGroup.Done()
	}(err)

	go func(errCh chan<- error) {
		if _, err := io.Copy(w, c.conn); err != nil {
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

	return c.conn.Close()
}
