package types

type UpdateAuditSettingsRequest struct {
	RetentionDays *int  `json:"retention_days" validate:"omitempty,min=0,max=36500"`
	Anonymize     *bool `json:"anonymize" validate:"omitempty"`
}
