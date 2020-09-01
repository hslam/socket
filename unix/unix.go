// Copyright (c) 2020 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package unix

import (
	"crypto/tls"
	"github.com/hslam/socket"
	"net"
	"os"
)

type UNIX struct {
	Config *tls.Config
}

type UNIXConn struct {
	net.Conn
}

func (c *UNIXConn) Messages() socket.Messages {
	return socket.NewMessages(c, 0, 0)
}

// NewSocket returns a new IPC socket.
func NewSocket() socket.Socket {
	return &UNIX{}
}

func NewTLSSocket(config *tls.Config) socket.Socket {
	return &UNIX{Config: config}
}

func (t *UNIX) Scheme() string {
	if t.Config == nil {
		return "unix"
	}
	return "unixs"
}

func (t *UNIX) Dial(address string) (socket.Conn, error) {
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
		return &UNIXConn{conn}, err
	}
	t.Config.ServerName = address
	tlsConn := tls.Client(conn, t.Config)
	if err = tlsConn.Handshake(); err != nil {
		tlsConn.Close()
		return nil, err
	}
	return &UNIXConn{tlsConn}, err
}

func (t *UNIX) Listen(address string) (socket.Listener, error) {
	os.RemoveAll(address)
	var addr *net.UnixAddr
	var err error
	if addr, err = net.ResolveUnixAddr("unix", address); err != nil {
		return nil, err
	}
	lis, err := net.ListenUnix("unix", addr)
	if err != nil {
		return nil, err
	}

	return &UNIXListener{l: lis, config: t.Config}, err
}

type UNIXListener struct {
	l      *net.UnixListener
	config *tls.Config
}

func (l *UNIXListener) Accept() (socket.Conn, error) {
	if conn, err := l.l.Accept(); err != nil {
		return nil, err
	} else {
		if l.config == nil {
			return &UNIXConn{conn}, err
		}
		tlsConn := tls.Server(conn, l.config)
		if err = tlsConn.Handshake(); err != nil {
			tlsConn.Close()
			return nil, err
		}
		return &UNIXConn{tlsConn}, err
	}
}

func (l *UNIXListener) Close() error {
	return l.l.Close()
}

func (l *UNIXListener) Addr() net.Addr {
	return l.l.Addr()
}
