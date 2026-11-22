package cryptography

import (
	"context"
	"log/slog"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/tubblelive/star/internal/password"
)

var ErrHashingBusy = fiber.NewError(503, "The server is busy, please try again in a moment")

type PasswordHasher interface {
	Algorithm() password.Algorithm
	Hash(ctx context.Context, value string) (string, error)
	Verify(ctx context.Context, value string, hash string) (bool, error)
	SimulateVerify(ctx context.Context, value string) error
	NeedsRehash(hash string) bool
}

type algorithmHasher interface {
	Algorithm() password.Algorithm
	Hash(value string) (string, error)
	Verify(value string, hash string) bool
	NeedsRehash(hash string) bool
}

var dummyHashes sync.Map

var hashingSlots = make(chan struct{}, max(2, runtime.GOMAXPROCS(0)/4))

type configuredHasher struct {
	algorithmHasher
}

func NewPasswordHasher(algorithm password.Algorithm) PasswordHasher {
	return configuredHasher{newAlgorithmHasher(algorithm)}
}

func newAlgorithmHasher(algorithm password.Algorithm) algorithmHasher {
	if algorithm == password.AlgorithmBcrypt {
		return bcryptHasher{cost: 12}
	}

	return argon2idHasher{
		memory:      64 * 1024,
		iterations:  3,
		parallelism: 4,
		saltLength:  16,
		keyLength:   32,
	}
}

func InferAlgorithm(value string) (password.Algorithm, bool) {
	if strings.HasPrefix(value, "$argon2id$") {
		return password.AlgorithmArgon2Id, true
	}

	for _, prefix := range []string{"$2a$", "$2b$", "$2y$"} {
		if strings.HasPrefix(value, prefix) {
			return password.AlgorithmBcrypt, true
		}
	}

	return "", false
}

func acquireHashingSlot(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	select {
	case hashingSlots <- struct{}{}:
		return nil
	case <-ctx.Done():
		slog.Warn("timed out waiting for a password hashing slot", "slots", cap(hashingSlots), "err", ctx.Err())
		return ErrHashingBusy
	}
}

func (h configuredHasher) Hash(ctx context.Context, value string) (string, error) {
	if err := acquireHashingSlot(ctx); err != nil {
		return "", err
	}

	defer func() { <-hashingSlots }()
	return h.algorithmHasher.Hash(value)
}

func (h configuredHasher) Verify(ctx context.Context, value string, hash string) (bool, error) {
	algorithm, ok := InferAlgorithm(hash)
	if !ok {
		slog.Error("unable to infer password hash algorithm")
		return false, nil
	}

	hasher := h.algorithmHasher
	if algorithm != h.Algorithm() {
		hasher = newAlgorithmHasher(algorithm)
	}

	if err := acquireHashingSlot(ctx); err != nil {
		return false, err
	}

	defer func() { <-hashingSlots }()
	return hasher.Verify(value, hash), nil
}

func (h configuredHasher) SimulateVerify(ctx context.Context, value string) error {
	dummy, _ := dummyHashes.LoadOrStore(h.Algorithm(), sync.OnceValue(func() string {
		hash, _ := h.algorithmHasher.Hash(CreateSecureString(32))
		return hash
	}))

	_, err := h.Verify(ctx, value, dummy.(func() string)())
	return err
}
