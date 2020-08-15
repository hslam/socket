// Copyright (c) 2020 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package socket

import (
	"github.com/hslam/autowriter"
	"io"
)

var BufferSize = 65536

type Messages interface {
	SetBatch(batch Batch)
	ReadMessage() ([]byte, error)
	WriteMessage([]byte) error
	Close() error
}

type Batch interface {
	Concurrency() int
}

type messages struct {
	Reader io.Reader
	Writer io.Writer
	Closer io.Closer
	Write  []byte
	Read   []byte
	buffer []byte
}

func NewMessages(rwc io.ReadWriteCloser, writeBufferSize int, readBufferSize int) Messages {
	if writeBufferSize < 1 {
		writeBufferSize = BufferSize
	}
	if readBufferSize < 1 {
		readBufferSize = BufferSize
	}
	return &messages{
		Reader: rwc,
		Writer: rwc,
		Closer: rwc,
		Write:  make([]byte, writeBufferSize),
		Read:   make([]byte, readBufferSize),
	}
}

func (m *messages) SetBatch(batch Batch) {
	m.Writer = autowriter.NewAutoWriter(m.Writer, false, 65536, 4, batch)
}

func (m *messages) ReadMessage() (p []byte, err error) {
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

func (m *messages) WriteMessage(b []byte) error {
	var length = uint64(len(b))
	var size = 8 + length
	if uint64(cap(m.Write)) >= size {
		m.Write = m.Write[:size]
	} else {
		m.Write = make([]byte, size)
	}
	var t = length
	var buf = m.Write[0:8]
	buf[0] = uint8(t)
	buf[1] = uint8(t >> 8)
	buf[2] = uint8(t >> 16)
	buf[3] = uint8(t >> 24)
	buf[4] = uint8(t >> 32)
	buf[5] = uint8(t >> 40)
	buf[6] = uint8(t >> 48)
	buf[7] = uint8(t >> 56)
	copy(m.Write[8:], b)
	_, err := m.Writer.Write(m.Write[:size])
	return err
}

func (m *messages) Close() error {
	return m.Closer.Close()
}
