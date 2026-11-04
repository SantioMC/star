package types

import "github.com/tubblelive/star/internal/database"

type AuditLogPage struct {
	Items []database.AuditLog `json:"items"`
	Next  *string             `json:"next"`
}

type UpdateAuditSettingsRequest struct {
	RetentionDays *int  `json:"retention_days" validate:"omitempty,min=0,max=36500"`
	Anonymize     *bool `json:"anonymize" validate:"omitempty"`
}
