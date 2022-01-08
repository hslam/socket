// Copyright (c) 2020 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package socket

import (
	"github.com/hslam/buffer"
	"github.com/hslam/writer"
	"io"
	"strings"
	"sync"
	"sync/atomic"
)

const bufferSize = 65536

// BufferedOutput sets the buffered writer with the buffer size.
type BufferedOutput interface {
	SetBufferedOutput(bufferSize int)
}

// BufferedInput sets the read buffer size.
type BufferedInput interface {
	SetBufferedInput(bufferSize int)
}

// Messages interface is used to read and write message.
type Messages interface {
	// ReadMessage reads single message frame from the Messages.
	ReadMessage(buf []byte) ([]byte, error)
	// WriteMessage writes data as a message frame to the Messages.
	WriteMessage([]byte) error
	// Close closes the Messages.
	Close() error
}

type messages struct {
	shared          bool
	scheduling      bool
	reading         sync.Mutex
	writing         sync.Mutex
	rwc             io.ReadWriteCloser
	reader          io.Reader
	writer          io.Writer
	closer          io.Closer
	readBufferSize  int
	readBuffer      []byte
	writeBufferSize int
	writeBuffer     []byte
	buffer          []byte
	readPool        *buffer.Pool
	writePool       *buffer.Pool
	closed          int32
}

// NewMessages returns a new messages.
func NewMessages(rwc io.ReadWriteCloser, shared bool) Messages {
	var readBuffer []byte
	var writeBuffer []byte
	var readPool *buffer.Pool
	var writePool *buffer.Pool
	if shared {
		readPool = buffer.AssignPool(bufferSize)
		writePool = buffer.AssignPool(bufferSize)
	} else {
		readBuffer = make([]byte, bufferSize)
		writeBuffer = make([]byte, bufferSize)
	}
	return &messages{
		shared:          shared,
		rwc:             rwc,
		reader:          rwc,
		writer:          rwc,
		closer:          rwc,
		writeBufferSize: bufferSize,
		readBufferSize:  bufferSize,
		readBuffer:      readBuffer,
		writeBuffer:     writeBuffer,
		readPool:        readPool,
		writePool:       writePool,
	}
}

// SetBufferedOutput sets the buffered writer with the buffer size.
func (m *messages) SetBufferedOutput(writeBufferSize int) {
	m.writing.Lock()
	if w, ok := m.writer.(*writer.Writer); ok {
		w.Close()
	}
	if writeBufferSize > 0 {
		m.writer = writer.NewWriter(m.rwc, writeBufferSize)
	} else {
		m.writer = m.rwc
		writeBufferSize = bufferSize
	}
	m.writeBufferSize = writeBufferSize
	if m.shared {
		m.writePool = buffer.AssignPool(writeBufferSize)
	} else {
		m.writeBuffer = make([]byte, writeBufferSize)
	}
	m.writing.Unlock()
}

// SetBufferedInput sets the read buffer size.
func (m *messages) SetBufferedInput(readBufferSize int) {
	if readBufferSize < 1 {
		readBufferSize = bufferSize
	}
	m.reading.Lock()
	m.readBufferSize = readBufferSize
	if m.shared {
		m.readPool = buffer.AssignPool(readBufferSize)
	} else {
		m.readBuffer = make([]byte, readBufferSize)
	}
	m.reading.Unlock()
}

func (m *messages) ReadMessage(buf []byte) (p []byte, err error) {
	m.reading.Lock()
	for {
		length := uint64(len(m.buffer))
		var i uint64 = 0
		if i < length {
			var s uint
			var msgLength uint64
			var t uint64
			var b byte
			if length < i+1 {
				goto read
			}
			b = m.buffer[i]
			if b > 127 {
				for b >= 0x80 {
					t |= uint64(b&0x7f) << s
					s += 7
					i++
					if length < i+1 {
						goto read
					}
					b = m.buffer[i]
				}
			}
			if i > 9 || i == 9 && b > 1 {
				panic("varint overflows a 64-bit integer")
			}
			t |= uint64(b) << s
			i++
			msgLength = t
			if length < i+msgLength {
				goto read
			}
			if uint64(cap(buf)) >= msgLength {
				p = buf[:msgLength]
			} else {
				p = make([]byte, msgLength)
			}
			copy(p, m.buffer[i:i+msgLength])
			i += msgLength
			n := copy(m.buffer, m.buffer[i:])
			m.buffer = m.buffer[:n]
			m.reading.Unlock()
			return
		}
	read:
		var readBuffer []byte
		if m.shared {
			readBuffer = m.readPool.GetBuffer(m.readBufferSize)
			readBuffer = readBuffer[:cap(readBuffer)]
		} else {
			readBuffer = m.readBuffer
		}
		n, err := m.reader.Read(readBuffer)
		if err != nil {
			m.reading.Unlock()
			errMsg := err.Error()
			if strings.Contains(errMsg, "use of closed network connection") || strings.Contains(errMsg, "connection reset by peer") {
				err = io.EOF
			}
			if m.shared {
				m.readPool.PutBuffer(readBuffer)
			}
			return nil, err
		} else if n > 0 {
			length := len(m.buffer)
			size := length + n
			if cap(m.buffer) >= size {
				m.buffer = m.buffer[:size]
				copy(m.buffer[length:], readBuffer[:n])
			} else {
				m.buffer = append(m.buffer, readBuffer[:n]...)
			}
			if m.shared {
				m.readPool.PutBuffer(readBuffer)
			}
		}
	}
}

func (m *messages) WriteMessage(b []byte) error {
	m.writing.Lock()
	var length = uint64(len(b))
	var size = 10 + length
	var writeBuffer []byte
	var big bool
	if m.writeBufferSize < int(size) {
		big = true
		writeBuffer = buffer.GetBuffer(int(size))
	} else if m.shared {
		writeBuffer = m.writePool.GetBuffer(m.writeBufferSize)
		writeBuffer = writeBuffer[:cap(writeBuffer)]
	} else {
		writeBuffer = m.writeBuffer
	}
	writeBuffer = writeBuffer[:size]
	var t = length
	var buf = writeBuffer[0:10]
	i := 0
	if t > 127 {
		for t >= 0x80 {
			buf[i] = byte(t) | 0x80
			t >>= 7
			i++
		}
	}
	buf[i] = byte(t)
	i++
	n := copy(writeBuffer[i:], b)
	i += n
	_, err := m.writer.Write(writeBuffer[:i])
	m.writing.Unlock()
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "use of closed network connection") || strings.Contains(errMsg, "connection reset by peer") {
			err = io.EOF
		}
	}
	if big {
		buffer.PutBuffer(writeBuffer)
	} else if m.shared {
		m.writePool.PutBuffer(writeBuffer)
	}
	return err
}

func (m *messages) Close() error {
	if !atomic.CompareAndSwapInt32(&m.closed, 0, 1) {
		return nil
	}
	if w, ok := m.writer.(*writer.Writer); ok {
		w.Close()
	}
	return m.closer.Close()
}
