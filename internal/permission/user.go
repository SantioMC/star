package permission

var PermissionChangePassword = PermissionDefinition{
	Name:        "user:password:write",
	Description: "Allows changing or removing the user's password",
	Default:     true,
}

var PermissionResetPassword = PermissionDefinition{
	Name:        "user:password:reset",
	Description: "Allows resetting the user's password through a link sent to a verified email",
	Default:     true,
}

var PermissionCreateAuthorization = PermissionDefinition{
	Name:        "user:application:authorize",
	Description: "Allows the user to authorize an application",
	Default:     true,
}

var PermissionRevokeAuthorization = PermissionDefinition{
	Name:        "user:application:revoke",
	Description: "Allows the user to revoke an application's access to their account",
	Default:     true,
}

var PermissionModifySelf = PermissionDefinition{
	Name:        "user:write",
	Description: "Allows the user to modify their own information",
	Default:     true,
}

var PermissionReadEmails = PermissionDefinition{
	Name:        "user:email:read",
	Description: "Allows viewing all associated emails with the user",
	Default:     true,
}

var PermissionAddEmails = PermissionDefinition{
	Name:        "user:email:create",
	Description: "Allows adding new emails to the user",
	Default:     true,
}

var PermissionDeleteEmails = PermissionDefinition{
	Name:        "user:email:delete",
	Description: "Allows deleting emails from the user",
	Default:     true,
}

var PermissionUpdatePrimaryEmail = PermissionDefinition{
	Name:        "user:email:primary",
	Description: "Allows updating the primary email for the user",
	Default:     true,
}

var PermissionManageConnections = PermissionDefinition{
	Name:        "user:connection:write",
	Description: "Allows connecting and removing social providers on the user",
	Default:     true,
}

var PermissionEnrollMFA = PermissionDefinition{
	Name:        "user:mfa:enroll",
	Description: "Create and enroll MFA methods for the user",
	Default:     true,
}

var PermissionDeleteMFA = PermissionDefinition{
	Name:        "user:mfa:delete",
	Description: "Remove existing MFA methods for the user",
	Default:     true,
}

var PermissionBypassNoLogin = PermissionDefinition{
	Name:        "user:login:bypass",
	Description: "Allows still logging in when logins are disabled",
}

var PermissionExportSelf = PermissionDefinition{
	Name:        "user:data:export",
	Description: "Allows the user to export all of their personal data",
	Default:     true,
}
