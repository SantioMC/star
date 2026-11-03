package transformer

import (
	"time"

	"github.com/google/uuid"
	"github.com/tubblelive/star/internal/database"
)

type UserPasskey struct {
	ID             string     `json:"id"`
	Name           string     `json:"name"`
	AAGUID         *string    `json:"aaguid"`
	BackupEligible bool       `json:"backup_eligible"`
	Synced         bool       `json:"synced"`
	LastUsedAt     *time.Time `json:"last_used_at"`
	CreatedAt      time.Time  `json:"created_at"`
}

func TransformPasskey(passkey database.Passkey) UserPasskey {
	credential := passkey.Credential.Data

	var aaguid *string
	if parsed, err := uuid.FromBytes(credential.Authenticator.AAGUID); err == nil && parsed != uuid.Nil {
		aaguid = new(parsed.String())
	}

	return UserPasskey{
		ID:             passkey.ID,
		Name:           passkey.Name,
		AAGUID:         aaguid,
		BackupEligible: credential.Flags.BackupEligible,
		Synced:         credential.Flags.BackupState,
		LastUsedAt:     passkey.LastUsedAt,
		CreatedAt:      passkey.CreatedAt,
	}
}

func TransformPasskeys(passkeys []database.Passkey) []UserPasskey {
	serialized := []UserPasskey{}

	for _, value := range passkeys {
		passkey := TransformPasskey(value)
		serialized = append(serialized, passkey)
	}

	return serialized
}
