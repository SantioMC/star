package verifier

import (
	"fmt"
	"regexp"
	"strings"
)

var ProviderRegex *regexp.Regexp = regexp.MustCompile("^[a-zA-Z0-9]{1,32}$")
var ErrProviderNameNotAllowed = fmt.Errorf("the provider name specified can not be used")

func VerifyProviderName(name string) error {
	sanitized := strings.ToLower(strings.TrimSpace(name))
	if sanitized == "internal" {
		return ErrProviderNameNotAllowed
	}

	if !ProviderRegex.MatchString(sanitized) {
		return fmt.Errorf("provider name can only contain alphanumeric characters, and no more than 32 characters")
	}

	return nil
}