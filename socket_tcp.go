// Copyright (c) 2020 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package socket

import (
	"crypto/tls"
	"github.com/hslam/netpoll"
	"net"
)

// TCP implements the Socket interface.
type TCP struct {
	Config *tls.Config
}

// TCPConn implements the Conn interface.
type TCPConn struct {
	net.Conn
}

// Messages returns a new Messages.
func (c *TCPConn) Messages() Messages {
	return NewMessages(c.Conn, false, 0, 0)
}

// Connection returns the net.Conn.
func (c *TCPConn) Connection() net.Conn {
	return c.Conn
}

// NewTCPSocket returns a new TCP socket.
func NewTCPSocket(config *tls.Config) Socket {
	return &TCP{Config: config}
}

// Scheme returns the socket's scheme.
func (t *TCP) Scheme() string {
	if t.Config == nil {
		return "tcp"
	}
	return "tcps"
}

// Dial connects to an address.
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

// Listen announces on the local address.
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

// TCPListener implements the Listener interface.
type TCPListener struct {
	l      *net.TCPListener
	server *netpoll.Server
	config *tls.Config
}

// Accept waits for and returns the next connection to the listener.
func (l *TCPListener) Accept() (Conn, error) {
	conn, err := l.l.AcceptTCP()
	if err != nil {
		return nil, err
	}
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

// Serve serves the netpoll.Handler by the netpoll.
func (l *TCPListener) Serve(handler netpoll.Handler) error {
	if handler == nil {
		return ErrHandler
	}
	l.server = &netpoll.Server{
		Handler: handler,
	}
	return l.server.Serve(l.l)
}

// ServeData serves the opened func and the serve func by the netpoll.
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
	l.server = &netpoll.Server{
		Handler: netpoll.NewHandler(Upgrade, Serve),
	}
	return l.server.Serve(l.l)
}

// ServeConn serves the opened func and the serve func by the netpoll.
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
	l.server = &netpoll.Server{
		Handler: netpoll.NewHandler(Upgrade, Serve),
	}
	return l.server.Serve(l.l)
}

// ServeMessages serves the opened func and the serve func by the netpoll.
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
		messages := NewMessages(conn, false, 0, 0)
		return opened(messages)
	}
	Serve := func(context netpoll.Context) error {
		return serve(context)
	}
	l.server = &netpoll.Server{
		Handler: netpoll.NewHandler(Upgrade, Serve),
	}
	return l.server.Serve(l.l)
}

// Close closes the listener.
func (l *TCPListener) Close() error {
	if l.server != nil {
		return l.server.Close()
	}
	return l.l.Close()
}

// Addr returns the listener's network address.
func (l *TCPListener) Addr() net.Addr {
	return l.l.Addr()
}
