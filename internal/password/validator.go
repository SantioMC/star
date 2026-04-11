package password

import (
	"fmt"
	"strings"
)

func ValidatePassword(password string, settings *PasswordSetttings) (string, error) {
	length := len(password)
	if length < settings.MinLength {
		return password, fmt.Errorf("password must be at least %d characters", settings.MinLength)
	} else if length > settings.MaxLength {
		return password, fmt.Errorf("password must be at most %d characters", settings.MaxLength)
	}

	symbols := 0
	upper := 0
	lower := 0
	numeric := 0

	for _, char := range password {
		if char >= 'a' && char <= 'z' {
			lower++
		} else if char >= 'A' && char <= 'Z' {
			upper++
		} else if char >= '0' && char <= '9' {
			numeric++
		} else {
			symbols++
		}
	}

	if symbols < settings.MinSymbols {
		return password, fmt.Errorf("password must contain at least %d symbols", settings.MinSymbols)
	} else if upper < settings.MinUpper {
		return password, fmt.Errorf("password must contain at least %d uppercase characters", settings.MinUpper)
	} else if lower < settings.MinLower {
		return password, fmt.Errorf("password must contain at least %d lowercase characters", settings.MinLower)
	}

	sanitized := strings.TrimSpace(password)
	return sanitized, nil
}
