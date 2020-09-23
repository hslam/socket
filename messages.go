// Copyright (c) 2020 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package socket

import (
	"github.com/hslam/writer"
	"io"
	"sync"
	"sync/atomic"
)

const bufferSize = 65528

var (
	buffers = sync.Map{}
	assign  int32
)

func assignPool(size int) *sync.Pool {
	for {
		if p, ok := buffers.Load(size); ok {
			return p.(*sync.Pool)
		}
		if atomic.CompareAndSwapInt32(&assign, 0, 1) {
			var pool = &sync.Pool{New: func() interface{} {
				return make([]byte, size)
			}}
			buffers.Store(size, pool)
			atomic.StoreInt32(&assign, 0)
			return pool
		}
	}
}

type Messages interface {
	SetConcurrency(concurrency func() int)
	ReadMessage() ([]byte, error)
	WriteMessage([]byte) error
	Close() error
}

type messages struct {
	shared          bool
	reading         sync.Mutex
	writing         sync.Mutex
	reader          io.Reader
	writer          io.Writer
	closer          io.Closer
	readBufferSize  int
	readBuffer      []byte
	writeBufferSize int
	writeBuffer     []byte
	buffer          []byte
	readPool        *sync.Pool
	writePool       *sync.Pool
	closed          int32
}

func NewMessages(rwc io.ReadWriteCloser, shared bool, writeBufferSize int, readBufferSize int) Messages {
	if writeBufferSize < 1 {
		writeBufferSize = bufferSize
	}
	if readBufferSize < 1 {
		readBufferSize = bufferSize
	}
	writeBufferSize += 8
	readBufferSize += 8
	var readBuffer []byte
	var writeBuffer []byte
	var readPool *sync.Pool
	var writePool *sync.Pool
	if shared {
		readPool = assignPool(readBufferSize)
		writePool = assignPool(writeBufferSize)
	} else {
		readBuffer = make([]byte, readBufferSize)
		writeBuffer = make([]byte, writeBufferSize)
	}
	return &messages{
		shared:          shared,
		reader:          rwc,
		writer:          rwc,
		closer:          rwc,
		writeBufferSize: writeBufferSize,
		readBufferSize:  readBufferSize,
		readBuffer:      readBuffer,
		writeBuffer:     writeBuffer,
		readPool:        readPool,
		writePool:       writePool,
	}
}

func (m *messages) SetConcurrency(concurrency func() int) {
	if concurrency == nil {
		return
	}
	m.writing.Lock()
	defer m.writing.Unlock()
	m.writer = writer.NewWriter(m.writer, concurrency, 65536, false)
}

func (m *messages) ReadMessage() (p []byte, err error) {
	m.reading.Lock()
	defer m.reading.Unlock()
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
			m.buffer = m.buffer[i:]
			return
		}
		var readBuffer []byte
		if m.shared {
			readBuffer = m.readPool.Get().([]byte)
			readBuffer = readBuffer[:cap(readBuffer)]
		} else {
			readBuffer = m.readBuffer
		}
		n, err := m.reader.Read(readBuffer)
		if err != nil {
			if m.shared {
				m.readPool.Put(readBuffer)
			}
			return nil, err
		} else if n > 0 {
			m.buffer = append(m.buffer, readBuffer[:n]...)
			if m.shared {
				m.readPool.Put(readBuffer)
			}
		}
	}
}

func (m *messages) WriteMessage(b []byte) error {
	m.writing.Lock()
	defer m.writing.Unlock()
	var length = uint64(len(b))
	var size = 8 + length
	var writeBuffer []byte
	if m.shared {
		writeBuffer = m.writePool.Get().([]byte)
		writeBuffer = writeBuffer[:cap(writeBuffer)]
	} else {
		writeBuffer = m.writeBuffer
	}
	if uint64(cap(writeBuffer)) >= size {
		writeBuffer = writeBuffer[:size]
	} else {
		writeBuffer = make([]byte, size)
	}
	var t = length
	var buf = writeBuffer[0:8]
	buf[0] = uint8(t)
	buf[1] = uint8(t >> 8)
	buf[2] = uint8(t >> 16)
	buf[3] = uint8(t >> 24)
	buf[4] = uint8(t >> 32)
	buf[5] = uint8(t >> 40)
	buf[6] = uint8(t >> 48)
	buf[7] = uint8(t >> 56)
	copy(writeBuffer[8:], b)
	_, err := m.writer.Write(writeBuffer[:size])
	if m.shared {
		m.writePool.Put(writeBuffer)
	} else {
		m.writeBuffer = writeBuffer
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
