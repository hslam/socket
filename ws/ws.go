package ws

import (
	"crypto/tls"
	"errors"
	"golang.org/x/net/websocket"
	"github.com/hslam/socket"
	"io"
	"net/http"
	"runtime"
)

var numCPU = runtime.NumCPU()

const (
	//HTTPPath defines the http path.
	HTTPPath = "/"
)

type WS struct {
	conn   chan *WSConn
	Config *tls.Config
}

type WSConn struct {
	io.ReadWriteCloser
	close chan bool
}

func (c *WSConn) Messages() socket.Messages {
	return socket.NewMessages(c, 0, 0)
}

func (c *WSConn) Close() error {
	c.close <- true
	return c.ReadWriteCloser.Close()
}

// NewSocket returns a new WS socket.
func NewSocket() socket.Socket {
	return &WS{}
}

func NewTLSSocket(config *tls.Config) socket.Socket {
	return &WS{Config: config}
}

func (t *WS) Scheme() string {
	if t.Config == nil {
		return "ws"
	}
	return "wss"
}

func (t *WS) Dial(address string) (socket.Conn, error) {
	var err error
	origin := "http://localhost/"
	url := "ws://" + address + HTTPPath
	conn, err := websocket.Dial(url, "", origin)
	if err != nil {
		return nil, err
	}
	if t.Config == nil {
		return &WSConn{conn, make(chan bool, 10)}, err
	}
	t.Config.ServerName = address
	tlsConn := tls.Client(conn, t.Config)
	if err = tlsConn.Handshake(); err != nil {
		tlsConn.Close()
		return nil, err
	}
	return &WSConn{conn, make(chan bool, 10)}, err
}

func (t *WS) Listen(address string) (socket.Listener, error) {
	httpServer := &http.Server{
		Addr: address,
	}
	httpServer.Handler = websocket.Handler(t.Conn)
	go httpServer.ListenAndServe()
	t.conn = make(chan *WSConn, numCPU*512)
	return &WSListener{httpServer: httpServer, conn: t.conn}, nil
}

func (t *WS) Conn(ws *websocket.Conn) {
	var wsConn *WSConn
	if t.Config == nil {

		wsConn = &WSConn{ws, make(chan bool, 10)}
	} else {
		tlsConn := tls.Server(ws, t.Config)
		if err := tlsConn.Handshake(); err != nil {
			tlsConn.Close()
			return
		}
		wsConn = &WSConn{tlsConn, make(chan bool, 10)}
	}
	t.conn <- wsConn
	<-wsConn.close
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
