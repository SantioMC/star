package model

type Suspension struct {
	BaseModel

	User   *User  `json:"-"`
	UserID string `json:"user_id"`

	IssuedBy   *User  `json:"-"`
	IssuedByID string `json:"issued_by"`

	Reason    string `json:"reason"`
	ExpiresAt *int64 `json:"expires_at"`
}

func NewSuspension(user *User, issuedBy *User, reason string, expiresAt *int64) Suspension {
	return Suspension{
		BaseModel: BaseModel{
			ID: CreateTypedId("sus", 12),
		},

		User:     user,
		IssuedBy: issuedBy,

		Reason:    reason,
		ExpiresAt: expiresAt,
	}
}
