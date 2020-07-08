package ipc

import (
	"crypto/tls"
	"github.com/hslam/socket"
	"io"
	"net"
	"os"
)

type IPC struct {
	Config *tls.Config
}

type IPCConn struct {
	io.ReadWriteCloser
}

func (c *IPCConn) Message() socket.Message {
	return socket.NewMessage(c, 0, 0)
}

// NewSocket returns a new IPC socket.
func NewSocket() socket.Socket {
	return &IPC{}
}

func NewTLSSocket(config *tls.Config) socket.Socket {
	return &IPC{Config: config}
}

func (t *IPC) Scheme() string {
	if t.Config == nil {
		return "ipc"
	}
	return "ipcs"
}

func (t *IPC) Dial(address string) (socket.Conn, error) {
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
		return &IPCConn{conn}, err
	}
	t.Config.ServerName = address
	tlsConn := tls.Client(conn, t.Config)
	if err = tlsConn.Handshake(); err != nil {
		tlsConn.Close()
		return nil, err
	}
	return &IPCConn{tlsConn}, err
}

func (t *IPC) Listen(address string) (socket.Listener, error) {
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

func (l *IPCListener) Accept() (socket.Conn, error) {
	if conn, err := l.l.Accept(); err != nil {
		return nil, err
	} else {
		if l.config == nil {
			return &IPCConn{conn}, err
		}
		tlsConn := tls.Server(conn, l.config)
		if err = tlsConn.Handshake(); err != nil {
			tlsConn.Close()
			return nil, err
		}
		return &IPCConn{tlsConn}, err
	}
}

func (l *IPCListener) Close() error {
	return l.l.Close()
}
