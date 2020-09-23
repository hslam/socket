// Copyright (c) 2020 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package socket

import (
	"errors"
	"github.com/hslam/netpoll"
	"net"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestSocket(t *testing.T) {
	testSocket(NewTCPSocket(nil), NewTCPSocket(nil), "tcp", t)
	testSocket(NewUNIXSocket(nil), NewUNIXSocket(nil), "unix", t)
	testSocket(NewHTTPSocket(nil), NewHTTPSocket(nil), "http", t)
	testSocket(NewWSSocket(nil), NewWSSocket(nil), "ws", t)
	testSocket(NewTCPSocket(DefalutTLSConfig()), NewTCPSocket(SkipVerifyTLSConfig()), "tcps", t)
	testSocket(NewUNIXSocket(DefalutTLSConfig()), NewUNIXSocket(SkipVerifyTLSConfig()), "unixs", t)
	testSocket(NewHTTPSocket(DefalutTLSConfig()), NewHTTPSocket(SkipVerifyTLSConfig()), "https", t)
	testSocket(NewWSSocket(DefalutTLSConfig()), NewWSSocket(SkipVerifyTLSConfig()), "wss", t)
}

func testSocket(serverSock Socket, clientSock Socket, scheme string, t *testing.T) {
	var addr = ":9999"
	if serverSock.Scheme() != scheme {
		t.Error(serverSock.Scheme())
	}
	if _, err := clientSock.Dial("-1"); err == nil {
		t.Error("should be missing port in address / no such file or directory")
	}
	if _, err := clientSock.Dial(addr); err == nil {
		t.Error("should be refused")
	}

	l, err := serverSock.Listen(addr)
	if err != nil {
		t.Error(err)
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			conn, err := l.Accept()
			if err != nil {
				return
			}
			wg.Add(1)
			go func(conn Conn) {
				defer wg.Done()
				messages := conn.Messages()
				for {
					msg, err := messages.ReadMessage()
					if err != nil {
						break
					}
					messages.WriteMessage(msg)
				}
				messages.Close()
			}(conn)
		}
	}()
	conn, err := clientSock.Dial(addr)
	if err != nil {
		t.Error(err)
	}
	conn.Connection()
	messages := conn.Messages()
	str := "Hello World"
	str = strings.Repeat(str, 50)
	messages.WriteMessage([]byte(str))
	msg, err := messages.ReadMessage()
	if err != nil {
		t.Error(err)
	}
	if string(msg) != str {
		t.Errorf("error %s != %s", string(msg), str)
	}
	messages.Close()
	l.Addr()
	l.Close()
	wg.Wait()
}

func TestSocketTLS(t *testing.T) {
	config := SkipVerifyTLSConfig()
	config.InsecureSkipVerify = false
	testSocketTLS(NewTCPSocket(DefalutTLSConfig()), NewTCPSocket(config), t)
	testSocketTLS(NewUNIXSocket(DefalutTLSConfig()), NewUNIXSocket(config), t)
	testSocketTLS(NewHTTPSocket(DefalutTLSConfig()), NewHTTPSocket(config), t)
	testSocketTLS(NewWSSocket(DefalutTLSConfig()), NewWSSocket(config), t)
}

func testSocketTLS(serverSock Socket, clientSock Socket, t *testing.T) {
	var addr = ":9999"
	l, err := serverSock.Listen(addr)
	if err != nil {
		t.Error(err)
	}
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			conn, err := l.Accept()
			if err != nil {
				return
			}
			wg.Add(1)
			go func(conn Conn) {
				defer wg.Done()
				messages := conn.Messages()
				for {
					msg, err := messages.ReadMessage()
					if err != nil {
						break
					}
					messages.WriteMessage(msg)
				}
				messages.Close()
			}(conn)
		}
	}()
	_, err = clientSock.Dial(addr)
	if err == nil {
		t.Error("should be x509: certificate is valid for hslam, not :9999")
	}
	time.Sleep(time.Millisecond)
	l.Close()
	wg.Wait()
}

