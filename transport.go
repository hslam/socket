package transport

import (
	"io"
)

type Conn interface {
	io.ReadWriteCloser
}

type Listener interface {
	Accept() (Conn, error)
}
type Transport interface {
	Dial(address string) (Conn, error)
	Listen(address string) (Listener, error)
}
