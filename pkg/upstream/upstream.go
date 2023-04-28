package upstream

import (
	"io"
)

type Client interface {
	Connect() (Connection, error)
}

type Connection interface {
	io.Closer

	Pipe(r io.Reader, w io.Writer) error
}
