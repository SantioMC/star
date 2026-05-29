package permission

var PermissionReadSessions = PermissionDefinition{
	Name:        "user:session:read",
	Description: "Allows reading the user's current ongoing sessions",
	Private:     true,
}

var PermissionDeleteSessions = PermissionDefinition{
	Name:        "user:session:delete",
	Description: "Allows deleting the user's current, specific, or all ongoing sessions",
	Private:     true,
}
