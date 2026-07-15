package permission

var PermissionChangePassword = PermissionDefinition{
	Name:        "user:password:write",
	Description: "Allows changing or removing the user's password",
}

var PermissionCreateAuthorization = PermissionDefinition{
	Name:        "user:application:authorize",
	Description: "Allows the user to authorize an application",
}

var PermissionReadEmails = PermissionDefinition{
	Name:        "user:email:read",
	Description: "Allows viewing all associated emails with the user",
}

var PermissionAddEmails = PermissionDefinition{
	Name:        "user:email:create",
	Description: "Allows adding new emails to the user",
}

var PermissionDeleteEmails = PermissionDefinition{
	Name:        "user:email:delete",
	Description: "Allows deleting emails from the user",
}

var PermissionUpdatePrimaryEmail = PermissionDefinition{
	Name:        "user:email:primary",
	Description: "Allows updating the primary email for the user",
}

var PermissionEnrollMFA = PermissionDefinition{
	Name:        "user:mfa:enroll",
	Description: "Create and enroll MFA methods for the user",
}

var PermissionDeleteMFA = PermissionDefinition{
	Name:        "user:mfa:delete",
	Description: "Remove existing MFA methods for the user",
}
