package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/sloweax/socksx/proxy"
	"github.com/sloweax/socksx/proxy/socks4"
	"github.com/sloweax/socksx/proxy/socks5"
	"io"
	"log"
	"net"
	"sync"
)

type StringArray []string

func main() {

	var proxy_files StringArray
	var addr string

	flag.Var(&proxy_files, "p", "load proxies from file")
	flag.StringVar(&addr, "a", "127.0.0.1:1080", "listen on address")
	flag.Parse()

	proxies_index := 0
	proxies_mutex := sync.RWMutex{}
	proxies_list, err := proxy.LoadFiles(proxy_files...)
	if err != nil {
		log.Fatal(err)
	}

	if len(proxies_list) == 0 {
		log.Fatal("no loaded proxies")
	}

	server := new(socks5.Server)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Print(err)
			continue
		}

		go func() {
			defer conn.Close()

			raddr, err := server.Handle(conn)
			if err != nil {
				log.Print(err)
				return
			}

			proxies_mutex.Lock()
			proxies := proxies_list[proxies_index%len(proxies_list)]
			proxies_index += 1
			proxies_mutex.Unlock()

			chain, err := ProxyDialerList(proxies...)
			if err != nil {
				log.Print(err)
				return
			}

			proxy := proxy.New(chain...)
			rconn, err := proxy.Dial("tcp", raddr.String())
			if err != nil {
				log.Print(err)
				return
			}
			defer rconn.Close()

			log.Print(fmt.Sprintf("connection from %s to %s (%s)", conn.RemoteAddr(), raddr.String(), proxy.String()))

			err = Bridge(conn, rconn)

			if err != nil {
				log.Print(err)
			}
		}()
	}
}

func Bridge(a, b net.Conn) error {
	done := make(chan error, 2)

	defer close(done)

	copy := func(a, b net.Conn, done chan error) {
		_, err := io.Copy(a, b)
		a.Close()
		b.Close()
		done <- err
	}

	go copy(a, b, done)
	go copy(b, a, done)

	err := <-done
	err2 := <-done
	if err2 == nil {
		err = nil
	}
	if errors.Is(err, net.ErrClosed) {
		err = nil
	}
	return err
}

func (a *StringArray) String() string {
	return fmt.Sprintf("%v", *a)
}

func (a *StringArray) Set(value string) error {
	*a = append(*a, value)
	return nil
}

func ProxyinfoToDialer(p proxy.ProxyInfo) (proxy.ProxyDialer, error) {
	switch p.Protocol {
	case "socks4", "socks4a":
		return socks4.FromProxyInfo(p)
	case "socks5", "socks5h":
		return socks5.FromProxyInfo(p)
	default:
		return nil, errors.New("unsupported protocol")
	}
}

func ProxyDialerList(proxies ...proxy.ProxyInfo) ([]proxy.ProxyDialer, error) {
	r := make([]proxy.ProxyDialer, 0, len(proxies))
	for _, p := range proxies {
		proxy, err := ProxyinfoToDialer(p)
		if err != nil {
			return nil, err
		}
		r = append(r, proxy)
	}
	return r, nil
}
