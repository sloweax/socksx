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
	split := make([][]string, 0)
	opts := make([]string, 0)

	for _, a := range args {
		if a == "|" {
			tmp := make([]string, len(opts))
			copy(tmp, opts)
			split = append(split, tmp)
			opts = opts[:0]
		} else {
			opts = append(opts, a)
		}
	}

	split = append(split, opts)

	r := make([]ProxyInfo, 0, len(split))

	for _, opts := range split {
		if len(opts) < 2 {
			return nil, errors.New("proxy: invalid proxy chain")
		}
		p := ProxyInfo{}
		p.Protocol = opts[0]
		p.Address = opts[1]
		if len(opts) > 2 {
			p.Args = opts[2:]
		}
		r = append(r, p)
	}

	return r, nil
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
		fields := parseFields(line)
		if len(fields) == 0 {
			continue
		}
		p, err := parseChain(fields)
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
