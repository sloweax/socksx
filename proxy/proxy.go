package proxy

import (
	"net"
)

type ProxyDialer interface {
	net.Addr
	Protocol() string
	KWArgs() map[string]string
	DialWithConn(conn net.Conn, network, address string) (net.Conn, error)
}
