package socket

import (
	"io"
)

type Conn interface {
	io.ReadWriteCloser
}

type Dialer interface {
	Dial(address string) (Conn, error)
}

type Listener interface {
	// Accept waits for and returns the next connection to the listener.
	Accept() (Conn, error)

	// Close closes the listener.
	// Any blocked Accept operations will be unblocked and return errors.
	Close() error
}

type Socket interface {
	Scheme() string
	Dialer
	Listen(address string) (Listener, error)
}

type NewSocket func() Socket
