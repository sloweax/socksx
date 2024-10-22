package socks5

import (
	"errors"
	"io"
	"math"
	"net"
)

type Dialer struct {
	address string
	network string
	config  Config
}

type Config struct {
	methods  []Method
	Username string
	Password string
}

func NewDialer(network, address string, config Config) *Dialer {
	d := new(Dialer)
	d.network = network
	d.address = address
	d.config = config
	return d
}

func (d *Dialer) Protocol() string {
	return "socks5"
}

func (d *Dialer) String() string {
	return d.address
}

func (d *Dialer) Network() string {
	return d.network
}

func (d *Dialer) request(rw io.ReadWriter, cmd Command, addr string) error {
	address, err := NewAddress(addr)
	if err != nil {
		return err
	}

	buf := make([]byte, 3, 4+1+255+2)
	buf[0] = Version
	buf[1] = byte(cmd)
	buf[2] = 0
	buf = append(buf, address.Bytes()...)

	_, err = rw.Write(buf)

	return err
}

func (d *Dialer) response(rw io.ReadWriter) (Reply, Addr, error) {
	buf := make([]byte, 3)

	if _, err := io.ReadFull(rw, buf); err != nil {
		return 0xff, Addr{}, err
	}

	reply := Reply(buf[1])

	if buf[0] != Version {
		return reply, Addr{}, errors.New("socks5: unknown version")
	}

	if buf[2] != 0 {
		return reply, Addr{}, errors.New("socks5: invalid rsv")
	}

	bnd, err := ReadAddress(rw)
	if err != nil {
		return reply, Addr{}, err
	}

	return reply, bnd, nil
}

func (d *Dialer) negotiateMethods(rw io.ReadWriter) (Method, error) {
	if len(d.config.methods) == 0 {
		return MethodNotAcceptable, errors.New("socks5: no methods")
	}

	if len(d.config.methods) > math.MaxUint8 {
		return MethodNotAcceptable, errors.New("socks5: too many methods")
	}

	buf := make([]byte, 2+len(d.config.methods))
	buf[0] = Version // ver
	buf[1] = byte(len(d.config.methods))
	for i, m := range d.config.methods {
		buf[i+2] = byte(m)
	}

	if _, err := rw.Write(buf); err != nil {
		return MethodNotAcceptable, err
	}

	if _, err := io.ReadFull(rw, buf[:2]); err != nil {
		return MethodNotAcceptable, err
	}

	if buf[0] != Version {
		return MethodNotAcceptable, errors.New("socks5: unknown version")
	}

	m := Method(buf[1])

	return m, nil
}

func (d *Dialer) handleAuth(rw io.ReadWriter, m Method) error {
	if !d.config.hasMethod(m) {
		return errors.New("socks5: unsupported method")
	}

	switch m {
	case MethodUserPass:
		return d.userPassAuth(rw)
	case MethodNoAuth:
		return nil
	case MethodNotAcceptable:
		return errors.New("socks5: method not acceptable")
	default:
		return errors.New("socks5: unknown method")
	}
}

func (d *Dialer) userPassAuth(rw io.ReadWriter) error {
	if len(d.config.Username) > math.MaxUint8 || len(d.config.Password) > math.MaxUint8 {
		return errors.New("socks5: username/password is too big")
	}

	buf := make([]byte, 1, 2+len(d.config.Username)+1+len(d.config.Password))
	buf[0] = VersionUserPass
	buf = append(buf, byte(len(d.config.Username)))
	buf = append(buf, []byte(d.config.Username)...)
	buf = append(buf, byte(len(d.config.Password)))
	buf = append(buf, []byte(d.config.Password)...)

	if _, err := rw.Write(buf); err != nil {
		return err
	}

	if _, err := io.ReadFull(rw, buf[:2]); err != nil {
		return err
	}

	if buf[0] != VersionUserPass {
		return errors.New("socks5: unknown username/password version")
	}
	if buf[1] != 0x00 {
		return errors.New("socks5: invalid username/password")
	}

	return nil
}

func (d *Dialer) DialWithConn(conn net.Conn, network, address string) (net.Conn, error) {
	if network != "tcp" {
		return nil, errors.New("socks5: tcp only")
	}

	method, err := d.negotiateMethods(conn)
	if err != nil {
		return nil, err
	}

	if err = d.handleAuth(conn, method); err != nil {
		return nil, err
	}

	if err = d.request(conn, CmdConnect, address); err != nil {
		return nil, err
	}

	reponse, bnd, err := d.response(conn)
	if err != nil {
		return nil, err
	}

	c := new(Conn)
	c.bnd = bnd
	c.remote, _ = NewAddress(address)
	c.conn = conn
	return c, reponse.Err("socks5")
}

func (c *Config) hasMethod(method Method) bool {
	for _, m := range c.methods {
		if m == method {
			return true
		}
	}
	return false
}
