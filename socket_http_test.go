package socket

import (
	"net"
	"sync"
	"testing"
	"time"
)

func TestResponse(t *testing.T) {
	var network = "tcp"
	var addr = ":9999"
	l, _ := net.Listen(network, addr)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			conn, err := l.Accept()
			if err != nil {
				break
			}
			wg.Add(1)
			go func(conn net.Conn) {
				defer wg.Done()
				buf := make([]byte, 1024*64)
				n, _ := conn.Read(buf)
				if n == 0 {
					t.Error("n == 0")
				}
			}(conn)
		}
	}()
	conn, _ := net.Dial(network, addr)
	res := &response{conn: conn}
	res.Header()
	res.WriteHeader(405)
	res.Write([]byte("405 must CONNECT\n"))
	time.Sleep(time.Millisecond * 10)
	conn.Close()
	l.Close()
	wg.Wait()
}

func TestHTTPSocketDial(t *testing.T) {
	var (
		serverSock = NewWSSocket(nil)
		clientSock = NewHTTPSocket(nil)
	)
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
		t.Error("should be dial-http tcp :9999: unexpected HTTP response: 405 Method Not Allowed")
	}
	time.Sleep(time.Millisecond * 10)
	l.Close()
	wg.Wait()
}

func TestHTTPSocketAccept(t *testing.T) {
	var (
		serverSock = NewHTTPSocket(nil)
		clientSock = NewWSSocket(nil)
	)
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
		t.Error("should be dial-http tcp :9999: unexpected HTTP response: 405 Method Not Allowed")
	}
	time.Sleep(time.Millisecond * 10)
	l.Close()
	wg.Wait()
}

func TestHTTPSocketAcceptTLS(t *testing.T) {
	var (
		serverSock = NewHTTPSocket(DefalutTLSConfig())
		clientSock = NewWSSocket(SkipVerifyTLSConfig())
	)
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
		t.Error("should be dial-http tcp :9999: unexpected HTTP response: 405 Method Not Allowed")
	}
	time.Sleep(time.Millisecond * 10)
	l.Close()
	wg.Wait()
}

func TestHTTPSocketServeData(t *testing.T) {
	var (
		serverSock = NewHTTPSocket(nil)
		clientSock = NewWSSocket(nil)
	)
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
	_, err = clientSock.Dial(addr)
	if err == nil {
		t.Error("should be dial-http tcp :9999: unexpected HTTP response: 405 Method Not Allowed")
	}
	time.Sleep(time.Millisecond * 10)
	l.Close()
	wg.Wait()
}

func TestHTTPSocketServeConn(t *testing.T) {
	var (
		serverSock = NewHTTPSocket(nil)
		clientSock = NewWSSocket(nil)
	)
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
		t.Error("should be dial-http tcp :9999: unexpected HTTP response: 405 Method Not Allowed")
	}
	time.Sleep(time.Millisecond * 10)
	l.Close()
	wg.Wait()
}

func TestHTTPSocketServeMessages(t *testing.T) {
	var (
		serverSock = NewHTTPSocket(nil)
		clientSock = NewWSSocket(nil)
	)
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
		t.Error("should be dial-http tcp :9999: unexpected HTTP response: 405 Method Not Allowed")
	}
	time.Sleep(time.Millisecond * 10)
	l.Close()
	wg.Wait()
}
