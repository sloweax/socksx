package socks4

import (
	"errors"

	"github.com/sloweax/socksx/proxy"
)

func FromProxyInfo(p proxy.ProxyInfo) (proxy.ProxyDialer, error) {
	if len(p.Args) > 1 {
		return nil, errors.New("socks4: invalid proxy options")
	}
	config := Config{}
	if p.Protocol == "socks4a" {
		config.T = TypeA
	}
	if len(p.Args) == 1 {
		config.ID = p.Args[0]
	}
	return NewDialer("tcp", p.Address, p.KWArgs, config), nil
}
