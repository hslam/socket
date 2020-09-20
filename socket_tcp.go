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

func (l *TCPListener) Serve(handler netpoll.Handler) error {
	if handler == nil {
		return ErrHandler
	}
	return netpoll.Serve(l.l, handler)
}

func (l *TCPListener) ServeData(opened func(net.Conn) error, serve func(req []byte) (res []byte)) error {
	if serve == nil {
		return ErrServe
	}
	type Context struct {
		Conn net.Conn
		buf  []byte
	}
	Upgrade := func(conn net.Conn) (netpoll.Context, error) {
		if l.config != nil {
			tlsConn := tls.Server(conn, l.config)
			if err := tlsConn.Handshake(); err != nil {
				conn.Close()
				return nil, err
			}
			conn = tlsConn
		}
		if opened != nil {
			if err := opened(conn); err != nil {
				conn.Close()
				return nil, err
			}
		}
		ctx := &Context{
			Conn: conn,
			buf:  make([]byte, 1024*64),
		}
		return ctx, nil
	}
	Serve := func(context netpoll.Context) error {
		c := context.(*Context)
		n, err := c.Conn.Read(c.buf)
		if err != nil {
			return err
		}
		res := serve(c.buf[:n])
		if len(res) == 0 {
			return nil
		}
		_, err = c.Conn.Write(res)
		return err
	}
	return netpoll.Serve(l.l, netpoll.NewHandler(Upgrade, Serve))
}

func (l *TCPListener) ServeConn(opened func(net.Conn) (Context, error), serve func(Context) error) error {
	if opened == nil {
		return ErrOpened
	} else if serve == nil {
		return ErrServe
	}
	Upgrade := func(conn net.Conn) (netpoll.Context, error) {
		if l.config != nil {
			tlsConn := tls.Server(conn, l.config)
			if err := tlsConn.Handshake(); err != nil {
				conn.Close()
				return nil, err
			}
			conn = tlsConn
		}
		return opened(conn)
	}
	Serve := func(context netpoll.Context) error {
		return serve(context)
	}
	return netpoll.Serve(l.l, netpoll.NewHandler(Upgrade, Serve))
}

func (l *TCPListener) ServeMessages(opened func(Messages) (Context, error), serve func(Context) error) error {
	if opened == nil {
		return ErrOpened
	} else if serve == nil {
		return ErrServe
	}
	Upgrade := func(conn net.Conn) (netpoll.Context, error) {
		if l.config != nil {
			tlsConn := tls.Server(conn, l.config)
			if err := tlsConn.Handshake(); err != nil {
				conn.Close()
				return nil, err
			}
			conn = tlsConn
		}
		messages := NewMessages(conn, 0, 0)
		return opened(messages)
	}
	Serve := func(context netpoll.Context) error {
		return serve(context)
	}
	return netpoll.Serve(l.l, netpoll.NewHandler(Upgrade, Serve))
}

func (l *TCPListener) Close() error {
	return l.l.Close()
}

func (l *TCPListener) Addr() net.Addr {
	return l.l.Addr()
}
