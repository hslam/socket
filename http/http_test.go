package http

import (
	"github.com/hslam/socket"
	"testing"
)

func TestHTTP(t *testing.T) {
	addr := ":9999"
	started := make(chan bool, 1)
	go server(addr, started, t)
	<-started
	client(addr, t)
}

func client(addr string, t *testing.T) {
	s := NewSocket()
	conn, err := s.Dial(addr)
	if err != nil {
		t.Error(err)
	}
	messages := conn.Messages()
	str := "abc"
	messages.WriteMessage([]byte(str))
	msg, err := messages.ReadMessage()
	if err != nil {
		t.Error(err)
	}
	if string(msg) != str {
		t.Errorf("error %s != %s", string(msg), str)
	}
	messages.Close()
}

func server(addr string, started chan bool, t *testing.T) {
	s := NewSocket()
	l, err := s.Listen(addr)
	if err != nil {
		t.Error(err)
	}
	started <- true
	for {
		conn, err := l.Accept()
		if err != nil {
			t.Error(err)
		}
		go func(conn socket.Conn) {
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
}
