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

const (
	TypeJWT         jose.ContentType = "JWT"
	TypeAccessToken jose.ContentType = "at+jwt"
)

type JWT string
type JWTOptions struct {
	Type     jose.ContentType
	Duration time.Duration
	Fields   map[string]any
}

func createJWT(
	state *state.StarServer,
	options JWTOptions,
) (JWT, error) {
	var signingKey *database.SigningKey = nil
	for _, key := range *state.SigningKeys {
		if key.Status == database.SigningKeyStatusActive {
			signingKey = &key
			break
		}
	}

	if signingKey == nil {
		return "", ErrNoActiveSigningKey
	}

	tokenType := options.Type
	if tokenType == "" {
		tokenType = TypeJWT
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
		}).WithType(tokenType),
	)

	if err != nil {
		return "", err
	}

	data := map[string]any{
		"iss": state.Config.ServerURL,
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(options.Duration).Unix(),
		"nbf": time.Now().Add(-clockSkew).Unix(),
		"naf": time.Now().Add(options.Duration + clockSkew).Unix(),
	}

	for key, value := range options.Fields {
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

func VerifyJWT(state *state.StarServer, jwt string, publicKey *ecdsa.PublicKey) (*AccessTokenJWT, error) {
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
		return nil, errors.New("token is expired")
	}

	if claims.NotBefore > time.Now().Unix() {
		return nil, errors.New("token is not yet valid")
	}

	if claims.Issuer != state.Config.ServerURL {
		return nil, errors.New("token is not issued by the correct server")
	}

	if len(claims.Audience) == 0 {
		return nil, errors.New("token has no audience")
	}

	return &claims, nil
}

type Audience []string

func (audience *Audience) UnmarshalJSON(data []byte) error {
	var single string
	if err := json.Unmarshal(data, &single); err == nil {
		*audience = Audience{single}
		return nil
	}

	var multiple []string
	if err := json.Unmarshal(data, &multiple); err != nil {
		return err
	}

	*audience = Audience(multiple)
	return nil
}

type AccessTokenJWT struct {
	ExpiresAt   int64    `json:"exp"`
	NotBefore   int64    `json:"nbf"`
	IssuedAt    int64    `json:"iat"`
	TokenID     string   `json:"jti"`
	Subject     string   `json:"sub"`
	Audience    Audience `json:"aud"`
	Issuer      string   `json:"iss"`
	ClientID    string   `json:"client_id"`
	SessionID   string   `json:"sid"`
	Scope       string   `json:"scope"`
	Permissions []string `json:"permissions"`
}
