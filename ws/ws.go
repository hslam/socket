// Copyright (c) 2020 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package ws

import (
	"errors"
	"github.com/hslam/socket"
	"github.com/hslam/websocket"
	"net/http"
	"runtime"
)

var numCPU = runtime.NumCPU()

const (
	//WSPath defines the ws path.
	WSPath = "/"
)

type WS struct {
	conn chan *WSConn
}

type WSConn struct {
	*websocket.Conn
}

func (c *WSConn) Messages() socket.Messages {
	return &Messages{WSConn: c}
}

// NewSocket returns a new WS socket.
func NewSocket() socket.Socket {
	return &WS{}
}

func (t *WS) Scheme() string {
	return "ws"
}

func (t *WS) Dial(address string) (socket.Conn, error) {
	var err error
	conn, err := websocket.Dial(address, WSPath)
	if err != nil {
		return nil, err
	}
	return &WSConn{conn}, err
}

func (t *WS) Listen(address string) (socket.Listener, error) {
	httpServer := &http.Server{
		Addr:    address,
		Handler: websocket.Handler(t.Serve),
	}
	go httpServer.ListenAndServe()
	t.conn = make(chan *WSConn, numCPU*512)
	return &WSListener{httpServer: httpServer, conn: t.conn, addr: &Address{t.Scheme(), address}}, nil
}

func (t *WS) Serve(ws *websocket.Conn) {
	wsConn := &WSConn{ws}
	t.conn <- wsConn
}

type WSListener struct {
	httpServer *http.Server
	conn       chan *WSConn
	addr       socket.Addr
}

type Address struct {
	network string
	address string
}

func (a *Address) Network() string {
	return a.network
}

func (a *Address) String() string {
	return a.address
}

func (l *WSListener) Accept() (socket.Conn, error) {
	if conn, ok := <-l.conn; ok {
		return conn, nil
	}
	return nil, errors.New("http: Server closed")
}

func (l *WSListener) Close() error {
	return l.httpServer.Close()
}

func (l *WSListener) Addr() socket.Addr {
	return l.addr
}
