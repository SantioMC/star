package audit

type Action string
type Severity string
type TargetType string

const (
	SeverityInfo     Severity = "info"
	SeverityNotice   Severity = "notice"
	SeverityCritical Severity = "critical"
)

const (
	TargetUser        TargetType = "user"
	TargetRole        TargetType = "role"
	TargetApplication TargetType = "application"
	TargetProvider    TargetType = "provider"
	TargetKey         TargetType = "key"
	TargetAlert       TargetType = "alert"
)

const (
	SessionStart     Action = "session.start"
	SessionEnd       Action = "session.end"
	SessionRevoke    Action = "session.revoke"
	SessionRevokeAll Action = "session.revoke_all"
	LoginFailed      Action = "login.failed"

	PasswordChange         Action = "password.change"
	PasswordRemove         Action = "password.remove"
	PasswordResetRequest   Action = "password.reset_request"
	PasswordReset          Action = "password.reset"
	PasswordSet            Action = "password.set"
	PasswordRequireChange  Action = "password.require_change"
	PasswordRequireAll     Action = "password.require_all"
	PasswordChangeRequired Action = "password.change_required"

	EmailAdd     Action = "email.add"
	EmailRemove  Action = "email.remove"
	EmailPrimary Action = "email.primary"
	EmailVerify  Action = "email.verify"

	MFATotpEnroll   Action = "mfa.totp_enroll"
	MFATotpRemove   Action = "mfa.totp_remove"
	MFAPasskeyAdd   Action = "mfa.passkey_add"
	MFAPasskeyEdit  Action = "mfa.passkey_rename"
	MFAPasskeyDrop  Action = "mfa.passkey_remove"
	MFABackupRotate Action = "mfa.backup_regenerate"

	GrantCreate    Action = "grant.create"
	GrantRevoke    Action = "grant.revoke"
	GrantAuthorize Action = "grant.authorize"

	DeviceApprove Action = "device.approve"
	DeviceDeny    Action = "device.deny"

	UserRegister     Action = "user.register"
	UserUpdate       Action = "user.update"
	UserDelete       Action = "user.delete"
	UserExport       Action = "user.export"
	UserRolesUpdate  Action = "user.roles_update"
	UserLockoutClear Action = "user.lockout_clear"
	UserSuspend      Action = "user.suspend"
	UserUnsuspend    Action = "user.unsuspend"

	UserInvite           Action = "user.invite"
	UserInviteRegenerate Action = "user.invite_regenerate"
	UserInviteEmail      Action = "user.invite_email"
	UserInviteAccept     Action = "user.invite_accept"
	UserInviteExpire     Action = "user.invite_expire"

	UserImpersonate       Action = "user.impersonate"
	UserImpersonateSignIn Action = "user.impersonate_sign_in"

	RoleCreate  Action = "role.create"
	RoleUpdate  Action = "role.update"
	RoleDelete  Action = "role.delete"
	RoleReorder Action = "role.reorder"

	ApplicationCreate       Action = "application.create"
	ApplicationUpdate       Action = "application.update"
	ApplicationDelete       Action = "application.delete"
	ApplicationSecretRotate Action = "application.secret_rotate"
	ApplicationScopeCreate  Action = "application.scope_create"
	ApplicationScopeUpdate  Action = "application.scope_update"
	ApplicationScopeDelete  Action = "application.scope_delete"

	ProviderCreate       Action = "provider.create"
	ProviderUpdate       Action = "provider.update"
	ProviderDelete       Action = "provider.delete"
	ProviderAssociate    Action = "provider.associate"
	ProviderDisassociate Action = "provider.disassociate"

	KeyRotate Action = "key.rotate"
	KeyDelete Action = "key.delete"

	TokenRevoke    Action = "token.revoke"
	TokenRevokeAll Action = "token.revoke_all"

	AlertResolve Action = "alert.resolve"

	SettingsUpdate      Action = "settings.update"
	BrandingUpdate      Action = "branding.update"
	AuditSettingsUpdate Action = "audit.settings_update"
)

type Definition struct {
	Name     Action   `json:"name"`
	Label    string   `json:"label"`
	Category string   `json:"category"`
	Severity Severity `json:"severity"`
}

