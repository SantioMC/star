package password

type Algorithm string

const (
	AlgorithmArgon2Id Algorithm = "argon2"
	AlgorithmBcrypt   Algorithm = "bcrypt"
)

var Algorithms = []Algorithm{AlgorithmArgon2Id, AlgorithmBcrypt}

type PasswordSetttings struct {
	MinLength  int       `json:"min_length"`
	MaxLength  int       `json:"max_length"`
	MinSymbols int       `json:"min_symbols"`
	MinUpper   int       `json:"min_upper"`
	MinLower   int       `json:"min_lower"`
	Algorithm  Algorithm `json:"algorithm"`
}

var DefaultSettings = PasswordSetttings{
	MinLength:  8,
	MaxLength:  128,
	MinSymbols: 1,
	MinUpper:   1,
	MinLower:   1,
	Algorithm:  AlgorithmArgon2Id,
}
