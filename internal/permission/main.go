package permission

type PermissionDefinition struct {
	Name        string
	Description string
	Default     bool
}

var BuiltinPermissions = []PermissionDefinition{
	PermissionModifySelf, PermissionChangePassword, PermissionCreateAuthorization, PermissionRevokeAuthorization,
	PermissionReadEmails, PermissionAddEmails, PermissionDeleteEmails, PermissionUpdatePrimaryEmail,
	PermissionCreateApplications, PermissionDeleteApplications, PermissionReadApplications, PermissionWriteApplications,
	PermissionEnrollMFA, PermissionDeleteMFA,
	PermissionDeleteSessions, PermissionExportSelf,
	PermissionBypassNoLogin,
	PermissionAdminReadUsers, PermissionAdminWriteUsers, PermissionAdminDeleteUsers, PermissionAdminExportUsers,
	PermissionAdminReadRoles, PermissionAdminCreateRoles, PermissionAdminWriteRoles, PermissionAdminDeleteRoles,
	PermissionAdminReadProviders, PermissionAdminCreateProviders, PermissionAdminWriteProviders, PermissionAdminDeleteProviders,
	PermissionAdminReadSettings, PermissionAdminWriteSettings,
	PermissionAdminReadKeys, PermissionAdminRotateKeys, PermissionAdminDeleteKeys,
	PermissionAdminReadTokens, PermissionAdminRevokeTokens,
	PermissionAdminReadMetrics, PermissionAdminReadAlerts, PermissionAdminResolveAlerts,
}
