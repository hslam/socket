package transport

import (
	"io"
)

type Conn interface {
	io.ReadWriteCloser
}

type Listener interface {
	// Accept waits for and returns the next connection to the listener.
	Accept() (Conn, error)

	// Close closes the listener.
	// Any blocked Accept operations will be unblocked and return errors.
	Close() error
}
type Transport interface {
	Scheme() string
	Dial(address string) (Conn, error)
	Listen(address string) (Listener, error)
}

type NewTransport func() Transport
