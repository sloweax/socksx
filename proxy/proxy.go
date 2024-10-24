package proxy

import (
	"io"
	"net"
)

type ProxyInfo struct {
	Protocol string
	Address  string
	Args     []string
	KWArgs   map[string]string
}

type Chain []ProxyInfo

type ChainPicker interface {
	Load(io.Reader) error
	Add(Chain)
	Next() Chain
	All() []Chain
	Len() int
}

type ProxyDialer interface {
	net.Addr
	Protocol() string
	KWArgs() map[string]string
	DialWithConn(conn net.Conn, network, address string) (net.Conn, error)
}
