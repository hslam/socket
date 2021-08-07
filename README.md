# socket
[![PkgGoDev](https://pkg.go.dev/badge/github.com/hslam/socket)](https://pkg.go.dev/github.com/hslam/socket)
[![Build Status](https://github.com/hslam/socket/workflows/build/badge.svg)](https://github.com/hslam/socket/actions)
[![codecov](https://codecov.io/gh/hslam/socket/branch/master/graph/badge.svg)](https://codecov.io/gh/hslam/socket)
[![Go Report Card](https://goreportcard.com/badge/github.com/hslam/socket)](https://goreportcard.com/report/github.com/hslam/socket)
[![LICENSE](https://img.shields.io/github/license/hslam/socket.svg?style=flat-square)](https://github.com/hslam/socket/blob/master/LICENSE)

Package socket implements a network socket that supports TCP, UNIX, HTTP, WS and INPROC.

## Feature
* TCP/UNIX/HTTP/WS/INPROC
* [Epoll/Kqueue](https://github.com/hslam/netpoll "netpoll")
* TLS

## [Benchmark](https://github.com/hslam/socket-benchmark "socket-benchmark")

##### Socket QPS

<img src="https://raw.githubusercontent.com/hslam/socket-benchmark/master/socket-qps.png" width = "480" height = "360" alt="socket" align=center>


## Get started

### Install
```
go get github.com/hslam/socket
```
### Import
```
import "github.com/hslam/socket"
```
### Usage
#### Example

server.go
```go
package main

import (
	"flag"
	"github.com/hslam/socket"
)

var network string
var addr string

func init() {
	flag.StringVar(&network, "network", "tcp", "-network=tcp, unix, http or ws")
	flag.StringVar(&addr, "addr", ":8080", "-addr=:8080")
	flag.Parse()
}

func main() {
	sock, err := socket.NewSocket(network, nil)
	if err != nil {
		panic(err)
	}
	l, err := sock.Listen(addr)
	if err != nil {
		panic(err)
	}
	for {
		conn, err := l.Accept()
		if err != nil {
			panic(err)
		}
		go func(conn socket.Conn) {
			messages := conn.Messages()
			buf := make([]byte, 65536)
			for {
				msg, err := messages.ReadMessage(buf)
				if err != nil {
					break
				}
				messages.WriteMessage(msg)
			}
			messages.Close()
		}(conn)
	}
}
```

server_poll.go
```go
package main

import (
	"flag"
	"github.com/hslam/socket"
	"sync"
)

var network string
var addr string

func init() {
	flag.StringVar(&network, "network", "tcp", "-network=tcp, unix, http or ws")
	flag.StringVar(&addr, "addr", ":8080", "-addr=:8080")
	flag.Parse()
}

func main() {
	sock, err := socket.NewSocket(network, nil)
	if err != nil {
		panic(err)
	}
	l, err := sock.Listen(addr)
	if err != nil {
		panic(err)
	}
	bufferPool := &sync.Pool{New: func() interface{} { return make([]byte, 65536) }}
	l.ServeMessages(func(messages socket.Messages) (socket.Context, error) {
		return messages, nil
	}, func(context socket.Context) error {
		messages := context.(socket.Messages)
		buf := bufferPool.Get().([]byte)
		defer bufferPool.Put(buf)
		msg, err := messages.ReadMessage(buf)
		if err != nil {
			return err
		}
		return messages.WriteMessage(msg)
	})
}
```

client.go
```go
package main

import (
	"flag"
	"fmt"
	"github.com/hslam/socket"
)

var network string
var addr string

func init() {
	flag.StringVar(&network, "network", "tcp", "-network=tcp, unix, http or ws")
	flag.StringVar(&addr, "addr", ":8080", "-addr=:8080")
	flag.Parse()
}

func main() {
	sock, err := socket.NewSocket(network, nil)
	if err != nil {
		panic(err)
	}
	conn, err := sock.Dial(addr)
	if err != nil {
		panic(err)
	}
	messages := conn.Messages()
	buf := make([]byte, 65536)
	for i := 0; i < 1; i++ {
		err := messages.WriteMessage([]byte("Hello World"))
		if err != nil {
			panic(err)
		}
		msg, err := messages.ReadMessage(buf)
		if err != nil {
			panic(err)
		}
		fmt.Println(string(msg))
	}
}
```

**Output**
```
Hello World
```

### License
This package is licensed under a MIT license (Copyright (c) 2020 Meng Huang)

### Author
socket was written by Meng Huang.

