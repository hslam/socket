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
}

type TCPConn struct {
	net.Conn
}

func (c *TCPConn) Messages() Messages {
	return NewMessages(c, 0, 0)
}

// NewTCPSocket returns a new TCP socket.
func NewTCPSocket(config *tls.Config) Socket {
	return &TCP{Config: config}
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
		conn.Close()
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
	return &TCPListener{l: lis, config: t.Config}, err
}

type TCPListener struct {
	l      *net.TCPListener
	config *tls.Config
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
			conn.Close()
			return nil, err
		}
		return &TCPConn{tlsConn}, err
	}
}

func (l *TCPListener) Serve(event *poll.Event) error {
	if event == nil {
		return ErrEvent
	}
	return poll.Serve(l.l, event)
}

func (l *TCPListener) ServeConn(handle func(req []byte) (res []byte)) error {
	event := &poll.Event{
		UpgradeConn: func(conn net.Conn) (upgrade net.Conn, err error) {
			if l.config != nil {
				tlsConn := tls.Server(conn, l.config)
				if err := tlsConn.Handshake(); err != nil {
					conn.Close()
					return nil, err
				}
				upgrade = tlsConn
			}
			upgrade = conn
			return
		},
		Handle: handle,
	}
	return poll.Serve(l.l, event)
}

func (l *TCPListener) ServeMessages(handle func(req []byte) (res []byte)) error {
	event := &poll.Event{
		UpgradeHandle: func(conn net.Conn) (func() error, error) {
			if l.config != nil {
				tlsConn := tls.Server(conn, l.config)
				if err := tlsConn.Handshake(); err != nil {
					conn.Close()
					return nil, err
				}
				conn = tlsConn
			}
			messages := NewMessages(conn, 0, 0)
			return func() error {
				req, err := messages.ReadMessage()
				if err != nil {
					return err
				}
				res := handle(req)
				if len(res) > 0 {
					err = messages.WriteMessage(res)
				}
				return err
			}, nil
		},
	}
	return poll.Serve(l.l, event)
}

func (l *TCPListener) Close() error {
	return l.l.Close()
}

func (l *TCPListener) Addr() net.Addr {
	return l.l.Addr()
}