var Definitions = []Definition{
	{SessionStart, "Signed in", "Sessions", SeverityInfo},
	{SessionEnd, "Signed out", "Sessions", SeverityInfo},
	{SessionRevoke, "Session revoked", "Sessions", SeverityNotice},
	{SessionRevokeAll, "All sessions revoked", "Sessions", SeverityNotice},
	{LoginFailed, "Sign in failed", "Sessions", SeverityNotice},

	{PasswordChange, "Password changed", "Passwords", SeverityNotice},
	{PasswordRemove, "Password removed", "Passwords", SeverityNotice},
	{PasswordResetRequest, "Password reset requested", "Passwords", SeverityInfo},
	{PasswordReset, "Password reset", "Passwords", SeverityNotice},
	{PasswordSet, "Password set by admin", "Passwords", SeverityCritical},
	{PasswordRequireChange, "Password change required", "Passwords", SeverityNotice},
	{PasswordRequireAll, "Password change required for all", "Passwords", SeverityCritical},
	{PasswordChangeRequired, "Required password change completed", "Passwords", SeverityInfo},

	{EmailAdd, "Email added", "Emails", SeverityInfo},
	{EmailRemove, "Email removed", "Emails", SeverityNotice},
	{EmailPrimary, "Primary email changed", "Emails", SeverityNotice},
	{EmailVerify, "Email verified", "Emails", SeverityInfo},

	{MFATotpEnroll, "Authenticator app enrolled", "Multi-factor", SeverityInfo},
	{MFATotpRemove, "Authenticator app removed", "Multi-factor", SeverityNotice},
	{MFAPasskeyAdd, "Passkey added", "Multi-factor", SeverityInfo},
	{MFAPasskeyEdit, "Passkey renamed", "Multi-factor", SeverityInfo},
	{MFAPasskeyDrop, "Passkey removed", "Multi-factor", SeverityNotice},
	{MFABackupRotate, "Backup codes regenerated", "Multi-factor", SeverityNotice},

	{GrantCreate, "Application authorized", "Authorizations", SeverityInfo},
	{GrantRevoke, "Application access revoked", "Authorizations", SeverityInfo},
	{GrantAuthorize, "Signed in to application", "Authorizations", SeverityInfo},
	{DeviceApprove, "Device sign in approved", "Authorizations", SeverityNotice},
	{DeviceDeny, "Device sign in denied", "Authorizations", SeverityInfo},

	{UserRegister, "Account registered", "Users", SeverityInfo},
	{UserUpdate, "Profile updated", "Users", SeverityInfo},
	{UserDelete, "User deleted", "Users", SeverityCritical},
	{UserExport, "User data exported", "Users", SeverityNotice},
	{UserRolesUpdate, "User roles changed", "Users", SeverityCritical},
	{UserLockoutClear, "Sign in lockout cleared", "Users", SeverityNotice},
	{UserSuspend, "User suspended", "Users", SeverityCritical},
	{UserUnsuspend, "User unsuspended", "Users", SeverityCritical},
	{UserInvite, "User invited", "Users", SeverityNotice},
	{UserInviteRegenerate, "User invite link regenerated", "Users", SeverityNotice},
	{UserInviteEmail, "User invite emailed", "Users", SeverityInfo},
	{UserInviteAccept, "User invite accepted", "Users", SeverityInfo},
	{UserInviteExpire, "User invite expired", "Users", SeverityInfo},
	{UserImpersonate, "User impersonation started", "Users", SeverityCritical},
	{UserImpersonateSignIn, "Signed in to application as impersonated user", "Users", SeverityCritical},

	{RoleCreate, "Role created", "Roles", SeverityNotice},
	{RoleUpdate, "Role updated", "Roles", SeverityCritical},
	{RoleDelete, "Role deleted", "Roles", SeverityCritical},
	{RoleReorder, "Role order changed", "Roles", SeverityCritical},

	{ApplicationCreate, "Application created", "Applications", SeverityNotice},
	{ApplicationUpdate, "Application updated", "Applications", SeverityNotice},
	{ApplicationDelete, "Application deleted", "Applications", SeverityCritical},
	{ApplicationSecretRotate, "Client secret rotated", "Applications", SeverityCritical},
	{ApplicationScopeCreate, "Application scope created", "Applications", SeverityInfo},
	{ApplicationScopeUpdate, "Application scope updated", "Applications", SeverityInfo},
	{ApplicationScopeDelete, "Application scope deleted", "Applications", SeverityNotice},

	{ProviderCreate, "Social provider created", "Providers", SeverityNotice},
	{ProviderUpdate, "Social provider updated", "Providers", SeverityNotice},
	{ProviderDelete, "Social provider deleted", "Providers", SeverityCritical},
	{ProviderAssociate, "Social provider linked", "Providers", SeverityNotice},
	{ProviderDisassociate, "Social provider unlinked", "Providers", SeverityNotice},

	{KeyRotate, "Signing keys rotated", "Signing Keys", SeverityCritical},
	{KeyDelete, "Signing key deleted", "Signing Keys", SeverityCritical},

	{TokenRevoke, "Access tokens revoked", "Tokens", SeverityNotice},
	{TokenRevokeAll, "All access tokens revoked", "Tokens", SeverityCritical},

	{AlertResolve, "Security alert resolved", "Security", SeverityNotice},

	{SettingsUpdate, "Server settings updated", "Settings", SeverityCritical},
	{BrandingUpdate, "Branding updated", "Settings", SeverityInfo},
	{AuditSettingsUpdate, "Audit log settings updated", "Settings", SeverityCritical},
}

func ActionsInCategory(category string) []string {
	actions := []string{}

	for _, definition := range Definitions {
		if definition.Category == category {
			actions = append(actions, string(definition.Name))
		}
	}

	return actions
}
