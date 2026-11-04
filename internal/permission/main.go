package permission

type PermissionDefinition struct {
	Name        string
	Description string
	Default     bool
}

var BuiltinPermissions = []PermissionDefinition{
	PermissionModifySelf, PermissionChangePassword, PermissionResetPassword, PermissionCreateAuthorization, PermissionRevokeAuthorization,
	PermissionReadEmails, PermissionAddEmails, PermissionDeleteEmails, PermissionUpdatePrimaryEmail,
	PermissionCreateApplications, PermissionDeleteApplications, PermissionReadApplications, PermissionWriteApplications,
	PermissionEnrollMFA, PermissionDeleteMFA,
	PermissionDeleteSessions, PermissionExportSelf,
	PermissionBypassNoLogin,
	PermissionAdminReadUsers, PermissionAdminWriteUsers, PermissionAdminDeleteUsers, PermissionAdminExportUsers, PermissionAdminSuspendUsers, PermissionAdminRequirePasswordChange,
	PermissionAdminReadRoles, PermissionAdminCreateRoles, PermissionAdminWriteRoles, PermissionAdminDeleteRoles,
	PermissionAdminReadProviders, PermissionAdminCreateProviders, PermissionAdminWriteProviders, PermissionAdminDeleteProviders,
	PermissionAdminReadSettings, PermissionAdminWriteSettings,
	PermissionAdminReadAudit, PermissionAdminWriteAudit,
	PermissionAdminReadKeys, PermissionAdminRotateKeys, PermissionAdminDeleteKeys,
	PermissionAdminReadTokens, PermissionAdminRevokeTokens,
	PermissionAdminReadMetrics, PermissionAdminReadAlerts, PermissionAdminResolveAlerts,
}
