// Copyright (c) 2020 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package socket

import (
	"bufio"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/hslam/poll"
	"io"
	"net"
	"net/http"
	"time"
)

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

func (c *HTTPConn) Messages() Messages {
	return NewMessages(c, 0, 0)
}

// NewHTTPSocket returns a new HTTP socket.
func NewHTTPSocket(config *tls.Config) Socket {
	return &HTTP{Config: config}
}

func (t *HTTP) Scheme() string {
	if t.Config == nil {
		return "http"
	}
	return "https"
}

func (t *HTTP) Dial(address string) (Conn, error) {
	var err error
	var conn net.Conn
	conn, err = net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}
	if t.Config != nil {
		t.Config.ServerName = address
		tlsConn := tls.Client(conn, t.Config)
		if err = tlsConn.Handshake(); err != nil {
			conn.Close()
			return nil, err
		}
		conn = tlsConn
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
	return &HTTPConn{conn}, nil
}

func (t *HTTP) Listen(address string) (Listener, error) {
	lis, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}
	return &HTTPListener{l: lis, config: t.Config}, nil
}

type HTTPListener struct {
	l      net.Listener
	config *tls.Config
}

func (l *HTTPListener) Accept() (Conn, error) {
	if conn, err := l.l.Accept(); err != nil {
		return nil, err
	} else {
		if l.config == nil {
			c := upgradeHTTPConn(conn)
			if c == nil {
				return nil, ErrConn
			}
			return &HTTPConn{c}, err
		}
		tlsConn := tls.Server(conn, l.config)
		if err = tlsConn.Handshake(); err != nil {
			conn.Close()
			return nil, err
		}
		c := upgradeHTTPConn(tlsConn)
		if c == nil {
			return nil, ErrConn
		}
		return &HTTPConn{c}, err
	}
}

func (l *HTTPListener) Serve(event *poll.Event) error {
	if event == nil {
		return ErrEvent
	}
	return poll.Serve(l.l, event)
}

func (l *HTTPListener) ServeData(opened func(net.Conn) error, handle func(req []byte) (res []byte)) error {
	event := &poll.Event{
		UpgradeConn: func(conn net.Conn) (upgrade net.Conn, err error) {
			if l.config != nil {
				tlsConn := tls.Server(conn, l.config)
				if err := tlsConn.Handshake(); err != nil {
					conn.Close()
					return nil, err
				}
				conn = tlsConn
			}
			c := upgradeHTTPConn(conn)
			if c == nil {
				return nil, ErrConn
			}
			upgrade = c
			if opened != nil {
				if err = opened(upgrade); err != nil {
					upgrade.Close()
					return
				}
			}
			return
		},
		Handle: handle,
	}
	return poll.Serve(l.l, event)
}

func (l *HTTPListener) ServeConn(opened func(net.Conn) (Context, error), handle func(Context) error) error {
	event := &poll.Event{
		UpgradeHandle: func(conn net.Conn) (func() error, error) {
			if l.config != nil {
				tlsConn := tls.Server(conn, l.config)
				if err := tlsConn.Handshake(); err != nil {
					conn.Close()
					return nil, err
				}
				conn = tlsConn
			}
			c := upgradeHTTPConn(conn)
			if c == nil {
				return nil, ErrConn
			}
			var context Context
			var err error
			if opened != nil {
				context, err = opened(c)
				if err != nil {
					return nil, err
				}
			}
			return func() error {
				err := handle(context)
				if err == poll.EOF {
					c.Close()
				}
				return err
			}, nil
		},
	}
	return poll.Serve(l.l, event)
}

func (l *HTTPListener) ServeMessages(opened func(Messages) (Context, error), handle func(Context) error) error {
	event := &poll.Event{
		UpgradeHandle: func(conn net.Conn) (func() error, error) {
			if l.config != nil {
				tlsConn := tls.Server(conn, l.config)
				if err := tlsConn.Handshake(); err != nil {
					conn.Close()
					return nil, err
				}
				conn = tlsConn
			}
			c := upgradeHTTPConn(conn)
			if c == nil {
				return nil, ErrConn
			}
			messages := NewMessages(c, 0, 0)
			var context Context
			var err error
			if opened != nil {
				context, err = opened(messages)
				if err != nil {
					return nil, err
				}
			}
			return func() error {
				err := handle(context)
				if err == poll.EOF {
					messages.Close()
				}
				return err
			}, nil
		},
	}
	return poll.Serve(l.l, event)
}

func (l *HTTPListener) Close() error {
	return l.l.Close()
}

func (l *HTTPListener) Addr() net.Addr {
	return l.l.Addr()
}

func upgradeHTTPConn(conn net.Conn) net.Conn {
	var b = bufio.NewReader(conn)
	req, err := http.ReadRequest(b)
	if err != nil {
		return nil
	}
	res := &response{conn: conn}
	return upgradeHttp(res, req)
}

func upgradeHttp(w http.ResponseWriter, r *http.Request) net.Conn {
	if r.Method != "CONNECT" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		io.WriteString(w, "405 must CONNECT\n")
		return nil

	}
	conn, _, err := w.(http.Hijacker).Hijack()
	if err != nil {
		return nil
	}
	io.WriteString(conn, "HTTP/1.0 "+HTTPConnected+"\n\n")
	return conn
}

type response struct {
	handlerHeader http.Header
	status        int
	conn          net.Conn
}

func (w *response) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return w.conn, bufio.NewReadWriter(bufio.NewReader(w.conn), bufio.NewWriter(w.conn)), nil
}

func (w *response) Header() http.Header {
	return w.handlerHeader
}

func (w *response) Write(data []byte) (n int, err error) {
	h := make([]byte, 0, 1024)
	h = append(h, fmt.Sprintf("HTTP/1.1 %03d %s\r\n", w.status, http.StatusText(w.status))...)
	h = append(h, fmt.Sprintf("Date: %s\r\n", time.Now().UTC().Format(http.TimeFormat))...)
	h = append(h, fmt.Sprintf("Content-Length: %d\r\n", len(data))...)
	h = append(h, "Content-Type: text/plain; charset=utf-8\r\n"...)
	h = append(h, "\r\n"...)
	h = append(h, data...)
	n, err = w.conn.Write(h)
	return len(data), err
}

func (w *response) WriteHeader(code int) {
	w.status = code
}
