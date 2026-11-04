package permission

var PermissionAdminReadSettings = PermissionDefinition{
	Name:        "admin:settings:read",
	Description: "Allows viewing the server configuration",
}

var PermissionAdminWriteSettings = PermissionDefinition{
	Name:        "admin:settings:write",
	Description: "Allows modifying the server configuration",
}

var PermissionAdminReadAudit = PermissionDefinition{
	Name:        "admin:audit:read",
	Description: "Allows viewing the audit log of every user",
}

var PermissionAdminWriteAudit = PermissionDefinition{
	Name:        "admin:audit:write",
	Description: "Allows changing how long audit logs are kept and whether they are anonymized",
}
