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
			ws, err := websocket.Upgrade(conn)
			if err != nil {
				return nil, err
			}
			return &WSConn{ws}, err
		}
		tlsConn := tls.Server(conn, l.config)
		if err = tlsConn.Handshake(); err != nil {
			conn.Close()
			return nil, err
		}
		ws, err := websocket.Upgrade(tlsConn)
		if err != nil {
			return nil, err
		}
		return &WSConn{ws}, err
	}
}

func (l *WSListener) Serve(handler netpoll.Handler) error {
	if handler == nil {
		return ErrHandler
	}
	return netpoll.Serve(l.l, handler)
}

func (l *WSListener) ServeData(opened func(net.Conn) error, serve func(req []byte) (res []byte)) error {
	if opened == nil {
		return ErrOpened
	} else if serve == nil {
		return ErrServe
	}
	Upgrade := func(conn net.Conn) (netpoll.Context, error) {
		if l.config != nil {
			tlsConn := tls.Server(conn, l.config)
			if err := tlsConn.Handshake(); err != nil {
				conn.Close()
				return nil, err
			}
			conn = tlsConn
		}
		messages, err := websocket.Upgrade(conn)
		if err != nil {
			conn.Close()
			return nil, err
		}
		opened(messages)
		return messages, nil
	}
	Serve := func(context netpoll.Context) error {
		ws := context.(*websocket.Conn)
		msg, err := ws.ReadMessage()
		if err != nil {
			return err
		}
		res := serve(msg)
		if len(res) == 0 {
			return nil
		}
		return ws.WriteMessage(res)
	}
	return netpoll.Serve(l.l, netpoll.NewHandler(Upgrade, Serve))
}

func (l *WSListener) ServeConn(opened func(net.Conn) (Context, error), serve func(Context) error) error {
	if opened == nil {
		return ErrOpened
	} else if serve == nil {
		return ErrServe
	}
	Upgrade := func(conn net.Conn) (netpoll.Context, error) {
		if l.config != nil {
			tlsConn := tls.Server(conn, l.config)
			if err := tlsConn.Handshake(); err != nil {
				conn.Close()
				return nil, err
			}
			conn = tlsConn
		}
		messages, err := websocket.Upgrade(conn)
		if err != nil {
			conn.Close()
			return nil, err
		}
		return opened(messages)
	}
	Serve := func(context netpoll.Context) error {
		return serve(context)
	}
	return netpoll.Serve(l.l, netpoll.NewHandler(Upgrade, Serve))
}

func (l *WSListener) ServeMessages(opened func(Messages) (Context, error), serve func(Context) error) error {
	if opened == nil {
		return ErrOpened
	} else if serve == nil {
		return ErrServe
	}
	Upgrade := func(conn net.Conn) (netpoll.Context, error) {
		if l.config != nil {
			tlsConn := tls.Server(conn, l.config)
			if err := tlsConn.Handshake(); err != nil {
				conn.Close()
				return nil, err
			}
			conn = tlsConn
		}
		messages, err := websocket.Upgrade(conn)
		if err != nil {
			conn.Close()
			return nil, err
		}
		return opened(messages)
	}
	Serve := func(context netpoll.Context) error {
		return serve(context)
	}
	return netpoll.Serve(l.l, netpoll.NewHandler(Upgrade, Serve))
}

func (l *WSListener) Close() error {
	return l.l.Close()
}

func (l *WSListener) Addr() net.Addr {
	return l.l.Addr()
}
