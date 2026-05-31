package permission

type PermissionDefinition struct {
	Name        string
	Description string
	Private     bool
}

var BuiltinPermissions = []PermissionDefinition{
	PermissionChangePassword, PermissionCreateAuthorization,
	PermissionReadEmails, PermissionAddEmails, PermissionDeleteEmails, PermissionUpdatePrimaryEmail,
	PermissionCreateApplications, PermissionDeleteApplications, PermissionReadApplications,
	PermissionEnrollMFA, PermissionDeleteMFA,
	PermissionReadSessions, PermissionDeleteSessions,
}