func TestSocketServe(t *testing.T) {
	var handler = netpoll.NewHandler(func(conn net.Conn) (netpoll.Context, error) {
		return conn, nil
	}, func(context netpoll.Context) error {
		return nil
	})
	testSocketServe(NewTCPSocket(nil), nil, t)
	testSocketServe(NewUNIXSocket(nil), nil, t)
	testSocketServe(NewHTTPSocket(nil), nil, t)
	testSocketServe(NewWSSocket(nil), nil, t)
	testSocketServe(NewTCPSocket(nil), handler, t)
	testSocketServe(NewUNIXSocket(nil), handler, t)
	testSocketServe(NewHTTPSocket(nil), handler, t)
	testSocketServe(NewWSSocket(nil), handler, t)
}

func testSocketServe(serverSock Socket, handler netpoll.Handler, t *testing.T) {
	var addr = ":9999"
	l, err := serverSock.Listen(addr)
	if err != nil {
		t.Error(err)
	}
	if handler != nil {
		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			defer wg.Done()
			l.Serve(handler)
		}()
		time.Sleep(time.Millisecond * 10)
		l.Close()
		wg.Wait()
	} else {
		if err := l.Serve(handler); err != ErrHandler {
			t.Error(err)
		}
		l.Close()
	}
}

func TestSocketServeData(t *testing.T) {
	testSocketServeData(NewTCPSocket(nil), NewTCPSocket(nil), t)
	testSocketServeData(NewUNIXSocket(nil), NewUNIXSocket(nil), t)
	testSocketServeData(NewHTTPSocket(nil), NewHTTPSocket(nil), t)
	testSocketServeData(NewWSSocket(nil), NewWSSocket(nil), t)
	testSocketServeData(NewTCPSocket(DefalutTLSConfig()), NewTCPSocket(SkipVerifyTLSConfig()), t)
	testSocketServeData(NewHTTPSocket(DefalutTLSConfig()), NewHTTPSocket(SkipVerifyTLSConfig()), t)
	testSocketServeData(NewWSSocket(DefalutTLSConfig()), NewWSSocket(SkipVerifyTLSConfig()), t)
	testSocketServeData(NewUNIXSocket(DefalutTLSConfig()), NewUNIXSocket(SkipVerifyTLSConfig()), t)

}

func testSocketServeData(serverSock Socket, clientSock Socket, t *testing.T) {
	var addr = ":9999"
	l, err := serverSock.Listen(addr)
	if err != nil {
		t.Error(err)
	}
	if err := l.ServeData(nil, nil); err != ErrServe && err != ErrOpened {
		t.Error(err)
	}
	if err := l.ServeData(func(conn net.Conn) error {
		return nil
	}, nil); err != ErrServe && err != ErrOpened {
		t.Error(err)
	}
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		l.ServeData(func(conn net.Conn) error {
			return nil
		}, func(req []byte) (res []byte) {
			res = req
			return
		})
	}()
	conn, err := clientSock.Dial(addr)
	if err != nil {
		t.Error(err)
	}
	str := "Hello World"
	str = strings.Repeat(str, 50)
	if n, err := conn.Write([]byte(str)); err != nil {
		t.Error(err)
	} else if n != len(str) {
		t.Error(n)
	}
	buf := make([]byte, len(str))
	if n, err := conn.Read(buf); err != nil {
		t.Error(err)
	} else if n != len(str) {
		t.Error(n)
	}
	if string(buf) != str {
		t.Errorf("error %s != %s", string(buf), str)
	}
	conn.Close()
	l.Close()
	wg.Wait()
}

func TestSocketServeDataOpened(t *testing.T) {
	testSocketServeDataOpened(NewTCPSocket(nil), NewTCPSocket(nil), t)
	testSocketServeDataOpened(NewUNIXSocket(nil), NewUNIXSocket(nil), t)
	testSocketServeDataOpened(NewHTTPSocket(nil), NewHTTPSocket(nil), t)
	testSocketServeDataOpened(NewWSSocket(nil), NewWSSocket(nil), t)

}

