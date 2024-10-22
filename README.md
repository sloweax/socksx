# Install
```sh
go install github.com/sloweax/socksx@latest # binary will likely be installed at ~/go/bin
```

# Usage
```
Usage of socksx
  -a string
    	listen on address (default "127.0.0.1:1080")
  -p value
    	load proxies from file
```

# Example
```sh
$ cat proxies.txt
socks5 1.2.3.4:123 user pass
socks5 4.3.2.1:321
# You can also chain proxies
socks5 9.8.7.6:1080 | socks5 11.22.33.44 1080

$ socksx -p proxies.txt

$ for i in {1..10}; do curl ifconfig.me -x socks5://127.0.0.1:1080; echo; done
1.2.3.4
4.3.2.1
11.22.33.44
1.2.3.4
4.3.2.1
11.22.33.44
....
```

# Supported protocols

- socks5 / socks5h
