package oauth

import (
	"errors"
	"fmt"
	"net/netip"
	"net/url"
	"regexp"
	"slices"
	"strings"

	"github.com/tubblelive/star/internal/database"
)

var privateSchemePattern = regexp.MustCompile(`^[a-z][a-z0-9+-]*(\.[a-z0-9+-]+)+$`)

func ValidateRedirectURI(value string) error {
	if len(value) > 255 {
		return errors.New("redirect URIs may not be longer than 255 characters")
	}

	if strings.Contains(value, "#") {
		return errors.New("redirect URIs may not contain a fragment")
	}

	parsed, err := url.Parse(value)
	if err != nil || parsed.Scheme == "" {
		return errors.New("redirect URIs must be absolute URIs")
	}

	if parsed.User != nil {
		return errors.New("redirect URIs may not contain credentials")
	}

	if parsed.Scheme == "https" || parsed.Scheme == "http" {
		return validateWebURL(parsed, "redirect URIs")
	}

	if !privateSchemePattern.MatchString(parsed.Scheme) {
		return errors.New("custom redirect URI schemes must be a reverse domain name, such as com.example.app")
	}

	return nil
}

func ValidateApplicationURL(value string) error {
	if len(value) > 255 {
		return errors.New("application URLs may not be longer than 255 characters")
	}

	parsed, err := url.Parse(value)
	if err != nil || (parsed.Scheme != "https" && parsed.Scheme != "http") {
		return errors.New("application URLs must be absolute http or https URLs")
	}

	if parsed.User != nil {
		return errors.New("application URLs may not contain credentials")
	}

	return validateWebURL(parsed, "application URLs")
}

func validateWebURL(parsed *url.URL, label string) error {
	if parsed.Hostname() == "" {
		return fmt.Errorf("%s must include a host", label)
	}

	if parsed.Scheme == "http" && !loopbackHost(parsed.Hostname()) {
		return fmt.Errorf("only loopback hosts (localhost, 127.0.0.1, [::1]) may use http %s", label)
	}

	return nil
}

func AllowedRedirectURI(application *database.Application, value string) bool {
	if ValidateRedirectURI(value) != nil {
		return false
	}

	if slices.Contains(application.RedirectURIs, value) {
		return true
	}

	return application.DevelopmentMode && loopbackRedirectURI(value)
}

func loopbackRedirectURI(value string) bool {
	parsed, err := url.Parse(value)
	if err != nil {
		return false
	}

	web := parsed.Scheme == "http" || parsed.Scheme == "https"
	return web && loopbackHost(parsed.Hostname())
}

func loopbackHost(host string) bool {
	if host == "localhost" {
		return true
	}

	address, err := netip.ParseAddr(host)
	return err == nil && address.IsLoopback()
}
