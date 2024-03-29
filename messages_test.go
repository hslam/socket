// Copyright (c) 2020 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package socket

import (
	"io"
	"os"
	"strings"
	"testing"
)

func TestMessages(t *testing.T) {
	name := "tmpTestMessages"
	file, _ := os.Create(name)
	messages := NewMessages(file, true)
	str := strings.Repeat("Hello World", 50)
	err := messages.WriteMessage([]byte(str))
	if err != nil {
		t.Log(err)
	}
	file.Seek(0, os.SEEK_SET)
	messages.(BufferedOutput).SetBufferedOutput(0)
	messages.(BufferedOutput).SetBufferedOutput(64)
	messages.(BufferedOutput).SetBufferedOutput(64)
	messages.(BufferedInput).SetBufferedInput(0)
	messages.(BufferedInput).SetBufferedInput(2)
	if msg, err := messages.ReadMessage(make([]byte, 65536)); err != nil {
		t.Error(err)
	} else if string(msg) != str {
		t.Error(string(msg))
	}
	if _, err := messages.ReadMessage(nil); err != io.EOF {
		t.Error(err)
	}
	messages.Close()
	messages.Close()
	os.Remove(name)
}

func TestMessagesRetain(t *testing.T) {
	name := "tmpTestMessages"
	file, _ := os.Create(name)
	defer os.Remove(name)
	writeBufferSize := 7
	readBufferSize := 7
	var readBuffer []byte
	var writeBuffer []byte
	readBuffer = make([]byte, readBufferSize)
	writeBuffer = make([]byte, writeBufferSize)
	messages := &messages{
		shared:          false,
		reader:          file,
		writer:          file,
		closer:          file,
		writeBufferSize: writeBufferSize,
		readBufferSize:  readBufferSize,
		readBuffer:      readBuffer,
		writeBuffer:     writeBuffer,
	}
	str := strings.Repeat("Hello World", 50)
	err := messages.WriteMessage([]byte(str))
	if err != nil {
		t.Log(err)
	}
	file.Seek(0, os.SEEK_SET)
	if msg, err := messages.ReadMessage(nil); err != nil {
		t.Error(err)
	} else if string(msg) != str {
		t.Error(string(msg))
	}
	messages.Close()
}
