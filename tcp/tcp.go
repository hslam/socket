package tcp

import (
	"crypto/tls"
	"github.com/hslam/transport"
	"net"
)

type TCP struct {
	Config *tls.Config
}

// NewTransport returns a new TCP transport.
func NewTransport() transport.Transport {
	return &TCP{}
}

func NewTLSTransport(config *tls.Config) transport.Transport {
	return &TCP{Config: config}
}

func (t *TCP) Scheme() string {
	return "tcp"
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
	conn.SetNoDelay(false)
	if t.Config == nil {
		return conn, err
	}
	t.Config.ServerName = address
	tlsConn := tls.Client(conn, t.Config)
	if err = tlsConn.Handshake(); err != nil {
		tlsConn.Close()
		return nil, err
	}
	return tlsConn, err
}

func (t *TCP) Listen(address string) (transport.Listener, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", address)
	if err != nil {
		return nil, err
	}
	lis, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		return nil, err
	}
	return &TCPListener{l: lis, config: t.Config}, err
}

type TCPListener struct {
	l      *net.TCPListener
	config *tls.Config
}

func (l *TCPListener) Accept() (transport.Conn, error) {
	if conn, err := l.l.AcceptTCP(); err != nil {
		return nil, err
	} else {
		conn.SetNoDelay(false)
		if l.config == nil {
			return conn, err
		}
		tlsConn := tls.Server(conn, l.config)
		if err = tlsConn.Handshake(); err != nil {
			tlsConn.Close()
			return nil, err
		}
		return tlsConn, err
	}
}

func (l *TCPListener) Close() error {
	return l.l.Close()
}
