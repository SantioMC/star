package config

import (
	"errors"
	"fmt"
	"net/netip"
	"slices"
	"strings"
)

type ProxyConfig struct {
	Enabled  bool     `env:"TRUST_PROXY,false"`
	Header   string   `env:"PROXY_HEADER,X-Forwarded-For"`
	Trusted  []string `env:"TRUSTED_PROXIES,"`
	prefixes []netip.Prefix
}

func (p *ProxyConfig) parse() error {
	for _, value := range p.Trusted {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}

		prefix, err := netip.ParsePrefix(trimmed)
		if err != nil {
			addr, addrErr := netip.ParseAddr(trimmed)
			if addrErr != nil {
				return fmt.Errorf("invalid TRUSTED_PROXIES entry: %s", trimmed)
			}

			prefix = netip.PrefixFrom(addr, addr.BitLen())
		}

		p.prefixes = append(p.prefixes, prefix.Masked())
	}

	if p.Enabled && len(p.prefixes) == 0 {
		return errors.New("TRUSTED_PROXIES is required when TRUST_PROXY is enabled")
	}

	return nil
}

func (p *ProxyConfig) isTrusted(addr netip.Addr) bool {
	return slices.ContainsFunc(p.prefixes, func(prefix netip.Prefix) bool {
		return prefix.Contains(addr.Unmap())
	})
}

func (p *ProxyConfig) ClientIP(remote netip.Addr, header string) string {
	if !p.Enabled || !p.isTrusted(remote) {
		return remote.Unmap().String()
	}

	hops := strings.Split(header, ",")
	for _, hop := range slices.Backward(hops) {
		addr, err := netip.ParseAddr(strings.TrimSpace(hop))
		if err != nil {
			break
		}

		if !p.isTrusted(addr) {
			return addr.Unmap().String()
		}
	}

	return remote.Unmap().String()
}
