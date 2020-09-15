// Copyright (c) 2020 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package socket

import (
	"crypto/tls"
	"github.com/hslam/netpoll"
	"github.com/hslam/websocket"
	"net"
)

const (
	//WSPath defines the ws path.
	WSPath = "/"
)

type WS struct {
	Config *tls.Config
}

type WSConn struct {
	*websocket.Conn
}

func (c *WSConn) Messages() Messages {
	return c.Conn
}

// NewWSSocket returns a new WS socket.
func NewWSSocket(config *tls.Config) Socket {
	return &WS{Config: config}
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
	return &WSListener{l: lis, config: t.Config}, nil
}

type WSListener struct {
	l      net.Listener
	config *tls.Config
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
			conn.Close()
			return nil, err
		}
		ws := websocket.UpgradeConn(tlsConn)
		if ws == nil {
			return nil, ErrConn
		}
		return &WSConn{ws}, err
	}
}

func (l *WSListener) Serve(event *netpoll.Event) error {
	if event == nil {
		return ErrEvent
	}
	return netpoll.Serve(l.l, event)
}

func (l *WSListener) ServeData(opened func(net.Conn) error, handler func(req []byte) (res []byte)) error {
	event := &netpoll.Event{
		UpgradeHandler: func(conn net.Conn) (func() error, error) {
			if l.config != nil {
				tlsConn := tls.Server(conn, l.config)
				if err := tlsConn.Handshake(); err != nil {
					conn.Close()
					return nil, err
				}
				conn = tlsConn
			}
			ws := websocket.UpgradeConn(conn)
			if ws == nil {
				return nil, ErrConn
			}
			if opened != nil {
				if err := opened(ws); err != nil {
					ws.Close()
					return nil, ErrConn
				}
			}
			return func() error {
				req, err := ws.ReadMessage()
				if err != nil {
					return err
				}
				res := handler(req)
				if len(res) > 0 {
					err = ws.WriteMessage(res)
				}
				return err
			}, nil
		},
	}
	return netpoll.Serve(l.l, event)
}

func (l *WSListener) ServeConn(opened func(net.Conn) (Context, error), handler func(Context) error) error {
	event := &netpoll.Event{
		UpgradeHandler: func(conn net.Conn) (func() error, error) {
			if l.config != nil {
				tlsConn := tls.Server(conn, l.config)
				if err := tlsConn.Handshake(); err != nil {
					conn.Close()
					return nil, err
				}
				conn = tlsConn
			}
			ws := websocket.UpgradeConn(conn)
			if ws == nil {
				return nil, ErrConn
			}
			var context Context
			var err error
			if opened != nil {
				context, err = opened(ws)
				if err != nil {
					return nil, err
				}
			}
			return func() error {
				err := handler(context)
				if err == netpoll.EOF {
					ws.Close()
				}
				return err
			}, nil
		},
	}
	return netpoll.Serve(l.l, event)
}

func (l *WSListener) ServeMessages(opened func(Messages) (Context, error), handler func(Context) error) error {
	event := &netpoll.Event{
		UpgradeHandler: func(conn net.Conn) (func() error, error) {
			if l.config != nil {
				tlsConn := tls.Server(conn, l.config)
				if err := tlsConn.Handshake(); err != nil {
					conn.Close()
					return nil, err
				}
				conn = tlsConn
			}
			ws := websocket.UpgradeConn(conn)
			if ws == nil {
				return nil, ErrConn
			}
			var context Context
			var err error
			if opened != nil {
				context, err = opened(ws)
				if err != nil {
					return nil, err
				}
			}
			return func() error {
				err := handler(context)
				if err == netpoll.EOF {
					ws.Close()
				}
				return err
			}, nil
		},
	}
	return netpoll.Serve(l.l, event)
}

func (l *WSListener) Close() error {
	return l.l.Close()
}

func (l *WSListener) Addr() net.Addr {
	return l.l.Addr()
}
