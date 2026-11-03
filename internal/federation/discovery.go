package federation

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/tubblelive/star/internal/database"
)

type Endpoints struct {
	Issuer           string `json:"issuer"`
	AuthorizationURL string `json:"authorization_endpoint"`
	TokenURL         string `json:"token_endpoint"`
	UserinfoURL      string `json:"userinfo_endpoint"`
	JWKSURL          string `json:"jwks_uri"`
}

type cachedEntry[T any] struct {
	value     *T
	fetchedAt time.Time
}

type cache[T any] struct {
	sync.Mutex
	entries map[string]cachedEntry[T]
}

var discoveryCache = cache[Endpoints]{entries: map[string]cachedEntry[Endpoints]{}}
var keySetCache = cache[jose.JSONWebKeySet]{entries: map[string]cachedEntry[jose.JSONWebKeySet]{}}

func (cache *cache[T]) get(key string, maxAge time.Duration) *T {
	cache.Lock()
	defer cache.Unlock()

	entry, ok := cache.entries[key]
	if !ok || time.Since(entry.fetchedAt) > maxAge {
		return nil
	}

	return entry.value
}

func (cache *cache[T]) set(key string, value *T) {
	cache.Lock()
	defer cache.Unlock()

	cache.entries[key] = cachedEntry[T]{value: value, fetchedAt: time.Now()}
}

func ResolveEndpoints(ctx context.Context, provider *database.Provider) (*Endpoints, error) {
	if provider.OIDCUrl == nil {
		endpoints := &Endpoints{
			AuthorizationURL: *provider.AuthorizationURL,
			TokenURL:         *provider.TokenURL,
			UserinfoURL:      *provider.UserinfoURL,
		}

		if err := requireSecure(endpoints.AuthorizationURL, endpoints.TokenURL, endpoints.UserinfoURL); err != nil {
			return nil, err
		}

		return endpoints, nil
	}

	if err := requireSecure(*provider.OIDCUrl); err != nil {
		return nil, err
	}

	if cached := discoveryCache.get(*provider.OIDCUrl, time.Hour); cached != nil {
		return cached, nil
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, *provider.OIDCUrl, nil)
	if err != nil {
		return nil, err
	}

	var endpoints Endpoints
	if err := doJSON(request, &endpoints); err != nil {
		return nil, fmt.Errorf("failed to fetch discovery document: %w", err)
	}

	if err := validateDiscovery(*provider.OIDCUrl, &endpoints); err != nil {
		return nil, err
	}

	discoveryCache.set(*provider.OIDCUrl, &endpoints)
	return &endpoints, nil
}

func validateDiscovery(discoveryURL string, endpoints *Endpoints) error {
	if endpoints.Issuer == "" || endpoints.AuthorizationURL == "" || endpoints.TokenURL == "" || endpoints.JWKSURL == "" {
		return errors.New("discovery document is missing the issuer, authorization, token or jwks endpoint")
	}

	templated := strings.Contains(endpoints.Issuer, "{")
	expected := strings.TrimSuffix(endpoints.Issuer, "/") + "/.well-known/openid-configuration"

	if !templated && discoveryURL != expected {
		return errors.New("discovery document issuer does not match the url it was fetched from")
	}

	return requireSecure(endpoints.AuthorizationURL, endpoints.TokenURL, endpoints.UserinfoURL, endpoints.JWKSURL)
}

func requireSecure(endpoints ...string) error {
	for _, endpoint := range endpoints {
		if endpoint == "" {
			continue
		}

		parsed, err := url.Parse(endpoint)
		if err != nil || parsed.Scheme != "https" || parsed.Host == "" || parsed.User != nil {
			return fmt.Errorf("provider endpoint %q must be an https url", endpoint)
		}
	}

	return nil
}

func fetchKeySet(ctx context.Context, jwksURL string, maxAge time.Duration) (*jose.JSONWebKeySet, error) {
	if cached := keySetCache.get(jwksURL, maxAge); cached != nil {
		return cached, nil
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, jwksURL, nil)
	if err != nil {
		return nil, err
	}

	var keySet jose.JSONWebKeySet
	if err := doJSON(request, &keySet); err != nil {
		return nil, fmt.Errorf("failed to fetch provider signing keys: %w", err)
	}

	keySetCache.set(jwksURL, &keySet)
	return &keySet, nil
}
