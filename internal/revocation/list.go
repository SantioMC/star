package revocation

import (
	"sync"
	"time"

	"github.com/tubblelive/star/internal/database"
)

type entry struct {
	cutoff  time.Time
	expires time.Time
}

type List struct {
	sync.RWMutex
	entries   map[string]entry
	watermark time.Time
}

func NewList() *List {
	return &List{entries: map[string]entry{}}
}

func (list *List) Apply(revocations []database.TokenRevocation) {
	list.Lock()
	defer list.Unlock()

	now := time.Now()
	for subject, existing := range list.entries {
		if existing.expires.Before(now) {
			delete(list.entries, subject)
		}
	}

	for _, revocation := range revocations {
		existing, ok := list.entries[revocation.Subject]
		if !ok || revocation.RevokedAt.After(existing.cutoff) {
			list.entries[revocation.Subject] = entry{cutoff: revocation.RevokedAt, expires: revocation.ExpiresAt}
		}

		if revocation.RevokedAt.After(list.watermark) {
			list.watermark = revocation.RevokedAt
		}
	}
}

func (list *List) Watermark() time.Time {
	list.RLock()
	defer list.RUnlock()

	return list.watermark
}

func (list *List) Revoked(issuedAt int64, subjects ...string) bool {
	list.RLock()
	defer list.RUnlock()

	for _, subject := range subjects {
		existing, ok := list.entries[subject]
		if ok && issuedAt <= existing.cutoff.Unix() {
			return true
		}
	}

	return false
}
