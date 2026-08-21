package permission

var PermissionAdminReadSettings = PermissionDefinition{
	Name:        "admin:settings:read",
	Description: "Allows viewing the server configuration",
}

var PermissionAdminWriteSettings = PermissionDefinition{
	Name:        "admin:settings:write",
	Description: "Allows modifying the server configuration",
}
