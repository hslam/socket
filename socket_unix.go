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

func (l *UNIXListener) Serve(event *netpoll.Event) error {
	if event == nil {
		return ErrEvent
	}
	return netpoll.Serve(l.l, event)
}

func (l *UNIXListener) ServeData(opened func(net.Conn) error, handler func(req []byte) (res []byte)) error {
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

func (l *UNIXListener) ServeConn(opened func(net.Conn) (Context, error), handler func(Context) error) error {
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

func (l *UNIXListener) ServeMessages(opened func(Messages) (Context, error), handler func(Context) error) error {
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

func (l *UNIXListener) Close() error {
	defer os.RemoveAll(l.address)
	return l.l.Close()
}

func (l *UNIXListener) Addr() net.Addr {
	return l.l.Addr()
}
