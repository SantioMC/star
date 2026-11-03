package username

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var allowedCharacters = regexp.MustCompile(`^[A-Za-z0-9._\- ]+$`)

func ValidateUsername(username string, settings *Settings) (string, error) {
	username = strings.TrimSpace(username)

	if len(username) < settings.MinLength {
		return "", fmt.Errorf("username must be at least %d characters", settings.MinLength)
	}

	if len(username) > settings.MaxLength {
		return "", fmt.Errorf("username must be at most %d characters", settings.MaxLength)
	}

	if !allowedCharacters.MatchString(username) {
		return "", errors.New("username may only contain english letters, numbers, dots, dashes, underscores and spaces")
	}

	if !settings.AllowSpaces && strings.Contains(username, " ") {
		return "", errors.New("username must not contain a space")
	}

	// If dashes are allowed, we will prevent them being a prefix, suffix, or consecutive.
	if !settings.AllowDash && strings.Contains(username, "-") {
		return "", errors.New("username must not contain a dash")
	} else if strings.HasPrefix(username, "-") || strings.HasSuffix(username, "-") {
		return "", errors.New("username must not start or end with a dash")
	} else if strings.Contains(username, "--") {
		return "", errors.New("username must not contain consecutive dashes")
	}

	// If dots are allowed, we will prevent them from being consecutive.
	if !settings.AllowDots && strings.Contains(username, ".") {
		return "", errors.New("username must not contain a dot")
	} else if strings.Contains(username, "..") {
		return "", errors.New("username must not contain consecutive dots")
	}

	// If we allow underscores, we'll allow them to be consecutive or as a prefix or suffix.
	if !settings.AllowUnderscore && strings.Contains(username, "_") {
		return "", errors.New("username must not contain an underscore")
	}

	// Enforce at least a single letter.
	regular := 0
	for _, char := range username {
		if char >= 'a' && char <= 'z' {
			regular++
		} else if char >= 'A' && char <= 'Z' {
			regular++
		}
	}

	if regular == 0 {
		return "", errors.New("username must contain at least one letter")
	}

	return username, nil
}
