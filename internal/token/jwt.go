package token

import (
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/tubblelive/star/internal/database"
	"github.com/tubblelive/star/internal/state"
)

const (
	clockSkew    = time.Second * 5
	JwtAlgorithm = jose.ES512
)

type JWT string
type JWTBody struct {
	Duration time.Duration
	Fields   map[string]any
}

func createJWT(
	state *state.StarServer,
	body JWTBody,
) (JWT, error) {
	var signingKey *database.SigningKey = nil
	for _, key := range *state.SigningKeys {
		if key.Status == database.SigningKeyStatusActive {
			signingKey = &key
			break
		}
	}

	if signingKey == nil {
		return "", errors.New("no active signing key found")
	}

	headers := map[jose.HeaderKey]any{
		"kid": signingKey.ID,
		"jku": fmt.Sprintf("%s/.well-known/jwks.json", state.Config.ServerURL),
	}

	privateKey, err := signingKey.GetPrivateKey()
	if err != nil {
		return "", err
	}

	signer, err := jose.NewSigner(
		jose.SigningKey{
			Algorithm: JwtAlgorithm,
			Key:       privateKey,
		},
		(&jose.SignerOptions{
			ExtraHeaders: headers,
		}).WithType("JWT"),
	)

	if err != nil {
		return "", err
	}

	// todo: change iss to be uri
	data := map[string]any{
		"iss": state.Config.ServerURL,
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(body.Duration).Unix(),
		"nbf": time.Now().Add(-clockSkew).Unix(),
		"naf": time.Now().Add(body.Duration + clockSkew).Unix(),
	}

	for key, value := range body.Fields {
		data[key] = value
	}

	payload, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	signature, err := signer.Sign(payload)
	if err != nil {
		return "", err
	}

	sanitized, err := signature.CompactSerialize()
	if err != nil {
		return "", err
	}

	return JWT(sanitized), nil
}

func VerifyJWT(jwt string, publicKey *ecdsa.PublicKey) (*AccessTokenJWT, error) {
	signature, err := jose.ParseSigned(jwt, []jose.SignatureAlgorithm{jose.ES512})
	if err != nil {
		return nil, err
	}

	data, err := signature.Verify(publicKey)
	if err != nil {
		return nil, err
	}

	var claims AccessTokenJWT
	if err := json.Unmarshal(data, &claims); err != nil {
		return nil, err
	}

	if claims.ExpiresAt < time.Now().Unix() {
		return nil, fmt.Errorf("token is expired")
	}

	if claims.NotBefore > time.Now().Unix() {
		return nil, fmt.Errorf("token is not yet valid")
	}

	return &claims, nil
}

// todo: include all jwt claims
type AccessTokenJWT struct {
	ExpiresAt   int64    `json:"exp"`
	NotBefore   int64    `json:"nbf"`
	IssuedAt    int64    `json:"iat"`
	Subject     string   `json:"sub"`
	Permissions []string `json:"permissions"`
}
