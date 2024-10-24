package socks5

import (
	"errors"

	"github.com/sloweax/socksx/proxy"
)

func FromProxyInfo(p proxy.ProxyInfo) (proxy.ProxyDialer, error) {
	config := Config{}
	config.methods = append(config.methods, MethodNoAuth)

	switch len(p.Args) {
	case 0:
		return NewDialer("tcp", p.Address, p.KWArgs, config), nil
	default:
		return nil, errors.New("socks5: invalid proxy options")
	case 2:
		config.Password = p.Args[1]
		fallthrough
	case 1:
		config.Username = p.Args[0]
		config.methods = append(config.methods, MethodUserPass)
		return NewDialer("tcp", p.Address, p.KWArgs, config), nil
	}
}
