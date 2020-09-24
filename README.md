# socket
Package socket implements a network socket that supports TCP, UNIX, HTTP and WS.

## Feature
* TCP / UNIX / HTTP / WS
* [netpoll](https://github.com/hslam/netpoll "netpoll")(Epoll / Kqueue)

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

**server.go**
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
```

**server_poll.go**
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
	l.ServeMessages(func(messages socket.Messages) (socket.Context, error) {
		return messages, nil
	}, func(context socket.Context) error {
		messages := context.(socket.Messages)
		msg, err := messages.ReadMessage()
		if err != nil {
			return err
		}
		return messages.WriteMessage(msg)
	})
}
```

**client.go**
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
	for i := 0; i < 1; i++ {
		err := messages.WriteMessage([]byte("Hello World"))
		if err != nil {
			panic(err)
		}
		msg, err := messages.ReadMessage()
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


### Authors
socket was written by Meng Huang.

