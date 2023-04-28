package upstream

import (
	"io"
)

type Client interface {
}

type Connection interface {
	io.Closer

	Pipe(r io.Reader, w io.Writer) error
}
