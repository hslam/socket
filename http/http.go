// Copyright (c) 2020 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package http

import (
	"bufio"
	"crypto/tls"
	"errors"
	"github.com/hslam/socket"
	"io"
	"net"
	"net/http"
	"runtime"
)

var numCPU = runtime.NumCPU()

const (
	//HTTPConnected defines the http connected.
	HTTPConnected = "200 Connected to Server"
	//HTTPPath defines the http path.
	HTTPPath = "/"
)

type HTTP struct {
	Config *tls.Config
}

type HTTPConn struct {
	net.Conn
}

func (c *HTTPConn) Messages() socket.Messages {
	return socket.NewMessages(c, 0, 0)
}

// NewSocket returns a new HTTP socket.
func NewSocket() socket.Socket {
	return &HTTP{}
}

func NewTLSSocket(config *tls.Config) socket.Socket {
	return &HTTP{Config: config}
}

func (t *HTTP) Scheme() string {
	if t.Config == nil {
		return "http"
	}
	return "https"
}

func (t *HTTP) Dial(address string) (socket.Conn, error) {
	var err error
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}
	io.WriteString(conn, "CONNECT "+HTTPPath+" HTTP/1.1\n\n")
	resp, err := http.ReadResponse(bufio.NewReader(conn), &http.Request{Method: "CONNECT"})
	if err != nil || resp.Status != HTTPConnected {
		if err == nil {
			err = errors.New("unexpected HTTP response: " + resp.Status)
		}
		conn.Close()
		return nil, &net.OpError{
			Op:   "dial-http",
			Net:  "tcp" + " " + address,
			Addr: nil,
			Err:  err,
		}
	}
	if t.Config == nil {
		return &HTTPConn{conn}, err
	}
	t.Config.ServerName = address
	tlsConn := tls.Client(conn, t.Config)
	if err = tlsConn.Handshake(); err != nil {
		tlsConn.Close()
		return nil, err
	}
	return &HTTPConn{tlsConn}, err
}

func (t *HTTP) Listen(address string) (socket.Listener, error) {
	httpServer := &http.Server{
		Addr: address,
	}
	h := new(handler)
	h.conn = make(chan net.Conn, numCPU*512)
	httpServer.Handler = h
	go httpServer.ListenAndServe()
	return &HTTPListener{httpServer: httpServer, handler: h, config: t.Config, addr: &Address{t.Scheme(), address}}, nil
}

type HTTPListener struct {
	httpServer *http.Server
	handler    *handler
	config     *tls.Config
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

func (l *HTTPListener) Accept() (socket.Conn, error) {
	if conn, ok := <-l.handler.conn; ok {
		if l.config == nil {
			return &HTTPConn{conn}, nil
		}
		tlsConn := tls.Server(conn, l.config)
		if err := tlsConn.Handshake(); err != nil {
			tlsConn.Close()
			return nil, err
		}
		return &HTTPConn{tlsConn}, nil
	}
	return nil, errors.New("http: Server closed")
}

func (l *HTTPListener) Close() error {
	return l.httpServer.Close()
}

func (l *HTTPListener) Addr() socket.Addr {
	return l.addr
}

type handler struct {
	conn chan net.Conn
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if r.Method == "CONNECT" {
		conn, _, err := w.(http.Hijacker).Hijack()
		if err != nil {
			return
		}
		io.WriteString(conn, "HTTP/1.0 "+HTTPConnected+"\n\n")
		h.conn <- conn
	} else {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusMethodNotAllowed)
		io.WriteString(w, "405 must CONNECT\n")
	}
}
