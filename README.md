# Install
```sh
go install github.com/sloweax/socksx@latest # binary will likely be installed at ~/go/bin
```

# Usage
```
Usage of socksx
  -a string
    	listen on address (default "127.0.0.1:1080")
  -c value
    	load config file
```

# Example
```sh
$ cat proxies.conf
socks5 1.2.3.4:123 user pass
socks5 4.3.2.1:321

$ socksx -c proxies.conf

$ for i in {1..10}; do curl ifconfig.me -x socks5://127.0.0.1:1080; echo; done
1.2.3.4
4.3.2.1
1.2.3.4
4.3.2.1
....
```


# Advanced config example
```sh
# Globally sets connection timeout to 5 seconds
set ConnTimeout 5s
# A duration string is a possibly signed sequence of
# decimal numbers, each with optional fraction and a unit suffix,
# such as "300ms", or "2h45m".
# Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".

# You can also chain proxies
socks5 1.2.3.4:1234 | socks4 4.3.2.1:4321

# Sets specific connection timeout for this chain only
set ConnTimeout 1s | socks5 1.2.3.4:1234
# This is also valid
set ConnTimeout 1s | socks5 1.2.3.4:1234 | set ConnTimeout 3s | socks5 4.3.2.1:4321

# Chains below will have no timeout
unset ConnTimeout

socks5 1.2.3.4:1234

# Read/Write timeout (only applied to the last proxy of the chain)
set ReadTimeout 1s
set WriteTimeout 1s
```

# Supported protocols

- socks5 / socks5h
- socks4 / socks4a
