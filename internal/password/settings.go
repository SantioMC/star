package password

type PasswordSetttings struct {
	MinLength  int `json:"min_length"`
	MaxLength  int `json:"max_length"`
	MinSymbols int `json:"min_symbols"`
	MinUpper   int `json:"min_upper"`
	MinLower   int `json:"min_lower"`
}

var DefaultSettings = PasswordSetttings{
	MinLength:  8,
	MaxLength:  128,
	MinSymbols: 1,
	MinUpper:   1,
	MinLower:   1,
}
