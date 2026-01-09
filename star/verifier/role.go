package verifier

import (
	"fmt"
	"strings"
)

func VerifyRoleName(name string) (string, error) {
	sanitized := strings.TrimSpace(name)

	if len(sanitized) == 0 {
		return "", fmt.Errorf("role name can not be empty")
	}

	if len(sanitized) > 32 {
		return "", fmt.Errorf("role name can not be longer than 32 characters")
	}

	return sanitized, nil
}

func VerifyPermission(permission string) (string, error) {
	sanitized := strings.TrimSpace(permission)

	if len(sanitized) == 0 {
		return "", fmt.Errorf("permission '%s' can not be empty", permission)
	}

	if len(sanitized) >= 128 {
		return "", fmt.Errorf("permission '%s' can not be longer than 128 characters", permission)
	}

	return sanitized, nil
}