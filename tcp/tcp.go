// Copyright (c) 2020 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package tcp

import (
	"crypto/tls"
	"github.com/hslam/socket"
)

// NewSocket returns a new TCP socket.
func NewSocket() socket.Socket {
	return &socket.TCP{}
}

func NewTLSSocket(config *tls.Config) socket.Socket {
	return &socket.TCP{Config: config}
}
