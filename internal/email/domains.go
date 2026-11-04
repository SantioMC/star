package email

import (
	_ "embed"
	"fmt"
	"slices"
	"strings"
	"sync"

	"github.com/gofiber/fiber/v3"
	"golang.org/x/net/idna"
)

type DomainMode string

const (
	DomainModeDisabled  DomainMode = "disabled"
	DomainModeWhitelist DomainMode = "whitelist"
	DomainModeBlacklist DomainMode = "blacklist"
)

var (
	ErrDomainNotAllowed = fiber.NewError(400, "Email addresses from this domain are not allowed")
	ErrDisposableEmail  = fiber.NewError(400, "Temporary email addresses are not allowed")
)

//go:embed disposable-domains.conf
var disposableList string

var disposableDomains = sync.OnceValue(func() map[string]struct{} {
	domains := map[string]struct{}{}
	for _, domain := range DisposableDomains() {
		domains[domain] = struct{}{}
	}

	return domains
})

func DisposableDomains() []string {
	return strings.Fields(disposableList)
}

type DomainPolicy struct {
	Mode            DomainMode
	Domains         []string
	BlockDisposable bool
}

func (policy DomainPolicy) Check(address string) error {
	if policy.Mode == DomainModeDisabled && !policy.BlockDisposable {
		return nil
	}

	index := strings.LastIndexByte(address, '@')
	domain, err := NormalizeDomain(address[index+1:])
	if err != nil {
		return ErrDomainNotAllowed
	}

	listed := matchesDomain(domain, func(candidate string) bool {
		return slices.Contains(policy.Domains, candidate)
	})

	if policy.Mode == DomainModeWhitelist {
		if listed {
			return nil
		}

		return ErrDomainNotAllowed
	}

	if policy.Mode == DomainModeBlacklist && listed {
		return ErrDomainNotAllowed
	}

	if policy.BlockDisposable && isDisposable(domain) {
		return ErrDisposableEmail
	}

	return nil
}

func isDisposable(domain string) bool {
	disposable := disposableDomains()

	return matchesDomain(domain, func(candidate string) bool {
		_, found := disposable[candidate]
		return found
	})
}

func matchesDomain(domain string, listed func(string) bool) bool {
	for candidate, more := domain, true; more; _, candidate, more = strings.Cut(candidate, ".") {
		if listed(candidate) {
			return true
		}
	}

	return false
}

func NormalizeDomain(domain string) (string, error) {
	trimmed := strings.TrimSuffix(strings.TrimSpace(domain), ".")

	normalized, err := idna.Lookup.ToASCII(trimmed)
	if err != nil {
		return "", err
	}

	if normalized == "" {
		return "", fmt.Errorf("the domain %q is empty", domain)
	}

	return normalized, nil
}

func NormalizeDomains(domains []string) ([]string, error) {
	normalized := make([]string, 0, len(domains))

	for _, domain := range domains {
		parsed, err := NormalizeDomain(domain)
		if err != nil {
			return nil, fmt.Errorf("%q is not a valid domain", domain)
		}

		normalized = append(normalized, parsed)
	}

	slices.Sort(normalized)
	return slices.Compact(normalized), nil
}
