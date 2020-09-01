// Copyright (c) 2020 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package socket

import (
	"errors"
	"fmt"
	"net"
	"strings"
)

type Conn interface {
	net.Conn
	Messages() Messages
}

type Messages interface {
	SetBatch(batch Batch)
	ReadMessage() ([]byte, error)
	WriteMessage([]byte) error
	Close() error
}

type Batch interface {
	Concurrency() int
}

type Dialer interface {
	Dial(address string) (Conn, error)
}

type Listener interface {
	Accept() (Conn, error)
	Close() error
	Addr() net.Addr
}

type Socket interface {
	Scheme() string
	Dialer
	Listen(address string) (Listener, error)
}

type NewSocket func() Socket

func Address(s Socket, url string) (string, error) {
	if !strings.HasPrefix(url, s.Scheme()+"://") {
		return url, errors.New("error url:" + url)
	}
	return url[len(s.Scheme()+"://"):], nil
}
func Url(s Socket, addr string) string {
	return fmt.Sprintf("%s://%s", s.Scheme(), addr)
}
