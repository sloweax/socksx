package proxy

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
)

var globalKWArgs = map[string]string{}

type ProxyInfo struct {
	Protocol string
	Address  string
	Args     []string
	KWArgs   map[string]string
}

func parseFields(line string) ([]string, error) {
	ret := make([]string, 0)
	str := strings.Builder{}

	for i := 0; i < len(line); i++ {
		r := rune(line[i])

		switch r {
		case '|':
			if str.Len() != 0 {
				ret = append(ret, str.String())
				str.Reset()
			}
			ret = append(ret, "|")
		case ' ', '\t', '\r', '\n', '\v', '\f':
			if str.Len() != 0 {
				ret = append(ret, str.String())
				str.Reset()
			}
		case '"', '\'':
			if str.Len() != 0 {
				ret = append(ret, str.String())
				str.Reset()
			}
			unquoted, len, err := parseQuoted(line[i:])
			if err != nil {
				return nil, err
			}
			ret = append(ret, unquoted)
			i += len
		default:
			str.WriteRune(r)
		}
	}

	if str.Len() != 0 {
		ret = append(ret, str.String())
	}

	return ret, nil
}

func parseQuoted(line string) (string, int, error) {
	if len(line) == 0 {
		return "", 0, errors.New("expected quote")
	}

	ret := strings.Builder{}
	quote := rune(line[0])
	linelen := len(line)

	switch quote {
	case '"', '\'':
		break
	default:
		return "", 0, errors.New("expected quote")
	}

	for i := 1; i < len(line); i++ {
		r := rune(line[i])
		switch r {
		case '\\':
			if linelen <= i+1 {
				return "", 0, fmt.Errorf("config: string `%s` ended with \\", line)
			}
			next := rune(line[i+1])
			switch next {
			case 'a':
				ret.WriteRune('\a')
			case 'b':
				ret.WriteRune('\b')
			case 't':
				ret.WriteRune('\t')
			case 'n':
				ret.WriteRune('\n')
			case 'f':
				ret.WriteRune('\f')
			case 'r':
				ret.WriteRune('\r')
			case 'v':
				ret.WriteRune('\v')
			default:
				ret.WriteRune(next)
			}
			i += 1
		case '\'', '"':
			if r == quote {
				return ret.String(), i, nil
			}
			ret.WriteRune(r)
		default:
			ret.WriteRune(r)
		}
	}

	return "", 0, fmt.Errorf("config: unterminated string `%s`", line)
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

	var err error
	kwargs := globalKWArgs

	for _, opts := range split {
		p := ProxyInfo{KWArgs: kwargs}

		switch len(opts) {
		case 0:
			return nil, errors.New("config: found invalid proxy chain")
		case 1:
			p.Protocol = opts[0]
		case 2:
			p.Protocol = opts[0]
			p.Address = opts[1]
		default:
			p.Protocol = opts[0]
			p.Address = opts[1]
			p.Args = opts[2:]
		}

		if isKWArgs(p) {
			kwargs, err = handleKWArgs(p, kwargs)
			if err != nil {
				return nil, err
			}
			continue
		}

		if len(opts) < 2 {
			return nil, errors.New("config: found invalid proxy chain")
		}

		r = append(r, p)
	}

	if len(r) == 0 && len(split) >= 1 {
		// changing global kwargs
		globalKWArgs = kwargs
	}

	return r, nil
}

func isKWArgs(p ProxyInfo) bool {
	switch p.Protocol {
	case "set", "unset", "clear":
		return true
	default:
		return false
	}
}

func handleKWArgs(p ProxyInfo, root map[string]string) (map[string]string, error) {
	r := map[string]string{}
	for k, v := range root {
		r[k] = v
	}

	switch p.Protocol {
	case "set":
		if len(p.Args) != 1 {
			return nil, fmt.Errorf("config: expected `set key value`, got `set %s`", p.Address)
		}
		r[p.Address] = p.Args[0]
	case "unset":
		delete(r, p.Address)
	case "clear":
		return map[string]string{}, nil
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
		fields, err := parseFields(line)
		if err != nil {
			return nil, err
		}
		if len(fields) == 0 {
			continue
		}
		p, err := parseChain(fields)
		if err != nil {
			return nil, err
		}
		if len(p) != 0 {
			ps = append(ps, p)
		}
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

func (p *ProxyInfo) String() string {
	a := p.Protocol
	if len(p.Address) != 0 {
		a += " " + p.Address
	}
	for _, arg := range p.Args {
		a += " " + fmt.Sprintf("%q", arg)
	}
	for k, v := range p.KWArgs {
		a += fmt.Sprintf(" %s=%q", k, v)
	}
	return a
}
