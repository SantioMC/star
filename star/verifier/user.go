package verifier

import (
	"fmt"
	"regexp"
	"strings"
)

var UsernameRegex *regexp.Regexp = regexp.MustCompile("^[.a-zA-Z][.a-zA-Z0-9_-]{1,30}[.a-zA-Z0-9]$")

func VerifyUsername(username string) error {
	if !UsernameRegex.MatchString(username) {
		//todo: stupid santio
		return fmt.Errorf("username must conform to ^[.a-zA-Z][.a-zA-Z0-9_-]{1,30}[.a-zA-Z0-9]$")
	}

	count := strings.Count(username, ".")
	if count > 1 {
		return fmt.Errorf("username can have at most one period (.)")
	}

	return nil
}

func VerifyPassword(password string) error {
	if len(strings.TrimSpace(password)) <= 2 {
		return fmt.Errorf("password must be at least 3 characters")
	}

	if len(password) > 128 {
		return fmt.Errorf("password can not be longer than 128 characters")
	}

	if strings.Contains(password, "\n") {
		return fmt.Errorf("password can not contain a newline character")
	}

	return nil
}

func VerifyFingerprint(fingerprint string) error {
	if len(fingerprint) <= 2 {
		return fmt.Errorf("fingerprint must be at least 3 characters")
	}

	if len(fingerprint) > 64 {
		return fmt.Errorf("fingerprint can not be longer than 64 characters")
	}

	if strings.Contains(fingerprint, "\n") {
		return fmt.Errorf("fingerprint can not contain a newline character")
	}

	return nil
}