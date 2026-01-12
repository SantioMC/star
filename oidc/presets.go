package oidc

/*
The oauth2 package that golang provides is kind of a mess, so we're not going to use it
*/

// The basic information required to use a social provider
type Configuration struct {
	AuthorizationURL string   `json:"auth_url"`
	TokenURL         string   `json:"token_url"`
	UserInfoURL      string   `json:"userinfo_url"`
	DefaultScopes    []string `json:"scopes"`

	// The tabler icon representing this config, only really used for the UI
	Icon string `json:"icon"`
}

// The provider information for Discord
var DiscordProvider = Configuration{
	AuthorizationURL: "https://discord.com/oauth2/authorize",
	TokenURL:         "https://discord.com/api/oauth2/token",
	UserInfoURL:      "https://discord.com/api/oauth2/userinfo",
	DefaultScopes:    []string{"identify", "email"},
	Icon:             "brand-discord",
}

// The provider information for Github
var GithubProvider = Configuration{
	AuthorizationURL: "https://github.com/login/oauth/authorize",
	TokenURL:         "https://github.com/login/oauth/access_token",
	UserInfoURL:      "https://api.github.com/user",
	DefaultScopes:    []string{"read:user", "user:email"},
	Icon:             "brand-github",
}

// The provider information for Slack
var SlackProvider = Configuration{
	AuthorizationURL: "https://slack.com/oauth/v2/authorize",
	TokenURL:         "https://slack.com/api/oauth.v2.access",
	UserInfoURL:      "https://slack.com/api/openid.connect.userInfo",
	DefaultScopes:    []string{"openid", "profile", "email"},
	Icon:             "brand-slack",
}

// The provider information for Twitch
var TwitchProvider = Configuration{
	AuthorizationURL: "https://id.twitch.tv/oauth2/authorize",
	TokenURL:         "https://id.twitch.tv/oauth2/token",
	UserInfoURL:      "https://id.twitch.tv/oauth2/userinfo",
	DefaultScopes:    []string{"openid", "user:read:email"},
	Icon:             "brand-twitch",
}

// The provider information for Google
var GoogleProvider = Configuration{
	AuthorizationURL: "https://accounts.google.com/o/oauth2/auth",
	TokenURL:         "https://oauth2.googleapis.com/token",
	UserInfoURL:      "https://openidconnect.googleapis.com/v1/userinfo",
	DefaultScopes:    []string{"openid", "https://www.googleapis.com/auth/userinfo.profile", "https://www.googleapis.com/auth/userinfo.email"},
	Icon:             "brand-google-filled",
}

// A list of well-known providers with their configurations
var Providers = map[string]Configuration{
	"discord": DiscordProvider,
	"github":  GithubProvider,
	"slack":   SlackProvider,
	"twitch":  TwitchProvider,
	"google":  GoogleProvider,
}