func testSocketServeDataOpened(serverSock Socket, clientSock Socket, t *testing.T) {
	var addr = ":9999"
	l, err := serverSock.Listen(addr)
	if err != nil {
		t.Error(err)
	}
	if err := l.ServeData(nil, nil); err != ErrServe && err != ErrOpened {
		t.Error(err)
	}
	if err := l.ServeData(func(conn net.Conn) error {
		return nil
	}, nil); err != ErrServe && err != ErrOpened {
		t.Error(err)
	}
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		l.ServeData(func(conn net.Conn) error {
			return errors.New("fake error")
		}, func(req []byte) (res []byte) {
			res = req
			return
		})
	}()
	conn, err := clientSock.Dial(addr)
	if err != nil {
		t.Error(err)
	}
	time.Sleep(time.Millisecond * 10)
	conn.Close()
	l.Close()
	wg.Wait()
}

func TestSocketServeDataServe(t *testing.T) {
	testSocketServeDataServe(NewTCPSocket(nil), NewTCPSocket(nil), t)
	testSocketServeDataServe(NewUNIXSocket(nil), NewUNIXSocket(nil), t)
	testSocketServeDataServe(NewHTTPSocket(nil), NewHTTPSocket(nil), t)
	testSocketServeDataServe(NewWSSocket(nil), NewWSSocket(nil), t)
}

func testSocketServeDataServe(serverSock Socket, clientSock Socket, t *testing.T) {
	var addr = ":9999"
	l, err := serverSock.Listen(addr)
	if err != nil {
		t.Error(err)
	}
	if err := l.ServeData(nil, nil); err != ErrServe && err != ErrOpened {
		t.Error(err)
	}
	if err := l.ServeData(func(conn net.Conn) error {
		return nil
	}, nil); err != ErrServe && err != ErrOpened {
		t.Error(err)
	}
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		l.ServeData(func(conn net.Conn) error {
			return nil
		}, func(req []byte) (res []byte) {
			return
		})
	}()
	conn, err := clientSock.Dial(addr)
	if err != nil {
		t.Error(err)
	}
	str := "Hello World"
	str = strings.Repeat(str, 50)
	if n, err := conn.Write([]byte(str)); err != nil {
		t.Error(err)
	} else if n != len(str) {
		t.Error(n)
	}
	time.Sleep(time.Millisecond * 10)
	conn.Close()
	l.Close()
	wg.Wait()
}

func TestSocketServeDataTLS(t *testing.T) {
	config := SkipVerifyTLSConfig()
	config.InsecureSkipVerify = false
	testSocketServeDataTLS(NewTCPSocket(DefalutTLSConfig()), NewTCPSocket(config), t)
	testSocketServeDataTLS(NewHTTPSocket(DefalutTLSConfig()), NewHTTPSocket(config), t)
	testSocketServeDataTLS(NewWSSocket(DefalutTLSConfig()), NewWSSocket(config), t)
	testSocketServeDataTLS(NewUNIXSocket(DefalutTLSConfig()), NewUNIXSocket(config), t)
}

func testSocketServeDataTLS(serverSock Socket, clientSock Socket, t *testing.T) {
	var addr = ":9999"
	l, err := serverSock.Listen(addr)
	if err != nil {
		t.Error(err)
	}
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		l.ServeData(func(conn net.Conn) error {
			return nil
		}, func(req []byte) (res []byte) {
			res = req
			return
		})
	}()
	_, err = clientSock.Dial(addr)
	if err == nil {
		t.Error("should be x509: certificate is valid for hslam, not :9999")
	}
	time.Sleep(time.Millisecond)
	l.Close()
	wg.Wait()
}

func TestSocketServeConn(t *testing.T) {
	testSocketServeConn(NewTCPSocket(nil), NewTCPSocket(nil), t)
	testSocketServeConn(NewUNIXSocket(nil), NewUNIXSocket(nil), t)
	testSocketServeConn(NewHTTPSocket(nil), NewHTTPSocket(nil), t)
	testSocketServeConn(NewWSSocket(nil), NewWSSocket(nil), t)
	testSocketServeConn(NewTCPSocket(DefalutTLSConfig()), NewTCPSocket(SkipVerifyTLSConfig()), t)
	testSocketServeConn(NewHTTPSocket(DefalutTLSConfig()), NewHTTPSocket(SkipVerifyTLSConfig()), t)
	testSocketServeConn(NewWSSocket(DefalutTLSConfig()), NewWSSocket(SkipVerifyTLSConfig()), t)
	testSocketServeConn(NewUNIXSocket(DefalutTLSConfig()), NewUNIXSocket(SkipVerifyTLSConfig()), t)

}

