package permission

type PermissionDefinition struct {
	Name        string
	Description string
}

var BuiltinPermissions = []PermissionDefinition{
	PermissionChangePassword, PermissionCreateAuthorization,
	PermissionReadEmails, PermissionAddEmails, PermissionDeleteEmails, PermissionUpdatePrimaryEmail,
	PermissionCreateApplications, PermissionDeleteApplications, PermissionReadApplications, PermissionWriteApplications,
	PermissionEnrollMFA, PermissionDeleteMFA,
	PermissionDeleteSessions,
	PermissionBypassNoLogin,
	PermissionAdminReadUsers, PermissionAdminWriteUsers, PermissionAdminDeleteUsers,
	PermissionAdminReadRoles, PermissionAdminCreateRoles, PermissionAdminWriteRoles, PermissionAdminDeleteRoles,
	PermissionAdminReadProviders, PermissionAdminCreateProviders, PermissionAdminWriteProviders, PermissionAdminDeleteProviders,
}
