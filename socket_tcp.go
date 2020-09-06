// Copyright (c) 2020 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package socket

import (
	"crypto/tls"
	"github.com/hslam/poll"
	"net"
)

type TCP struct {
	Config *tls.Config
	Event  *poll.Event
}

type TCPConn struct {
	net.Conn
}

func (c *TCPConn) Messages() Messages {
	return NewMessages(c, 0, 0)
}

// NewTCPSocket returns a new TCP socket.
func NewTCPSocket(config *tls.Config, event *poll.Event) Socket {
	return &TCP{Config: config, Event: event}
}

func (t *TCP) Scheme() string {
	if t.Config == nil {
		return "tcp"
	}
	return "tcps"
}

func (t *TCP) Dial(address string) (Conn, error) {
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

func (t *TCP) Listen(address string) (Listener, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", address)
	if err != nil {
		return nil, err
	}
	lis, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		return nil, err
	}
	return &TCPListener{l: lis, config: t.Config, event: t.Event}, err
}

type TCPListener struct {
	l      *net.TCPListener
	config *tls.Config
	event  *poll.Event
}

func (l *TCPListener) Accept() (Conn, error) {
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

func (l *TCPListener) Serve() error {
	if l.event == nil {
		return ErrEvent
	}
	return poll.Serve(l.l, l.event)
}

func (l *TCPListener) Close() error {
	return l.l.Close()
}

func (l *TCPListener) Addr() net.Addr {
	return l.l.Addr()
}
