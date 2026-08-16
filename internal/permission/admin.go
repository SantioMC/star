package permission

var PermissionAdminReadUsers = PermissionDefinition{
	Name:        "admin:user:read",
	Description: "Allows viewing all users within the server",
}

var PermissionAdminWriteUsers = PermissionDefinition{
	Name:        "admin:user:write",
	Description: "Allows modifying users and their assigned roles",
}

var PermissionAdminDeleteUsers = PermissionDefinition{
	Name:        "admin:user:delete",
	Description: "Allows deleting users and their sessions",
}

var PermissionAdminReadRoles = PermissionDefinition{
	Name:        "admin:role:read",
	Description: "Allows viewing roles and their permissions",
}

var PermissionAdminCreateRoles = PermissionDefinition{
	Name:        "admin:role:create",
	Description: "Allows creating new roles",
}

var PermissionAdminWriteRoles = PermissionDefinition{
	Name:        "admin:role:write",
	Description: "Allows modifying roles and their permissions",
}

var PermissionAdminDeleteRoles = PermissionDefinition{
	Name:        "admin:role:delete",
	Description: "Allows deleting roles",
}

var PermissionAdminReadProviders = PermissionDefinition{
	Name:        "admin:provider:read",
	Description: "Allows viewing configured social login providers",
}

var PermissionAdminCreateProviders = PermissionDefinition{
	Name:        "admin:provider:create",
	Description: "Allows adding new social login providers",
}

var PermissionAdminWriteProviders = PermissionDefinition{
	Name:        "admin:provider:write",
	Description: "Allows modifying social login providers",
}

var PermissionAdminDeleteProviders = PermissionDefinition{
	Name:        "admin:provider:delete",
	Description: "Allows deleting social login providers",
}
