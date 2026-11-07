package export

import (
	"context"
	"log/slog"
	"time"

	"github.com/tubblelive/star/internal/audit"
	"github.com/tubblelive/star/internal/database"
	"github.com/tubblelive/star/internal/database/models"
)

type UserExport struct {
	ExportedAt     time.Time       `json:"exported_at"`
	Profile        Profile         `json:"profile"`
	Emails         []Email         `json:"emails"`
	Accounts       []Account       `json:"accounts"`
	Roles          []string        `json:"roles"`
	Permissions    []string        `json:"permissions"`
	MFA            MFA             `json:"mfa"`
	Sessions       []Session       `json:"sessions"`
	Authorizations []Authorization `json:"authorizations"`
	LoginAttempts  []LoginAttempt  `json:"login_attempts"`
	SecurityAlerts []SecurityAlert `json:"security_alerts"`
	Suspensions    []Suspension    `json:"suspensions"`
	AuditLogs      []AuditLog      `json:"audit_logs"`
}

type Profile struct {
	ID             string       `json:"id"`
	Username       string       `json:"username"`
	FirstName      *string      `json:"first_name"`
	LastName       *string      `json:"last_name"`
	ProfilePicture *string      `json:"profile_picture"`
	Birthday       *models.Date `json:"birthday"`
	CreatedAt      time.Time    `json:"created_at"`
}

type Email struct {
	Email     string    `json:"email"`
	Verified  bool      `json:"verified"`
	Primary   bool      `json:"primary"`
	CreatedAt time.Time `json:"created_at"`
}