func testSocketServeConn(serverSock Socket, clientSock Socket, t *testing.T) {
	var addr = ":9999"
	l, err := serverSock.Listen(addr)
	if err != nil {
		t.Error(err)
	}
	if err := l.ServeConn(nil, nil); err != ErrServe && err != ErrOpened {
		t.Error(err)
	}
	if err := l.ServeConn(func(conn net.Conn) (Context, error) {
		return conn, nil
	}, nil); err != ErrServe && err != ErrOpened {
		t.Error(err)
	}
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		type context struct {
			Conn net.Conn
			buf  []byte
		}
		l.ServeConn(func(conn net.Conn) (Context, error) {
			ctx := &context{
				Conn: conn,
				buf:  make([]byte, 1024*64),
			}
			return ctx, nil
		}, func(ctx Context) error {
			c := ctx.(*context)
			n, err := c.Conn.Read(c.buf)
			if err != nil {
				return err
			}
			_, err = c.Conn.Write(c.buf[:n])
			return nil
		})
	}()
	conn, err := clientSock.Dial(addr)
	if err != nil {
		t.Error(err)
	}
	str := "Hello World"
	str = strings.Repeat(str, 50)
	if n, err := conn.Write([]byte(str)); err != nil {
		t.Error(err)
	} else if n != len(str) {
		t.Error(n)
	}
	buf := make([]byte, len(str))
	if n, err := conn.Read(buf); err != nil {
		t.Error(err)
	} else if n != len(str) {
		t.Error(n)
	}
	if string(buf) != str {
		t.Errorf("error %s != %s", string(buf), str)
	}
	conn.Close()
	l.Close()
	wg.Wait()
}

func TestSocketServeConnTLS(t *testing.T) {
	config := SkipVerifyTLSConfig()
	config.InsecureSkipVerify = false
	testSocketServeConnTLS(NewTCPSocket(DefalutTLSConfig()), NewTCPSocket(config), t)
	testSocketServeConnTLS(NewHTTPSocket(DefalutTLSConfig()), NewHTTPSocket(config), t)
	testSocketServeConnTLS(NewWSSocket(DefalutTLSConfig()), NewWSSocket(config), t)
	testSocketServeConnTLS(NewUNIXSocket(DefalutTLSConfig()), NewUNIXSocket(config), t)
}

func testSocketServeConnTLS(serverSock Socket, clientSock Socket, t *testing.T) {
	var addr = ":9999"
	l, err := serverSock.Listen(addr)
	if err != nil {
		t.Error(err)
	}
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		type context struct {
			Conn net.Conn
			buf  []byte
		}
		l.ServeConn(func(conn net.Conn) (Context, error) {
			ctx := &context{
				Conn: conn,
				buf:  make([]byte, 1024*64),
			}
			return ctx, nil
		}, func(ctx Context) error {
			c := ctx.(*context)
			n, err := c.Conn.Read(c.buf)
			if err != nil {
				return err
			}
			_, err = c.Conn.Write(c.buf[:n])
			return nil
		})
	}()
	_, err = clientSock.Dial(addr)
	if err == nil {
		t.Error("should be x509: certificate is valid for hslam, not :9999")
	}
	time.Sleep(time.Millisecond)
	l.Close()
	wg.Wait()
}

func TestSocketServeMessages(t *testing.T) {
	testSocketServeMessages(NewTCPSocket(nil), NewTCPSocket(nil), t)
	testSocketServeMessages(NewUNIXSocket(nil), NewUNIXSocket(nil), t)
	testSocketServeMessages(NewHTTPSocket(nil), NewHTTPSocket(nil), t)
	testSocketServeMessages(NewWSSocket(nil), NewWSSocket(nil), t)
	testSocketServeMessages(NewTCPSocket(DefalutTLSConfig()), NewTCPSocket(SkipVerifyTLSConfig()), t)
	testSocketServeMessages(NewHTTPSocket(DefalutTLSConfig()), NewHTTPSocket(SkipVerifyTLSConfig()), t)
	testSocketServeMessages(NewWSSocket(DefalutTLSConfig()), NewWSSocket(SkipVerifyTLSConfig()), t)
	testSocketServeMessages(NewUNIXSocket(DefalutTLSConfig()), NewUNIXSocket(SkipVerifyTLSConfig()), t)
}

