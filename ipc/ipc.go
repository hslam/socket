package ipc

import (
	"github.com/hslam/transport"
	"net"
	"os"
)

type IPC struct {
}

// NewTransport returns a new IPC transport.
func NewTransport() transport.Transport {
	return &IPC{}
}
func (t *IPC) Dial(address string) (transport.Conn, error) {
	var addr *net.UnixAddr
	var err error
	if addr, err = net.ResolveUnixAddr("unix", address); err != nil {
		return nil, err
	}
	conn, err := net.DialUnix("unix", nil, addr)
	if err != nil {
		return nil, err
	}
	return conn, err
}

func (t *IPC) Listen(address string) (transport.Listener, error) {
	os.Remove(address)
	var addr *net.UnixAddr
	var err error
	if addr, err = net.ResolveUnixAddr("unix", address); err != nil {
		return nil, err
	}
	lis, err := net.ListenUnix("unix", addr)
	if err != nil {
		return nil, err
	}

	return &IPCListener{lis}, err
}

type IPCListener struct {
	l *net.UnixListener
}

func (l *IPCListener) Accept() (transport.Conn, error) {
	return l.l.Accept()
}
