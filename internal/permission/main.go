package permission

type PermissionDefinition struct {
	Name        string
	Description string
}

var BuiltinPermissions = []PermissionDefinition{
	PermissionChangePassword, PermissionCreateAuthorization,
	PermissionReadEmails, PermissionAddEmails, PermissionDeleteEmails, PermissionUpdatePrimaryEmail,
	PermissionCreateApplications, PermissionDeleteApplications, PermissionReadApplications,
	PermissionEnrollMFA, PermissionDeleteMFA,
	PermissionDeleteSessions,
	PermissionAdminReadUsers,
}
