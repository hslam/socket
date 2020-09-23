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
var ErrHandler = errors.New("handler is nil")
var ErrOpened = errors.New("opened is nil")
var ErrServe = errors.New("serve is nil")
var ErrConn = errors.New("conn is nil")
var ErrNetwork = errors.New("network is not supported")

type Conn interface {
	net.Conn
	Messages() Messages
	Connection() net.Conn
}

type Dialer interface {
	Dial(address string) (Conn, error)
}

type Listener interface {
	Accept() (Conn, error)
	Close() error
	Addr() net.Addr
	Serve(handler netpoll.Handler) error
	ServeData(opened func(net.Conn) error, serve func(req []byte) (res []byte)) error
	ServeConn(opened func(net.Conn) (Context, error), serve func(Context) error) error
	ServeMessages(opened func(Messages) (Context, error), serve func(Context) error) error
}

type Context interface{}

type Socket interface {
	Scheme() string
	Dialer
	Listen(address string) (Listener, error)
}

func Address(s Socket, url string) (string, error) {
	if !strings.HasPrefix(url, s.Scheme()+"://") {
		return url, errors.New("error url:" + url)
	}
	return url[len(s.Scheme()+"://"):], nil
}
func Url(s Socket, addr string) string {
	return fmt.Sprintf("%s://%s", s.Scheme(), addr)
}

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
