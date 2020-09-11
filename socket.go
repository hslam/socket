// Copyright (c) 2020 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package socket

import (
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/hslam/poll"
	"net"
	"runtime"
	"strings"
)

var numCPU = runtime.NumCPU()
var ErrEvent = errors.New("event is nil")
var ErrConn = errors.New("conn is nil")

type Conn interface {
	net.Conn
	Messages() Messages
}

type Messages interface {
	SetBatch(batch func() int)
	ReadMessage() ([]byte, error)
	WriteMessage([]byte) error
	Close() error
}

type Dialer interface {
	Dial(address string) (Conn, error)
}

type Listener interface {
	Accept() (Conn, error)
	Close() error
	Addr() net.Addr
	Serve(event *poll.Event) error
	ServeConn(handle func(req []byte) (res []byte)) error
	ServeMessages(handle func(req []byte) (res []byte)) error
}

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
		return nil, errors.New("not supported")
	}
}
