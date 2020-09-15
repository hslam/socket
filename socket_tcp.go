// Copyright (c) 2020 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package socket

import (
	"crypto/tls"
	"github.com/hslam/netpoll"
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

func (l *TCPListener) Serve(event *netpoll.Event) error {
	if event == nil {
		return ErrEvent
	}
	return netpoll.Serve(l.l, event)
}

func (l *TCPListener) ServeData(opened func(net.Conn) error, handler func(req []byte) (res []byte)) error {
	event := &netpoll.Event{
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
			if opened != nil {
				if err = opened(upgrade); err != nil {
					upgrade.Close()
					return
				}
			}
			return
		},
		Handler: handler,
	}
	return netpoll.Serve(l.l, event)
}

func (l *TCPListener) ServeConn(opened func(net.Conn) (Context, error), handler func(Context) error) error {
	event := &netpoll.Event{
		UpgradeHandler: func(conn net.Conn) (func() error, error) {
			if l.config != nil {
				tlsConn := tls.Server(conn, l.config)
				if err := tlsConn.Handshake(); err != nil {
					conn.Close()
					return nil, err
				}
				conn = tlsConn
			}
			var context Context
			var err error
			if opened != nil {
				context, err = opened(conn)
				if err != nil {
					return nil, err
				}
			}
			return func() error {
				err := handler(context)
				if err == netpoll.EOF {
					conn.Close()
				}
				return err
			}, nil
		},
	}
	return netpoll.Serve(l.l, event)
}

func (l *TCPListener) ServeMessages(opened func(Messages) (Context, error), handler func(Context) error) error {
	event := &netpoll.Event{
		UpgradeHandler: func(conn net.Conn) (func() error, error) {
			if l.config != nil {
				tlsConn := tls.Server(conn, l.config)
				if err := tlsConn.Handshake(); err != nil {
					conn.Close()
					return nil, err
				}
				conn = tlsConn
			}
			messages := NewMessages(conn, 0, 0)
			var context Context
			var err error
			if opened != nil {
				context, err = opened(messages)
				if err != nil {
					return nil, err
				}
			}
			return func() error {
				err := handler(context)
				if err == netpoll.EOF {
					messages.Close()
				}
				return err
			}, nil
		},
	}
	return netpoll.Serve(l.l, event)
}

func (l *TCPListener) Close() error {
	return l.l.Close()
}

func (l *TCPListener) Addr() net.Addr {
	return l.l.Addr()
}
