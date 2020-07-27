package ws

import (
	"github.com/hslam/socket"
)

type Messages struct {
	*WSConn
}

func (m *Messages) SetBatch(batch socket.Batch) {
	m.Conn.SetBatch(batch)
}

func (m *Messages) ReadMessage() (p []byte, err error) {
	return m.Conn.ReadBinaryMessage()
}

func (m *Messages) WriteMessage(b []byte) (err error) {
	return m.Conn.WriteBinaryMessage(b)
}

func (s *Messages) Close() error {
	return s.Conn.Close()
}
