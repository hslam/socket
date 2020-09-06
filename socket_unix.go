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
	Event  *poll.Event
}

type UNIXConn struct {
	net.Conn
}

func (c *UNIXConn) Messages() Messages {
	return NewMessages(c, 0, 0)
}

// NewUNIXSocket returns a new UNIX socket.
func NewUNIXSocket(config *tls.Config, event *poll.Event) Socket {
	return &UNIX{Config: config, Event: event}
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
		tlsConn.Close()
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

	return &UNIXListener{l: lis, config: t.Config, event: t.Event}, err
}

type UNIXListener struct {
	l      *net.UnixListener
	config *tls.Config
	event  *poll.Event
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
			tlsConn.Close()
			return nil, err
		}
		return &UNIXConn{tlsConn}, err
	}
}

func (l *UNIXListener) Serve() error {
	if l.event == nil {
		return ErrEvent
	}
	return poll.Serve(l.l, l.event)
}

func (l *UNIXListener) Close() error {
	return l.l.Close()
}

func (l *UNIXListener) Addr() net.Addr {
	return l.l.Addr()
}
