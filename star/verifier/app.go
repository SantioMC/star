package verifier

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

var AppSlugRegex *regexp.Regexp = regexp.MustCompile("^[a-z0-9_-]{1,32}$")
var ErrApplicationSlugNotAllowed = fmt.Errorf("the application slug specified can not be used")
var ErrUrlIsEmpty = fmt.Errorf("the provided url can not be empty")

func VerifyApplicationSlug(slug string) error {
	sanitized := strings.ToLower(strings.TrimSpace(slug))

	if !ProviderRegex.MatchString(sanitized) {
		return ErrApplicationSlugNotAllowed
	}

	return nil
}

func VerifyURL(value string) error {
	sanitized := strings.ToLower(strings.TrimSpace(value))
	if sanitized == "" {
		return ErrUrlIsEmpty
	}

	if !strings.Contains(value, "://") {
		value = "https://" + value
	}

	parts := strings.Split(strings.TrimSpace(value), "://")
	if len(parts) != 2 {
		return fmt.Errorf("invalid amount of parts")
	}

	if parts[1] == "" {
		return fmt.Errorf("invalid url, can not just be protocol")
	}

	_, err := url.ParseRequestURI(value)
	return err
}