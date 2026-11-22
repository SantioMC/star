package cryptography

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/tubblelive/star/internal/password"
	"golang.org/x/crypto/argon2"
)

type argon2idHasher struct {
	memory      uint32
	iterations  uint32
	parallelism uint8
	saltLength  int
	keyLength   uint32
}

type argon2idHash struct {
	memory      uint32
	iterations  uint32
	parallelism uint8
	salt        []byte
	key         []byte
}

// Create a random salt for hashing
func createSalt(length int) []byte {
	salt := make([]byte, length)

	_, err := rand.Read(salt)
	if err != nil {
		panic("rand.Read() returns an error which is said to not be possible")
	}

	return salt
}

func parseArgon2id(hash string) (*argon2idHash, error) {
	if hash == "" {
		return nil, errors.New("empty hash")
	}

	section := strings.Split(hash[1:], "$")
	if len(section) != 5 {
		return nil, errors.New("invalid hash format")
	}

	if section[0] != "argon2id" {
		return nil, fmt.Errorf("invalid argon2id constant %q", section[0])
	}

	version, err := strconv.Atoi(strings.TrimPrefix(section[1], "v="))
	if err != nil || version != argon2.Version {
		return nil, fmt.Errorf("invalid argon2id version %q", section[1])
	}

	var parsed argon2idHash

	_, err = fmt.Sscanf(section[2], "m=%d,t=%d,p=%d", &parsed.memory, &parsed.iterations, &parsed.parallelism)
	if err != nil {
		return nil, fmt.Errorf("invalid argon2id parameters: %w", err)
	}

	if parsed.salt, err = base64.RawStdEncoding.Strict().DecodeString(section[3]); err != nil {
		return nil, fmt.Errorf("invalid base64 salt: %w", err)
	}

	if parsed.key, err = base64.RawStdEncoding.Strict().DecodeString(section[4]); err != nil {
		return nil, fmt.Errorf("invalid base64 hash: %w", err)
	}

	return &parsed, nil
}

func (h argon2idHasher) Algorithm() password.Algorithm {
	return password.AlgorithmArgon2Id
}

func (h argon2idHasher) Hash(value string) (string, error) {
	salt := createSalt(h.saltLength)
	hash := argon2.IDKey([]byte(value), salt, h.iterations, h.memory, h.parallelism, h.keyLength)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	encoded := fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		h.memory,
		h.iterations,
		h.parallelism,
		b64Salt,
		b64Hash,
	)

	return encoded, nil
}

func (h argon2idHasher) Verify(value string, hash string) bool {
	parsed, err := parseArgon2id(hash)
	if err != nil {
		slog.Error("invalid argon2id hash", "err", err)
		return false
	}

	hashed := argon2.IDKey([]byte(value), parsed.salt, parsed.iterations, parsed.memory, parsed.parallelism, uint32(len(parsed.key)))
	return subtle.ConstantTimeCompare(parsed.key, hashed) == 1
}

func (h argon2idHasher) NeedsRehash(hash string) bool {
	parsed, err := parseArgon2id(hash)
	if err != nil {
		return true
	}

	return parsed.memory != h.memory ||
		parsed.iterations != h.iterations ||
		parsed.parallelism != h.parallelism ||
		len(parsed.salt) != h.saltLength ||
		len(parsed.key) != int(h.keyLength)
}
