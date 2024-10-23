package proxy

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"time"
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
	durationstr, ok := d.proxies[0].KWArgs()["ConnTimeout"]
	if ok {
		duration, err := time.ParseDuration(durationstr)
		if err != nil {
			return nil, err
		}
		dialer.Timeout = duration
	}

	ipconn, err := dialer.Dial(d.proxies[0].Network(), d.proxies[0].String())
	if err != nil {
		return nil, err
	}
	conn := ipconn

	for i, p := range d.proxies[0 : len(d.proxies)-1] {
		err = setDialerConn(conn, p)
		if err != nil {
			conn.Close()
			return nil, err
		}

		pconn, err := p.DialWithConn(conn, d.proxies[i+1].Network(), d.proxies[i+1].String())
		if err != nil {
			conn.Close()
			return nil, err
		}

		conn = pconn
	}

	err = setDialerConn(conn, d.proxies[len(d.proxies)-1])
	if err != nil {
		conn.Close()
		return nil, err
	}

	pconn, err := d.proxies[len(d.proxies)-1].DialWithConn(conn, network, address)
	if err != nil {
		conn.Close()
		return nil, err
	}

	err = conn.SetDeadline(time.Time{})
	if err != nil {
		conn.Close()
		return nil, err
	}

	return pconn, nil
}

func setDialerConn(conn net.Conn, dialer ProxyDialer) error {
	durationstr, ok := dialer.KWArgs()["ConnTimeout"]

	if !ok {
		err := conn.SetDeadline(time.Time{})
		if err != nil {
			return err
		}
		return nil
	}

	duration, err := time.ParseDuration(durationstr)
	if err != nil {
		return err
	}

	err = conn.SetDeadline(time.Now().Add(duration))
	if err != nil {
		return err
	}

	return nil
}
