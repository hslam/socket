// Copyright (c) 2021 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package socket

import (
	"crypto/tls"
	"errors"
	"github.com/hslam/inproc"
	"github.com/hslam/netpoll"
	"net"
)

// INPROC implements the Socket interface.
type INPROC struct {
	Config *tls.Config
}

// INPROConn implements the Conn interface.
type INPROConn struct {
	net.Conn
}

// Messages returns a new Messages.
func (c *INPROConn) Messages() Messages {
	return NewMessages(c.Conn, false, 0, 0)
}

// Connection returns the net.Conn.
func (c *INPROConn) Connection() net.Conn {
	return c.Conn
}

// NewINPROCSocket returns a new TCP socket.
func NewINPROCSocket(config *tls.Config) Socket {
	return &INPROC{Config: config}
}

// Scheme returns the socket's scheme.
func (t *INPROC) Scheme() string {
	if t.Config == nil {
		return "inproc"
	}
	return "inprocs"
}

// Dial connects to an address.
func (t *INPROC) Dial(address string) (Conn, error) {
	conn, err := inproc.Dial(address)
	if err != nil {
		return nil, err
	}
	if t.Config == nil {
		return &INPROConn{conn}, err
	}
	t.Config.ServerName = address
	tlsConn := tls.Client(conn, t.Config)
	if err = tlsConn.Handshake(); err != nil {
		conn.Close()
		return nil, err
	}
	return &INPROConn{tlsConn}, err
}

// Listen announces on the local address.
func (t *INPROC) Listen(address string) (Listener, error) {
	lis, err := inproc.Listen(address)
	if err != nil {
		return nil, err
	}
	return &INPROCListener{l: lis, config: t.Config}, err
}

// INPROCListener implements the Listener interface.
type INPROCListener struct {
	l      net.Listener
	config *tls.Config
}

// Accept waits for and returns the next connection to the listener.
func (l *INPROCListener) Accept() (Conn, error) {
	conn, err := l.l.Accept()
	if err != nil {
		return nil, err
	}
	if l.config == nil {
		return &INPROConn{conn}, err
	}
	tlsConn := tls.Server(conn, l.config)
	if err = tlsConn.Handshake(); err != nil {
		conn.Close()
		return nil, err
	}
	return &INPROConn{tlsConn}, err
}

// Serve serves the netpoll.Handler by the netpoll.
func (l *INPROCListener) Serve(handler netpoll.Handler) error {
	return errors.New("not pollable")
}

// ServeData serves the opened func and the serve func by the netpoll.
func (l *INPROCListener) ServeData(opened func(net.Conn) error, serve func(req []byte) (res []byte)) error {
	return errors.New("not pollable")
}

// ServeConn serves the opened func and the serve func by the netpoll.
func (l *INPROCListener) ServeConn(opened func(net.Conn) (Context, error), serve func(Context) error) error {
	return errors.New("not pollable")
}

// ServeMessages serves the opened func and the serve func by the netpoll.
func (l *INPROCListener) ServeMessages(opened func(Messages) (Context, error), serve func(Context) error) error {
	return errors.New("not pollable")
}

// Close closes the listener.
func (l *INPROCListener) Close() error {
	return l.l.Close()
}

// Addr returns the listener's network address.
func (l *INPROCListener) Addr() net.Addr {
	return l.l.Addr()
}
