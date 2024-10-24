package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"

	"github.com/sloweax/socksx/proxy"
	"github.com/sloweax/socksx/proxy/socks4"
	"github.com/sloweax/socksx/proxy/socks5"
)

type StringArray []string

func main() {

	var (
		proxy_files StringArray
		addr        string
		verbose     bool
		retry       int
	)

	flag.Var(&proxy_files, "c", "load config file")
	flag.StringVar(&addr, "a", "127.0.0.1:1080", "listen on address")
	flag.IntVar(&retry, "r", 0, "retry chain connection x times until success")
	flag.BoolVar(&verbose, "verbose", false, "log additional info")
	flag.Parse()

	picker := proxy.RoundRobin{}

	for _, file := range proxy_files {
		f, err := os.Open(file)
		if err != nil {
			log.Fatal(err)
		}

		if err := picker.Load(f); err != nil {
			log.Fatal(err)
		}

		f.Close()
	}

	if len(proxy_files) == 0 {
		log.Print("no specified config files, reading from stdin")
		err := picker.Load(os.Stdin)
		if err != nil {
			log.Fatal(err)
		}
	}

	if picker.Len() == 0 {
		log.Fatal("no loaded proxies")
	}

	if verbose {
		for i, ps := range picker.All() {
			chain := make([]string, len(ps))
			for i, p := range ps {
				chain[i] = p.String()
			}
			log.Printf("chain %d: %s", i, strings.Join(chain, " | "))
		}
	}

	server := new(socks5.Server)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Print(err)
			continue
		}

		go func() {
			defer conn.Close()

			var (
				err   error
				rconn net.Conn
				chain []proxy.ProxyDialer
			)

			raddr, err := server.Handle(conn)
			if err != nil {
				log.Print(err)
				return
			}

			for i := 0; i < retry+1; i++ {
				proxies := picker.Next()

				chain, err = ProxyDialerList(proxies...)
				if err != nil {
					log.Print(err)
					return
				}

				proxy := proxy.New(chain...)
				rconn, err = proxy.Dial("tcp", raddr.String())
				if err != nil {
					log.Print(err)
					continue
				}
				defer rconn.Close()

				log.Print(fmt.Sprintf("connection from %s to %s (%s)", conn.RemoteAddr(), raddr.String(), proxy.String()))

				break
			}

			if err != nil {
				return
			}

			err = Bridge(conn, rconn)

			if err != nil {
				log.Print(err)
			}
		}()
	}
}

func Bridge(a, b io.ReadWriteCloser) error {
	done := make(chan error, 2)

	defer close(done)

	copy := func(a, b io.ReadWriteCloser, done chan error) {
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
