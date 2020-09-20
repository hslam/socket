// Copyright (c) 2020 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package socket

import (
	"crypto/tls"
	"github.com/hslam/netpoll"
	"net"
	"os"
)

type UNIX struct {
	Config *tls.Config
}

type UNIXConn struct {
	net.Conn
}

func (c *UNIXConn) Messages() Messages {
	return NewMessages(c, 0, 0)
}

// NewUNIXSocket returns a new UNIX socket.
func NewUNIXSocket(config *tls.Config) Socket {
	return &UNIX{Config: config}
}

func (t *UNIX) Scheme() string {
	if t.Config == nil {
		return "unix"
	}
	return "unixs"
}

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

type UNIXListener struct {
	l       *net.UnixListener
	config  *tls.Config
	address string
}

func (l *UNIXListener) Accept() (Conn, error) {
	if conn, err := l.l.Accept(); err != nil {
		return nil, err
	} else {
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
}

func (l *UNIXListener) Serve(handler netpoll.Handler) error {
	if handler == nil {
		return ErrHandler
	}
	return netpoll.Serve(l.l, handler)
}

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
	return netpoll.Serve(l.l, netpoll.NewHandler(Upgrade, Serve))
}

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
	return netpoll.Serve(l.l, netpoll.NewHandler(Upgrade, Serve))
}

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
		messages := NewMessages(conn, 0, 0)
		return opened(messages)
	}
	Serve := func(context netpoll.Context) error {
		return serve(context)
	}
	return netpoll.Serve(l.l, netpoll.NewHandler(Upgrade, Serve))
}

func (l *UNIXListener) Close() error {
	defer os.RemoveAll(l.address)
	return l.l.Close()
}

func (l *UNIXListener) Addr() net.Addr {
	return l.l.Addr()
}
