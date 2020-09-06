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
	Serve() error
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

func NewSocket(network string, config *tls.Config, event *poll.Event) (Socket, error) {
	switch network {
	case "tcp", "tcps":
		return NewTCPSocket(config, event), nil
	case "unix", "unixs":
		return NewUNIXSocket(config, event), nil
	case "http", "https":
		return NewHTTPSocket(config, event), nil
	case "ws", "wss":
		return NewWSSocket(config, event), nil
	default:
		return nil, errors.New("not supported")
	}
}
