// Copyright (c) 2020 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package socket

import (
	"crypto/tls"
	"github.com/hslam/netpoll"
	"net"
	"os"
)

// UNIX implements the Socket interface.
type UNIX struct {
	Config *tls.Config
}

// UNIXConn implements the Conn interface.
type UNIXConn struct {
	net.Conn
}

// Messages returns a new Messages.
func (c *UNIXConn) Messages() Messages {
	return NewMessages(c.Conn, false, 0, 0)
}

// Connection returns the net.Conn.
func (c *UNIXConn) Connection() net.Conn {
	return c.Conn
}

// NewUNIXSocket returns a new UNIX socket.
func NewUNIXSocket(config *tls.Config) Socket {
	return &UNIX{Config: config}
}

// Scheme returns the socket's scheme.
func (t *UNIX) Scheme() string {
	if t.Config == nil {
		return "unix"
	}
	return "unixs"
}

// Dial connects to an address.
func (t *UNIX) Dial(address string) (Conn, error) {
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
		conn.Close()
		return nil, err
	}
	return &UNIXConn{tlsConn}, err
}

// Listen announces on the local address.
func (t *UNIX) Listen(address string) (Listener, error) {
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

	return &UNIXListener{l: lis, config: t.Config, address: address}, err
}

// UNIXListener implements the Listener interface.
type UNIXListener struct {
	l       *net.UnixListener
	server  *netpoll.Server
	config  *tls.Config
	address string
}

// Accept waits for and returns the next connection to the listener.
func (l *UNIXListener) Accept() (Conn, error) {
	conn, err := l.l.Accept()
	if err != nil {
		return nil, err
	}
	if l.config == nil {
		return &UNIXConn{conn}, err
	}
	tlsConn := tls.Server(conn, l.config)
	if err = tlsConn.Handshake(); err != nil {
		conn.Close()
		return nil, err
	}
	return &UNIXConn{tlsConn}, err
}

// Serve serves the netpoll.Handler by the netpoll.
func (l *UNIXListener) Serve(handler netpoll.Handler) error {
	if handler == nil {
		return ErrHandler
	}
	l.server = &netpoll.Server{
		Handler: handler,
	}
	return l.server.Serve(l.l)
}

// ServeData serves the opened func and the serve func by the netpoll.
func (l *UNIXListener) ServeData(opened func(net.Conn) error, serve func(req []byte) (res []byte)) error {
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
func (l *UNIXListener) ServeConn(opened func(net.Conn) (Context, error), serve func(Context) error) error {
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
func (l *UNIXListener) ServeMessages(opened func(Messages) (Context, error), serve func(Context) error) error {
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
func (l *UNIXListener) Close() error {
	defer os.RemoveAll(l.address)
	if l.server != nil {
		return l.server.Close()
	}
	return l.l.Close()
}

// Addr returns the listener's network address.
func (l *UNIXListener) Addr() net.Addr {
	return l.l.Addr()
}