type Account struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`
	ExternalID  *string   `json:"external_id"`
	Name        *string   `json:"name"`
	Email       *string   `json:"email"`
	Avatar      *string   `json:"avatar"`
	HasPassword bool      `json:"has_password"`
	CreatedAt   time.Time `json:"created_at"`
}

type MFA struct {
	TotpEnabled          bool       `json:"totp_enabled"`
	TotpCreatedAt        *time.Time `json:"totp_created_at"`
	BackupCodesRemaining int        `json:"backup_codes_remaining"`
}

type Session struct {
	ID              string    `json:"id"`
	ApplicationID   string    `json:"application_id"`
	ImpersonatorID  *string   `json:"impersonator_id"`
	IPAddress       string    `json:"ip_address"`
	UserAgent       string    `json:"user_agent"`
	Device          string    `json:"device"`
	Geolocation     *string   `json:"geolocation"`
	Scopes          []string  `json:"scopes"`
	Status          string    `json:"status"`
	AuthenticatedAt time.Time `json:"authenticated_at"`
	CreatedAt       time.Time `json:"created_at"`
	ExpiresAt       time.Time `json:"expires_at"`
}

type Authorization struct {
	ApplicationID string     `json:"application_id"`
	RedirectURI   string     `json:"redirect_uri"`
	Scopes        []string   `json:"scopes"`
	CreatedAt     time.Time  `json:"created_at"`
	RedeemedAt    *time.Time `json:"redeemed_at"`
}

type LoginAttempt struct {
	Stage     string    `json:"stage"`
	CreatedAt time.Time `json:"created_at"`
}

type SecurityAlert struct {
	Kind          string      `json:"kind"`
	Severity      string      `json:"severity"`
	Message       string      `json:"message"`
	Data          models.JSON `json:"data"`
	ApplicationID *string     `json:"application_id"`
	CreatedAt     time.Time   `json:"created_at"`
}

type Suspension struct {
	ExpiresAt *time.Time `json:"expires_at"`
	LiftedAt  *time.Time `json:"lifted_at"`
	CreatedAt time.Time  `json:"created_at"`
}

type AuditLog struct {
	Action        string      `json:"action"`
	Actor         string      `json:"actor"`
	TargetType    *string     `json:"target_type"`
	TargetID      *string     `json:"target_id"`
	TargetName    *string     `json:"target_name"`
	ApplicationID *string     `json:"application_id"`
	ProviderID    *string     `json:"provider_id"`
	IPAddress     *string     `json:"ip_address"`
	UserAgent     *string     `json:"user_agent"`
	Metadata      models.JSON `json:"metadata"`
	CreatedAt     time.Time   `json:"created_at"`
}

func User(ctx context.Context, db *database.Database, user *database.User) (*UserExport, error) {
	logs, err := auditLogs(ctx, db, user.ID)
	if err != nil {
		return nil, err
	}

	suspensions, err := db.GetUserSuspensions(ctx, user.ID, nil, 1000)
	if err != nil {
		return nil, err
	}

	return &UserExport{
		ExportedAt: time.Now().UTC(),
		Profile: Profile{
			ID:             user.ID,
			Username:       user.Name,
			FirstName:      user.FirstName,
			LastName:       user.LastName,
			ProfilePicture: user.ProfilePicture,
			Birthday:       user.Birthday,
			CreatedAt:      user.CreatedAt,
		},
		Emails:         convert(db.GetEmailsByUserID(ctx, user.ID), email),
		Accounts:       convert(db.GetAccountsByUserID(ctx, user.ID), account),
		Roles:          convert(db.GetUserRoles(ctx, user.ID), roleName),
		Permissions:    convert(db.GetUserPermissions(ctx, user.ID), permissionName),
		MFA:            mfa(ctx, db, user.ID),
		Sessions:       convert(db.GetUserSessions(ctx, user.ID), session),
		Authorizations: convert(db.GetAuthorizationCodesByUserID(ctx, user.ID), authorization),
		LoginAttempts:  convert(db.GetLoginFlowsByUserID(ctx, user.ID), loginAttempt),
		SecurityAlerts: convert(db.GetSecurityAlertsByUserID(ctx, user.ID), securityAlert),
		Suspensions:    convert(suspensions, suspension),
		AuditLogs:      logs,
	}, nil
}

func auditLogs(ctx context.Context, db *database.Database, userID string) ([]AuditLog, error) {
	exported := []AuditLog{}
	filter := database.AuditLogFilter{UserID: &userID, Limit: 1000}

	for {
		page, err := db.GetAuditLogs(ctx, filter)
		if err != nil {
			return nil, err
		}

		for _, record := range page {
			exported = append(exported, auditLog(record, userID))
		}

		if len(page) < filter.Limit {
			return exported, nil
		}

		last := page[len(page)-1]
		filter.Before = &database.Cursor{CreatedAt: last.CreatedAt, ID: last.ID}
	}
}

func auditLog(record database.AuditLog, userID string) AuditLog {
	log := AuditLog{
		Action:        record.Action,
		Actor:         "unknown",
		TargetType:    record.TargetType,
		TargetID:      record.TargetID,
		TargetName:    record.TargetName,
		ApplicationID: record.Application,
		ProviderID:    record.Provider,
		Metadata:      record.Metadata,
		CreatedAt:     record.CreatedAt,
	}

	if record.ActorID != nil && *record.ActorID == userID {
		log.Actor = "you"
		log.IPAddress = record.Address
		log.UserAgent = record.UserAgent
	} else if record.ActorID != nil {
		log.Actor = "administrator"
	}

	if record.Action == string(audit.UserSuspend) || record.Action == string(audit.UserUnsuspend) {
		log.Metadata = nil
	}

	isOtherUser := record.TargetType != nil && *record.TargetType == "user" && record.TargetID != nil && *record.TargetID != userID
	if isOtherUser {
		log.TargetID = nil
		log.TargetName = nil
	}

	return log
}

func convert[T any, R any](records []T, transform func(T) R) []R {
	converted := make([]R, 0, len(records))
	for _, record := range records {
		converted = append(converted, transform(record))
	}

	return converted
}

func email(record database.UserEmail) Email {
	return Email{
		Email:     record.Email,
		Verified:  record.Verified,
		Primary:   record.Primary,
		CreatedAt: record.CreatedAt,
	}
}

func account(record database.Account) Account {
	return Account{
		ID:          record.ID,
		Type:        record.Type,
		ExternalID:  record.ExternalID,
		Name:        record.Name,
		Email:       record.Email,
		Avatar:      record.Avatar,
		HasPassword: record.Password != nil,
		CreatedAt:   record.CreatedAt,
	}
}

func roleName(record database.Role) string {
	return record.Name
}

func permissionName(record database.Permission) string {
	return record.Name
}

func mfa(ctx context.Context, db *database.Database, userID string) MFA {
	totp := db.GetUserTotp(ctx, userID)
	if totp == nil || !totp.Verified {
		return MFA{}
	}

	remaining, err := db.CountBackupCodes(ctx, userID)
	if err != nil {
		slog.Error("failed to count backup codes for export", "err", err, "user", userID)
	}

	return MFA{
		TotpEnabled:          true,
		TotpCreatedAt:        &totp.CreatedAt,
		BackupCodesRemaining: remaining,
	}
}

func session(record database.Session) Session {
	return Session{
		ID:              record.ID,
		ApplicationID:   record.ApplicationID,
		ImpersonatorID:  record.ImpersonatorID,
		IPAddress:       record.Address,
		UserAgent:       record.UserAgent,
		Device:          record.Device,
		Geolocation:     record.Geolocation,
		Scopes:          record.Scopes,
		Status:          string(record.Status),
		AuthenticatedAt: record.AuthenticatedAt,
		CreatedAt:       record.CreatedAt,
		ExpiresAt:       record.ExpiresAt,
	}
}

func authorization(record database.AuthorizationCode) Authorization {
	return Authorization{
		ApplicationID: record.ApplicationID,
		RedirectURI:   record.RedirectURI,
		Scopes:        record.Scopes,
		CreatedAt:     record.CreatedAt,
		RedeemedAt:    record.RedeemedAt,
	}
}

func loginAttempt(record database.LoginFlow) LoginAttempt {
	return LoginAttempt{
		Stage:     string(record.Stage),
		CreatedAt: record.CreatedAt,
	}
}

func suspension(record database.UserSuspension) Suspension {
	return Suspension{
		ExpiresAt: record.ExpiresAt,
		LiftedAt:  record.LiftedAt,
		CreatedAt: record.CreatedAt,
	}
}

func securityAlert(record database.SecurityAlert) SecurityAlert {
	return SecurityAlert{
		Kind:          string(record.Kind),
		Severity:      string(record.Severity),
		Message:       record.Message,
		Data:          record.Data,
		ApplicationID: record.Application,
		CreatedAt:     record.CreatedAt,
	}
}
