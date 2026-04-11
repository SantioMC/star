package permission

var PermissionCreateApplications = PermissionDefinition{
	Name:        "application:create",
	Description: "Allows creating new applications",
}

var PermissionDeleteApplications = PermissionDefinition{
	Name:        "application:delete",
	Description: "Allows deleting existing applications",
}

var PermissionReadApplications = PermissionDefinition{
	Name:        "application:read",
	Description: "Allows viewing applications registered in star",
}

var PermissionWriteApplications = PermissionDefinition{
	Name:        "application:write",
	Description: "Allows modifying existing applications",
}
