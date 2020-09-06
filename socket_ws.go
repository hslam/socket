// Copyright (c) 2020 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package socket

import (
	"crypto/tls"
	"github.com/hslam/poll"
	"github.com/hslam/websocket"
	"net"
)

const (
	//WSPath defines the ws path.
	WSPath = "/"
)

type WS struct {
	Config *tls.Config
	Event  *poll.Event
}

type WSConn struct {
	*websocket.Conn
}

func (c *WSConn) Messages() Messages {
	return c.Conn
}

// NewWSSocket returns a new WS socket.
func NewWSSocket(config *tls.Config, event *poll.Event) Socket {
	return &WS{Config: config, Event: event}
}

func (t *WS) Scheme() string {
	return "ws"
}

func (t *WS) Dial(address string) (Conn, error) {
	var err error
	conn, err := websocket.Dial("tcp", address, WSPath, t.Config)
	if err != nil {
		return nil, err
	}
	return &WSConn{conn}, err
}

func (t *WS) Listen(address string) (Listener, error) {
	lis, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}
	return &WSListener{l: lis, config: t.Config, event: t.Event}, nil
}

type WSListener struct {
	l      net.Listener
	config *tls.Config
	event  *poll.Event
}

func (l *WSListener) Accept() (Conn, error) {
	if conn, err := l.l.Accept(); err != nil {
		return nil, err
	} else {
		if l.config == nil {
			ws := websocket.UpgradeConn(conn)
			if ws == nil {
				return nil, ErrConn
			}
			return &WSConn{ws}, err
		}
		tlsConn := tls.Server(conn, l.config)
		if err = tlsConn.Handshake(); err != nil {
			tlsConn.Close()
			return nil, err
		}
		ws := websocket.UpgradeConn(tlsConn)
		if ws == nil {
			return nil, ErrConn
		}
		return &WSConn{ws}, err
	}
}

func (l *WSListener) Serve() error {
	if l.event == nil {
		return ErrEvent
	}
	return poll.Serve(l.l, l.event)
}

func (l *WSListener) Close() error {
	return l.l.Close()
}

func (l *WSListener) Addr() net.Addr {
	return l.l.Addr()
}