func testSocketServeMessages(serverSock Socket, clientSock Socket, t *testing.T) {
	var addr = ":9999"
	l, err := serverSock.Listen(addr)
	if err != nil {
		t.Error(err)
	}
	if err := l.ServeMessages(nil, nil); err != ErrServe && err != ErrOpened {
		t.Error(err)
	}
	if err := l.ServeMessages(func(messages Messages) (Context, error) {
		return messages, nil
	}, nil); err != ErrServe && err != ErrOpened {
		t.Error(err)
	}
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		l.ServeMessages(func(messages Messages) (Context, error) {
			return messages, nil
		}, func(context Context) error {
			messages := context.(Messages)
			msg, err := messages.ReadMessage()
			if err != nil {
				return err
			}
			return messages.WriteMessage(msg)
		})
	}()
	conn, err := clientSock.Dial(addr)
	if err != nil {
		t.Error(err)
	}
	messages := conn.Messages()
	str := "Hello World"
	str = strings.Repeat(str, 50)
	messages.WriteMessage([]byte(str))
	msg, err := messages.ReadMessage()
	if err != nil {
		t.Error(err)
	}
	if string(msg) != str {
		t.Errorf("error %s != %s", string(msg), str)
	}
	messages.Close()
	l.Close()
	wg.Wait()
}

func TestSocketServeMessagesTLS(t *testing.T) {
	config := SkipVerifyTLSConfig()
	config.InsecureSkipVerify = false
	testSocketServeMessagesTLS(NewTCPSocket(DefalutTLSConfig()), NewTCPSocket(config), t)
	testSocketServeMessagesTLS(NewHTTPSocket(DefalutTLSConfig()), NewHTTPSocket(config), t)
	testSocketServeMessagesTLS(NewWSSocket(DefalutTLSConfig()), NewWSSocket(config), t)
	testSocketServeMessagesTLS(NewUNIXSocket(DefalutTLSConfig()), NewUNIXSocket(config), t)
}

func testSocketServeMessagesTLS(serverSock Socket, clientSock Socket, t *testing.T) {
	var addr = ":9999"
	l, err := serverSock.Listen(addr)
	if err != nil {
		t.Error(err)
	}
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		l.ServeMessages(func(messages Messages) (Context, error) {
			return messages, nil
		}, func(context Context) error {
			messages := context.(Messages)
			msg, err := messages.ReadMessage()
			if err != nil {
				return err
			}
			return messages.WriteMessage(msg)
		})
	}()
	_, err = clientSock.Dial(addr)
	if err == nil {
		t.Error("should be x509: certificate is valid for hslam, not :9999")
	}
	time.Sleep(time.Millisecond)
	l.Close()
	wg.Wait()
}

func TestNewSocket(t *testing.T) {
	if sock, err := NewSocket("tcp", nil); err != nil {
		t.Error(err)
	} else if _, ok := sock.(*TCP); !ok {
		t.Error(sock)
	}
	if sock, err := NewSocket("unix", nil); err != nil {
		t.Error(err)
	} else if _, ok := sock.(*UNIX); !ok {
		t.Error(sock)
	}
	if sock, err := NewSocket("http", nil); err != nil {
		t.Error(err)
	} else if _, ok := sock.(*HTTP); !ok {
		t.Error(sock)
	}
	if sock, err := NewSocket("ws", nil); err != nil {
		t.Error(err)
	} else if _, ok := sock.(*WS); !ok {
		t.Error(sock)
	}
	if _, err := NewSocket("", nil); err != ErrNetwork {
		t.Error(err)
	}
}

func TestAddress(t *testing.T) {
	sock := NewTCPSocket(nil)
	if addr, err := Address(sock, "tcp://localhost:9999"); err != nil {
		t.Error(err)
	} else if addr != "localhost:9999" {
		t.Error(addr)
	}

	if _, err := Address(sock, "http://localhost:9999"); err == nil {
		t.Error("here")
	}
}

func TestURL(t *testing.T) {
	sock := NewTCPSocket(nil)
	url := URL(sock, "localhost:9999")
	if url != "tcp://localhost:9999" {
		t.Error(url)
	}
}
