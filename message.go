package socket

import (
	"io"
)

type Message interface {
	SetReader(io.Reader)
	GetReader() io.Reader
	SetWriter(io.Writer)
	GetWriter() io.Writer
	SetCloser(io.Closer)
	GetCloser() io.Closer
	ReadMessage() (p []byte, err error)
	WriteMessage(b []byte) (err error)
	Close() error
}

type message struct {
	Reader io.Reader
	Writer io.Writer
	Closer io.Closer
	Send   []byte
	Read   []byte
	buffer []byte
}

func NewMessage(r io.Reader, w io.Writer, c io.Closer, bufferSize int) Message {
	if bufferSize < 1 {
		bufferSize = 1024
	}
	return &message{
		Reader: r,
		Writer: w,
		Closer: c,
		Send:   make([]byte, bufferSize+8),
		Read:   make([]byte, bufferSize),
	}
}

func (m *message) SetReader(r io.Reader) {
	m.Reader = r
}

func (m *message) GetReader() io.Reader {
	return m.Reader
}

func (m *message) SetWriter(w io.Writer) {
	m.Writer = w
}

func (m *message) GetWriter() io.Writer {
	return m.Writer
}

func (m *message) SetCloser(c io.Closer) {
	m.Closer = c
}

func (m *message) GetCloser() io.Closer {
	return m.Closer
}

func (m *message) ReadMessage() (p []byte, err error) {
	for {
		length := uint64(len(m.buffer))
		var i uint64 = 0
		for i < length {
			if length < i+8 {
				break
			}
			var msgLength uint64
			buf := m.buffer[i : i+8]
			var t uint64
			t = uint64(buf[0])
			t |= uint64(buf[1]) << 8
			t |= uint64(buf[2]) << 16
			t |= uint64(buf[3]) << 24
			t |= uint64(buf[4]) << 32
			t |= uint64(buf[5]) << 40
			t |= uint64(buf[6]) << 48
			t |= uint64(buf[7]) << 56
			msgLength = t
			if length < i+8+msgLength {
				break
			}
			p = m.buffer[i+8 : i+8+msgLength]
			i += 8 + msgLength
			break
		}
		m.buffer = m.buffer[i:]
		if i > 0 {
			break
		}
		n, err := m.Reader.Read(m.Read)
		if err != nil {
			return p, err
		}
		if n > 0 {
			m.buffer = append(m.buffer, m.Read[:n]...)
		}
	}
	return
}

func (m *message) WriteMessage(b []byte) error {
	var length = uint64(len(b))
	var size = 8 + length
	if uint64(cap(m.Send)) >= size {
		m.Send = m.Send[:size]
	} else {
		m.Send = make([]byte, size)
	}
	var t = length
	var buf = m.Send[0:8]
	buf[0] = uint8(t)
	buf[1] = uint8(t >> 8)
	buf[2] = uint8(t >> 16)
	buf[3] = uint8(t >> 24)
	buf[4] = uint8(t >> 32)
	buf[5] = uint8(t >> 40)
	buf[6] = uint8(t >> 48)
	buf[7] = uint8(t >> 56)
	copy(m.Send[8:], b)
	_, err := m.Writer.Write(m.Send[:size])
	return err
}

func (m *message) Close() error {
	return m.Closer.Close()
}
