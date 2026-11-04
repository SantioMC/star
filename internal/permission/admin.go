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

var PermissionAdminExportUsers = PermissionDefinition{
	Name:        "admin:user:export",
	Description: "Allows exporting all personal data of any user",
}

var PermissionAdminRequirePasswordChange = PermissionDefinition{
	Name:        "admin:password:require",
	Description: "Allows signing out every user and requiring them to choose a new password",
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

var PermissionAdminReadKeys = PermissionDefinition{
	Name:        "admin:key:read",
	Description: "Allows viewing the token signing keys",
}

var PermissionAdminRotateKeys = PermissionDefinition{
	Name:        "admin:key:rotate",
	Description: "Allows rotating the token signing keys",
}

var PermissionAdminDeleteKeys = PermissionDefinition{
	Name:        "admin:key:delete",
	Description: "Allows deleting retired token signing keys",
}

var PermissionAdminReadTokens = PermissionDefinition{
	Name:        "admin:token:read",
	Description: "Allows viewing revoked sessions and access tokens",
}

var PermissionAdminRevokeTokens = PermissionDefinition{
	Name:        "admin:token:revoke",
	Description: "Allows revoking the access tokens of a user or of every user",
}

var PermissionAdminReadMetrics = PermissionDefinition{
	Name:        "admin:metrics:read",
	Description: "Allows viewing server metrics",
}

var PermissionAdminReadAlerts = PermissionDefinition{
	Name:        "admin:alert:read",
	Description: "Allows viewing security alerts",
}

var PermissionAdminResolveAlerts = PermissionDefinition{
	Name:        "admin:alert:resolve",
	Description: "Allows resolving security alerts",
}
