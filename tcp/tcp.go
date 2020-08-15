// Copyright (c) 2020 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package tcp

import (
	"crypto/tls"
	"github.com/hslam/socket"
	"io"
	"net"
)

type TCP struct {
	Config *tls.Config
}

type TCPConn struct {
	io.ReadWriteCloser
}

func (c *TCPConn) Messages() socket.Messages {
	return socket.NewMessages(c, 0, 0)
}

// NewSocket returns a new TCP socket.
func NewSocket() socket.Socket {
	return &TCP{}
}

func NewTLSSocket(config *tls.Config) socket.Socket {
	return &TCP{Config: config}
}

func (t *TCP) Scheme() string {
	if t.Config == nil {
		return "tcp"
	}
	return "tcps"
}

func (t *TCP) Dial(address string) (socket.Conn, error) {
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
		return &TCPConn{conn}, err
	}
	t.Config.ServerName = address
	tlsConn := tls.Client(conn, t.Config)
	if err = tlsConn.Handshake(); err != nil {
		tlsConn.Close()
		return nil, err
	}
	return &TCPConn{tlsConn}, err
}

func (t *TCP) Listen(address string) (socket.Listener, error) {
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

func (l *TCPListener) Accept() (socket.Conn, error) {
	if conn, err := l.l.AcceptTCP(); err != nil {
		return nil, err
	} else {
		conn.SetNoDelay(false)
		if l.config == nil {
			return &TCPConn{conn}, err
		}
		tlsConn := tls.Server(conn, l.config)
		if err = tlsConn.Handshake(); err != nil {
			tlsConn.Close()
			return nil, err
		}
		return &TCPConn{tlsConn}, err
	}
}

func (l *TCPListener) Close() error {
	return l.l.Close()
}
