package ipc

import (
	"crypto/tls"
	"github.com/hslam/network"
	"net"
	"os"
)

type IPC struct {
	Config *tls.Config
}

// NewSocket returns a new IPC socket.
func NewSocket() network.Socket {
	return &IPC{}
}

func NewTLSSocket(config *tls.Config) network.Socket {
	return &IPC{Config: config}
}

func (t *IPC) Scheme() string {
	if t.Config == nil {
		return "ipc"
	}
	return "ipcs"
}

func (t *IPC) Dial(address string) (network.Conn, error) {
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

func (t *IPC) Listen(address string) (network.Listener, error) {
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

func (l *IPCListener) Accept() (network.Conn, error) {
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

func (l *IPCListener) Close() error {
	return l.l.Close()
}
