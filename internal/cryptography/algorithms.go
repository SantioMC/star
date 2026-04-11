package cryptography

import "strings"

type Algorithm string

var (
	AlgorithmArgon2 Algorithm = "argon2"
)

// IsHashed will try to infer if the given string is hashed and with what algorithm
func IsHashed(value string) *Algorithm {
	if strings.HasPrefix(value, "$argon2id$") {
		return &AlgorithmArgon2
	}

	return nil
}
