package socks5

import (
	"errors"
	"github.com/sloweax/socksx/proxy"
)

func FromProxyInfo(p proxy.ProxyInfo) (proxy.ProxyDialer, error) {
	if len(p.Args) > 2 {
		return nil, errors.New("socks5: invalid proxy options")
	}
	if len(p.Args) == 0 {
		return NewDialer("tcp", p.Address, Config{}), nil
	}
	config := Config{}
	config.Username = p.Args[0]
	if len(p.Args) == 2 {
		config.Password = p.Args[1]
	}
	return NewDialer("tcp", p.Address, config), nil
}
