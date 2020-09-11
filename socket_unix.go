// Copyright (c) 2020 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package socket

import (
	"crypto/tls"
	"github.com/hslam/poll"
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

	return &UNIXListener{l: lis, config: t.Config}, err
}

type UNIXListener struct {
	l      *net.UnixListener
	config *tls.Config
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

func (l *UNIXListener) Serve(event *poll.Event) error {
	if event == nil {
		return ErrEvent
	}
	return poll.Serve(l.l, event)
}

func (l *UNIXListener) ServeConn(handle func(req []byte) (res []byte)) error {
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

func (l *UNIXListener) ServeMessages(handle func(req []byte) (res []byte)) error {
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

func (l *UNIXListener) Close() error {
	return l.l.Close()
}

func (l *UNIXListener) Addr() net.Addr {
	return l.l.Addr()
}
