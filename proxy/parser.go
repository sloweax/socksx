package proxy

import (
	"bufio"
	"errors"
	"os"
	"strings"
	"unicode"
)

type ProxyInfo struct {
	Protocol string
	Address  string
	Args     []string
}

func parseFields(line string) []string {
	return strings.FieldsFunc(line, unicode.IsSpace)
}

func parseChain(args []string) ([]ProxyInfo, error) {
	ps := make([]ProxyInfo, 0)

	for {
		p := ProxyInfo{}

		if len(args) < 2 {
			return nil, errors.New("proxy: invalid proxy chain")
		}

		p.Protocol = args[0]
		p.Address = args[1]
		args = args[2:]

		for i, a := range args {
			if a == "|" {
				args = args[i+1:]
				break
			} else {
				p.Args = append(p.Args, a)
			}
		}

		ps = append(ps, p)

		if len(args) == len(p.Args) {
			return ps, nil
		}
	}
}

func loadFile(path string) ([][]ProxyInfo, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	ps := make([][]ProxyInfo, 0)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") {
			continue
		}
		p, err := parseChain(parseFields(line))
		if err != nil {
			return nil, err
		}
		ps = append(ps, p)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return ps, nil
}

func LoadFiles(paths ...string) ([][]ProxyInfo, error) {
	ps := make([][]ProxyInfo, 0)

	for _, path := range paths {
		p, err := loadFile(path)
		if err != nil {
			return nil, err
		}
		ps = append(ps, p...)
	}

	return ps, nil
}
