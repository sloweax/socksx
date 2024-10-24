package proxy

import (
	"io"
	"net"
)

type ChainPicker interface {
	Load(io.Reader) error
	Add([]ProxyInfo)
	Next() []ProxyInfo
	All() [][]ProxyInfo
	Len() int
}

type ProxyDialer interface {
	net.Addr
	Protocol() string
	KWArgs() map[string]string
	DialWithConn(conn net.Conn, network, address string) (net.Conn, error)
}
