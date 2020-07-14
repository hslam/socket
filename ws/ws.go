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
	return socket.NewMessages(c, 0, 0)
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
	return &WSListener{httpServer: httpServer, conn: t.conn}, nil
}

func (t *WS) Serve(ws *websocket.Conn) {
	wsConn := &WSConn{ws}
	t.conn <- wsConn
}

type WSListener struct {
	httpServer *http.Server
	conn       chan *WSConn
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
