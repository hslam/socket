package tcp

import (
	"github.com/hslam/transport"
	"net"
)

type TCP struct {
}

// NewTransport returns a new TCP transport.
func NewTransport() transport.Transport {
	return &TCP{}
}
func (t *TCP) Dial(address string) (transport.Conn, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", address)
	if err != nil {
		return nil, err
	}
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		return nil, err
	}
	conn.SetNoDelay(true)
	return conn, err
}

func (t *TCP) Listen(address string) (transport.Listener, error) {
	lis, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}
	return &TCPListener{lis}, err
}

type TCPListener struct {
	l net.Listener
}

func (l *TCPListener) Accept() (transport.Conn, error) {
	return l.l.Accept()
}
