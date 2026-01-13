package verifier

import (
	"fmt"
	"regexp"
	"strings"
)

var UsernameRegex *regexp.Regexp = regexp.MustCompile("^[.a-zA-Z][.a-zA-Z0-9_-]{1,30}[.a-zA-Z0-9]$")
var EmailRegex *regexp.Regexp = regexp.MustCompile("(?:[a-z0-9!#$%&'*+/=?^_`{|}~-]+(?:\\.[a-z0-9!#$%&'*+/=?^_`{|}~-]+)*|\"(?:[\\x01-\\x08\\x0b\\x0c\\x0e-\\x1f\\x21\\x23-\\x5b\\x5d-\\x7f]|\\\\[\\x01-\\x09\\x0b\\x0c\\x0e-\\x7f])*\")@(?:(?:[a-z0-9](?:[a-z0-9-]*[a-z0-9])?\\.)+[a-z0-9](?:[a-z0-9-]*[a-z0-9])?|\\[(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?|[a-z0-9-]*[a-z0-9]:(?:[\\x01-\\x08\\x0b\\x0c\\x0e-\\x1f\\x21-\\x5a\\x53-\\x7f]|\\\\[\\x01-\\x09\\x0b\\x0c\\x0e-\\x7f])+)\\])")

func VerifyUsername(username string) (string, error) {
	sanitized := strings.TrimSpace(username)

	if !UsernameRegex.MatchString(sanitized) {
		//todo: stupid santio
		return "", fmt.Errorf("username must conform to ^[.a-zA-Z][.a-zA-Z0-9_-]{1,30}[.a-zA-Z0-9]$")
	}

	count := strings.Count(sanitized, ".")
	if count > 1 {
		return "", fmt.Errorf("username can have at most one period (.)")
	}

	return sanitized, nil
}

func VerifyEmail(email string) (string, error) {
	sanitized := strings.TrimSpace(email)

	if !EmailRegex.MatchString(sanitized) {
		return "", fmt.Errorf("invalid email provided")
	}

	return sanitized, nil
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