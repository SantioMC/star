package outbound

import (
	"errors"
	"net"
	"net/http"
	"net/netip"
	"slices"
	"syscall"
	"time"
)

var ErrRedirect = errors.New("outbound requests do not follow redirects")
var ErrNonPublicAddress = errors.New("outbound requests may only reach public addresses")

var Public = newClient(true)
var Trusted = newClient(false)

var reservedPrefixes = []netip.Prefix{
	netip.MustParsePrefix("0.0.0.0/8"),
	netip.MustParsePrefix("100.64.0.0/10"),
	netip.MustParsePrefix("192.0.0.0/24"),
	netip.MustParsePrefix("198.18.0.0/15"),
	netip.MustParsePrefix("240.0.0.0/4"),
	netip.MustParsePrefix("64:ff9b::/96"),
	netip.MustParsePrefix("64:ff9b:1::/48"),
	netip.MustParsePrefix("2002::/16"),
}

func newClient(public bool) *http.Client {
	dialer := &net.Dialer{
		Timeout:   time.Second * 5,
		KeepAlive: time.Second * 30,
	}

	if public {
		dialer.Control = rejectNonPublic
	}

	return &http.Client{
		Timeout: time.Second * 10,
		Transport: &http.Transport{
			DialContext:           dialer.DialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          50,
			IdleConnTimeout:       time.Second * 90,
			TLSHandshakeTimeout:   time.Second * 5,
			ResponseHeaderTimeout: time.Second * 10,
		},
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return ErrRedirect
		},
	}
}

func rejectNonPublic(_ string, address string, _ syscall.RawConn) error {
	target, err := netip.ParseAddrPort(address)
	if err != nil {
		return err
	}

	ip := target.Addr().Unmap()
	reserved := slices.ContainsFunc(reservedPrefixes, func(prefix netip.Prefix) bool {
		return prefix.Contains(ip)
	})

	if reserved || !ip.IsGlobalUnicast() || ip.IsPrivate() {
		return ErrNonPublicAddress
	}

	return nil
}
