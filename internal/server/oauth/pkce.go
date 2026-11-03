package oauth

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"regexp"

	"github.com/tubblelive/star/internal/database"
)

const CodeChallengeMethodS256 = "S256"

var pkcePattern = regexp.MustCompile(`^[A-Za-z0-9\-._~]{43,128}$`)

func ValidateCodeChallenge(application *database.Application, challenge string, method string) (Oauth2Error, string, bool) {
	if challenge == "" {
		if application.Kind == database.ApplicationKindPublic {
			return Oauth2ErrorInvalidRequest, "Public clients must provide a code_challenge", false
		}

		return "", "", true
	}

	if method != CodeChallengeMethodS256 {
		return Oauth2ErrorInvalidRequest, "Transform algorithm not supported, only S256 is supported", false
	}

	if !pkcePattern.MatchString(challenge) {
		return Oauth2ErrorInvalidRequest, "The code_challenge must be 43-128 characters of [A-Za-z0-9-._~]", false
	}

	return "", "", true
}

func ValidCodeVerifierFormat(verifier string) bool {
	return pkcePattern.MatchString(verifier)
}

func CreateCodeChallenge(verifier string) string {
	hash := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(hash[:])
}

func VerifyCodeVerifier(verifier string, challenge string) bool {
	computed := CreateCodeChallenge(verifier)
	return subtle.ConstantTimeCompare([]byte(computed), []byte(challenge)) == 1
}
