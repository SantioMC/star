package cryptography

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"golang.org/x/crypto/argon2"
)

// Create a random salt for hashing
func createSalt(length int) []byte {
	salt := make([]byte, length)

	_, err := rand.Read(salt)
	if err != nil {
		panic("rand.Read() returns an error which is said to not be possible")
	}

	return salt
}

// Hash a value with recommended Argon2id parameters
func HashString(value string) (string, Algorithm) {
	salt := createSalt(16)

	iterations := uint32(3)
	memory := uint32(64 * 1024)
	parallelism := uint8(4)

	hash := argon2.IDKey([]byte(value), salt, iterations, memory, parallelism, 32)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	encoded := fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		memory,
		iterations,
		parallelism,
		b64Salt,
		b64Hash,
	)

	return encoded, AlgorithmArgon2
}

// Verify a hashed value against a plaintext value
func VerifyHash(value string, hash string) bool {
	section := strings.Split(hash[1:], "$")
	if len(section) != 5 {
		slog.Error("invalid hash format", "hash", hash)
		return false
	}

	if section[0] != "argon2id" {
		slog.Error("invalid argon2id constant", "constant", section[0])
		return false
	}

	version, err := strconv.Atoi(section[1][2:])
	if err != nil || version != argon2.Version {
		slog.Error("invalid argon2id version", "version", version, "error", err)
		return false
	}

	var memory uint32
	var iterations uint32
	var parallelism uint8

	_, err = fmt.Sscanf(section[2], "m=%d,t=%d,p=%d", &memory, &iterations, &parallelism)
	if err != nil {
		slog.Error("invalid argon2id parameters", "error", err)
		return false
	}

	salt, err := base64.RawStdEncoding.Strict().DecodeString(section[3])
	if err != nil {
		slog.Error("invalid base64 salt", "error", err)
		return false
	}

	rawBytes, err := base64.RawStdEncoding.Strict().DecodeString(section[4])
	if err != nil {
		slog.Error("invalid base64 hash", "error", err)
		return false
	}

	hashed := argon2.IDKey([]byte(value), salt, iterations, memory, parallelism, uint32(len(rawBytes)))
	return subtle.ConstantTimeCompare(rawBytes, hashed) == 1
}
