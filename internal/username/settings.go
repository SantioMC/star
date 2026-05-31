package username

type Settings struct {
	MinLength       int  `json:"min_length"`
	MaxLength       int  `json:"max_length"`
	AllowDash       bool `json:"allow_dash"`
	AllowDots       bool `json:"allow_dots"`
	AllowUnderscore bool `json:"allow_underscore"`
	AllowSpaces     bool `json:"allow_spaces"`
}

var DefaultSettings = Settings{
	MinLength:       3,
	MaxLength:       32,
	AllowDash:       false,
	AllowDots:       true,
	AllowUnderscore: true,
	AllowSpaces:     false,
}
