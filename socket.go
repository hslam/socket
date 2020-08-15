// Copyright (c) 2020 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package socket

import (
	"errors"
	"fmt"
	"io"
	"strings"
)

type Conn interface {
	io.ReadWriteCloser
	Messages() Messages
}

type Dialer interface {
	Dial(address string) (Conn, error)
}

type Listener interface {
	// Accept waits for and returns the next connection to the listener.
	Accept() (Conn, error)

	// Close closes the listener.
	// Any blocked Accept operations will be unblocked and return errors.
	Close() error
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
