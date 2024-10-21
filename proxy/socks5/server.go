package socks5

import (
	"errors"
	"io"
	"net"
)

type Server struct {
}

func (s *Server) Handle(conn net.Conn) (Addr, error) {
	if err := s.NegotiateMethods(conn); err != nil {
		return Addr{}, err
	}

	reply, _, addr, err := s.GetRequest(conn)
	if err != nil {
		return Addr{}, err
	}

	bnd, _ := NewAddress(conn.LocalAddr().String())

	if err := s.Reply(conn, reply, bnd); err != nil {
		return Addr{}, err
	}

	if reply != ReplyOK {
		return Addr{}, reply.Error("server")
	}

	return addr, nil
}

func (s *Server) NegotiateMethods(rw io.ReadWriter) error {
	buf := make([]byte, 2+255)
	if _, err := io.ReadFull(rw, buf[:2]); err != nil {
		return err
	}

	if buf[0] != Version {
		return errors.New("server: unknown method version")
	}

	nmethods := int(buf[1])
	if nmethods == 0 {
		return errors.New("server: no methods")
	}

	methods := make([]Method, nmethods)
	if _, err := io.ReadFull(rw, buf[:nmethods]); err != nil {
		return err
	}
	for i, b := range buf[:nmethods] {
		methods[i] = Method(b)
	}

	if !hasMethod(methods, MethodNoAuth) {
		return errors.New("server: no supported methods")
	}
	buf = buf[:0]
	buf = append(buf, Version)
	buf = append(buf, byte(MethodNoAuth))
	_, err := rw.Write(buf)
	return err
}

func (s *Server) GetRequest(rw io.ReadWriter) (Reply, Command, Addr, error) {
	buf := make([]byte, 4+255+2)
	if _, err := io.ReadFull(rw, buf[:3]); err != nil {
		return 0xff, 0xff, Addr{}, err
	}
	if buf[0] != Version {
		return 0xff, 0xff, Addr{}, errors.New("server: unknown request version")
	}
	if buf[2] != 0 {
		return 0xff, 0xff, Addr{}, errors.New("server: invalid rsv")
	}

	cmd := Command(buf[1])
	if cmd != CmdConnect {
		return ReplyCmdNotSupported, cmd, Addr{}, nil
	}

	addr, err := ReadAddress(rw)
	if err != nil {
		return 0xff, cmd, Addr{}, errors.New("server: failed to read address")
	}

	return ReplyOK, cmd, addr, nil
}

func (s *Server) Reply(rw io.ReadWriter, reply Reply, address Addr) error {
	buf := make([]byte, 0, 4+255+2)
	buf = append(buf, Version)
	buf = append(buf, byte(reply))
	buf = append(buf, 0)
	buf = append(buf, address.Bytes()...)
	if _, err := rw.Write(buf); err != nil {
		return err
	}
	return nil
}

func hasMethod(methods []Method, method Method) bool {
	for _, m := range methods {
		if m == method {
			return true
		}
	}
	return false
}
