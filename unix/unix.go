// Copyright (c) 2020 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package unix

import (
	"crypto/tls"
	"github.com/hslam/socket"
)

// NewSocket returns a new UNIX socket.
func NewSocket() socket.Socket {
	return &socket.UNIX{}
}

func NewTLSSocket(config *tls.Config) socket.Socket {
	return &socket.UNIX{Config: config}
}
