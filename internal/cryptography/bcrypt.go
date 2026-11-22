package cryptography

import (
	"errors"
	"log/slog"

	"github.com/tubblelive/star/internal/password"
	"golang.org/x/crypto/bcrypt"
)

type bcryptHasher struct {
	cost int
}

func (h bcryptHasher) Algorithm() password.Algorithm {
	return password.AlgorithmBcrypt
}

func (h bcryptHasher) Hash(value string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(value), h.cost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

func (h bcryptHasher) Verify(value string, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(value))
	if err != nil && !errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		slog.Error("invalid bcrypt hash", "err", err)
	}

	return err == nil
}

func (h bcryptHasher) NeedsRehash(hash string) bool {
	cost, err := bcrypt.Cost([]byte(hash))
	return err != nil || cost != h.cost
}
