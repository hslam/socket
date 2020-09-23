// Copyright (c) 2020 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package socket

import (
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/hslam/netpoll"
	"net"
	"runtime"
	"strings"
)

var numCPU = runtime.NumCPU()

// ErrHandler is the error when the handler is nil
var ErrHandler = errors.New("handler is nil")

// ErrOpened is the error when the opened is nil
var ErrOpened = errors.New("opened is nil")

// ErrServe is the error when the serve is nil
var ErrServe = errors.New("serve is nil")

// ErrConn is the error when the conn is nil
var ErrConn = errors.New("conn is nil")

// ErrNetwork is the error when the network is not supported
var ErrNetwork = errors.New("network is not supported")

// Conn is a generic stream-oriented network connection.
type Conn interface {
	net.Conn
	// Messages returns a new Messages.
	Messages() Messages
	// Connection returns the net.Conn.
	Connection() net.Conn
}

// Dialer is a generic network dialer for stream-oriented protocols.
type Dialer interface {
	// Dial connects to an address.
	Dial(address string) (Conn, error)
}

// Listener is a generic network listener for stream-oriented protocols.
type Listener interface {
	// Accept waits for and returns the next connection to the listener.
	Accept() (Conn, error)
	// Close closes the listener.
	// Any blocked Accept operations will be unblocked and return errors.
	Close() error
	// Addr returns the listener's network address.
	Addr() net.Addr
	// Serve serves the netpoll.Handler by the netpoll.
	Serve(handler netpoll.Handler) error
	// ServeData serves the opened func and the serve func by the netpoll.
	ServeData(opened func(net.Conn) error, serve func(req []byte) (res []byte)) error
	// ServeConn serves the opened func and the serve func by the netpoll.
	ServeConn(opened func(net.Conn) (Context, error), serve func(Context) error) error
	// ServeMessages serves the opened func and the serve func by the netpoll.
	ServeMessages(opened func(Messages) (Context, error), serve func(Context) error) error
}

// Context represents a context.
type Context interface{}

// Socket contains the Dialer and the Listener.
type Socket interface {
	// Scheme returns the socket's scheme.
	Scheme() string
	Dialer
	// Listen announces on the local address.
	Listen(address string) (Listener, error)
}

// Address returns the socket's address by a url.
func Address(s Socket, url string) (string, error) {
	if !strings.HasPrefix(url, s.Scheme()+"://") {
		return url, errors.New("error url:" + url)
	}
	return url[len(s.Scheme()+"://"):], nil
}

// URL returns the socket's url by a address.
func URL(s Socket, addr string) string {
	return fmt.Sprintf("%s://%s", s.Scheme(), addr)
}

// NewSocket returns a new socket by a network and a TLS config.
func NewSocket(network string, config *tls.Config) (Socket, error) {
	switch network {
	case "tcp", "tcps":
		return NewTCPSocket(config), nil
	case "unix", "unixs":
		return NewUNIXSocket(config), nil
	case "http", "https":
		return NewHTTPSocket(config), nil
	case "ws", "wss":
		return NewWSSocket(config), nil
	default:
		return nil, ErrNetwork
	}
}
