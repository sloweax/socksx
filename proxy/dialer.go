package proxy

import (
	"errors"
	"fmt"
	"net"
	"strings"
)

type Dialer struct {
	proxies []ProxyDialer
}

func New(proxies ...ProxyDialer) Dialer {
	d := Dialer{}
	d.proxies = proxies
	return d
}

func (d *Dialer) String() string {
	a := make([]string, 0, len(d.proxies))
	for _, p := range d.proxies {
		a = append(a, fmt.Sprintf("%s %s", p.Protocol(), p.String()))
	}
	return strings.Join(a, " | ")
}

func (d *Dialer) Dial(network, address string) (net.Conn, error) {
	if len(d.proxies) == 0 {
		return nil, errors.New("no dialers")
	}

	dialer := net.Dialer{}
	ipconn, err := dialer.Dial(d.proxies[0].Network(), d.proxies[0].String())
	if err != nil {
		return nil, err
	}
	conn := ipconn

	for i, p := range d.proxies[0 : len(d.proxies)-1] {
		pconn, err := p.DialWithConn(conn, d.proxies[i+1].Network(), d.proxies[i+1].String())
		if err != nil {
			conn.Close()
			return nil, err
		}
		conn = pconn
	}

	pconn, err := d.proxies[len(d.proxies)-1].DialWithConn(conn, network, address)
	if err != nil {
		conn.Close()
		return nil, err
	}

	return pconn, nil
}
