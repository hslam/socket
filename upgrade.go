package socket

import (
	"crypto/tls"
	"github.com/hslam/poll"
	"github.com/hslam/websocket"
	"net"
)

func Upgrade() func(c net.Conn) (net.Conn, poll.Messages, error) {
	return nil
}

func UpgradeMessages() func(c net.Conn) (net.Conn, poll.Messages, error) {
	return func(c net.Conn) (net.Conn, poll.Messages, error) {
		messages := NewMessages(c, 0, 0)
		return nil, messages, nil
	}
}

func UpgradeTLS(config *tls.Config) func(c net.Conn) (net.Conn, poll.Messages, error) {
	return func(c net.Conn) (net.Conn, poll.Messages, error) {
		tlsConn := tls.Server(c, config)
		if err := tlsConn.Handshake(); err != nil {
			tlsConn.Close()
			return nil, nil, err
		}
		return tlsConn, nil, nil
	}
}

func UpgradeTLSMessages(config *tls.Config) func(c net.Conn) (net.Conn, poll.Messages, error) {
	return func(c net.Conn) (net.Conn, poll.Messages, error) {
		tlsConn := tls.Server(c, config)
		if err := tlsConn.Handshake(); err != nil {
			tlsConn.Close()
			return nil, nil, err
		}
		messages := NewMessages(tlsConn, 0, 0)
		return nil, messages, nil
	}
}

func UpgradeWS() func(c net.Conn) (net.Conn, poll.Messages, error) {
	return func(c net.Conn) (net.Conn, poll.Messages, error) {
		ws := websocket.UpgradeConn(c)
		return ws, ws, nil
	}
}

func UpgradeTLSWS(config *tls.Config) func(c net.Conn) (net.Conn, poll.Messages, error) {
	return func(c net.Conn) (net.Conn, poll.Messages, error) {
		tlsConn := tls.Server(c, config)
		if err := tlsConn.Handshake(); err != nil {
			tlsConn.Close()
			return nil, nil, err
		}
		ws := websocket.UpgradeConn(tlsConn)
		return ws, ws, nil
	}
}

func UpgradeHTTP() func(c net.Conn) (net.Conn, poll.Messages, error) {
	return func(c net.Conn) (net.Conn, poll.Messages, error) {
		conn := upgradeHTTPConn(c)
		return conn, nil, nil
	}
}

func UpgradeHTTPMessages() func(c net.Conn) (net.Conn, poll.Messages, error) {
	return func(c net.Conn) (net.Conn, poll.Messages, error) {
		conn := upgradeHTTPConn(c)
		messages := NewMessages(conn, 0, 0)
		return nil, messages, nil
	}
}

func UpgradeTLSHTTP(config *tls.Config) func(c net.Conn) (net.Conn, poll.Messages, error) {
	return func(c net.Conn) (net.Conn, poll.Messages, error) {
		tlsConn := tls.Server(c, config)
		if err := tlsConn.Handshake(); err != nil {
			tlsConn.Close()
			return nil, nil, err
		}
		conn := upgradeHTTPConn(tlsConn)
		return conn, nil, nil
	}
}

func UpgradeTLSHTTPMessages(config *tls.Config) func(c net.Conn) (net.Conn, poll.Messages, error) {
	return func(c net.Conn) (net.Conn, poll.Messages, error) {
		tlsConn := tls.Server(c, config)
		if err := tlsConn.Handshake(); err != nil {
			tlsConn.Close()
			return nil, nil, err
		}
		conn := upgradeHTTPConn(tlsConn)
		messages := NewMessages(conn, 0, 0)
		return nil, messages, nil
	}
}
