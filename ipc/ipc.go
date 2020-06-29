package ipc

import (
	"crypto/tls"
	"hslam.com/git/x/transport"
	"net"
	"os"
)

type IPC struct {
	Config *tls.Config
}

// NewTransport returns a new IPC transport.
func NewTransport() transport.Transport {
	return &IPC{}
}
func NewTLSTransport(config *tls.Config) transport.Transport {
	return &IPC{Config: config}
}
func (t *IPC) Dial(address string) (transport.Conn, error) {
	var addr *net.UnixAddr
	var err error
	if addr, err = net.ResolveUnixAddr("unix", address); err != nil {
		return nil, err
	}
	conn, err := net.DialUnix("unix", nil, addr)
	if err != nil {
		return nil, err
	}
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

func (t *IPC) Listen(address string) (transport.Listener, error) {
	os.Remove(address)
	var addr *net.UnixAddr
	var err error
	if addr, err = net.ResolveUnixAddr("unix", address); err != nil {
		return nil, err
	}
	lis, err := net.ListenUnix("unix", addr)
	if err != nil {
		return nil, err
	}

	return &IPCListener{l: lis, config: t.Config}, err
}

type IPCListener struct {
	l      *net.UnixListener
	config *tls.Config
}

func (l *IPCListener) Accept() (transport.Conn, error) {
	if conn, err := l.l.Accept(); err != nil {
		return nil, err
	} else {
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
	return l.l.Accept()
}
