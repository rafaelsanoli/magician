package tor

import (
	"fmt"
	"golang.org/x/net/proxy"
	"net"
	"strings"
)

// DialOrDirect detecta se é um .onion e usa proxy SOCKS5 se necessário
func DialOrDirect(address string) (net.Conn, error) {
	if strings.HasSuffix(address, ".onion:1337") {
		dialer, err := proxy.SOCKS5("tcp", "127.0.0.1:9050", nil, proxy.Direct)
		if err != nil {
			return nil, fmt.Errorf("erro ao criar dialer SOCKS5: %w", err)
		}
		return dialer.Dial("tcp", address)
	}
	return net.Dial("tcp", address)
}
