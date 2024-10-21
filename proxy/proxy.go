package proxy

import (
	"net"
)

type ProxyDialer interface {
	net.Addr
	Protocol() string
	DialWithConn(conn net.Conn, network, address string) (net.Conn, error)
}
