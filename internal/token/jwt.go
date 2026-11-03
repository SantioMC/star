package token

import (
	"crypto"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"slices"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/tubblelive/star/internal/database"
	"github.com/tubblelive/star/internal/state"
)

const clockSkew = time.Second * 5

const (
	TypeJWT         jose.ContentType = "JWT"
	TypeAccessToken jose.ContentType = "at+jwt"
)

type JWT string
type JWTOptions struct {
	Type      jose.ContentType
	Algorithm database.KeyAlgorithm
	Duration  time.Duration
	IssuedAt  time.Time
	Fields    map[string]any
}

func createJWT(
	state *state.StarServer,
	options JWTOptions,
) (JWT, error) {
	keys := state.SigningKeys.Load()
	if keys == nil {
		return "", ErrNoActiveSigningKey
	}

	algorithm := options.Algorithm
	if algorithm == "" {
		algorithm = database.KeyAlgorithmES512
	}

	index := slices.IndexFunc(*keys, func(key database.SigningKey) bool {
		return key.Algorithm == algorithm && key.IsUsable()
	})

	if index == -1 {
		return "", ErrNoActiveSigningKey
	}

	signingKey := (*keys)[index]

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
			Algorithm: jose.SignatureAlgorithm(signingKey.Algorithm),
			Key:       privateKey,
		},
		(&jose.SignerOptions{
			ExtraHeaders: headers,
		}).WithType(tokenType),
	)

	if err != nil {
		return "", err
	}

	issuedAt := options.IssuedAt
	if issuedAt.IsZero() {
		issuedAt = time.Now()
	}

	data := map[string]any{
		"iss": state.Config.ServerURL,
		"iat": issuedAt.Unix(),
		"exp": time.Now().Add(options.Duration).Unix(),
		"nbf": time.Now().Add(-clockSkew).Unix(),
		"naf": time.Now().Add(options.Duration + clockSkew).Unix(),
	}

	maps.Copy(data, options.Fields)

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

var ErrMalformedToken = errors.New("the access token is malformed")
var ErrMissingKeyID = errors.New("the access token is missing a key id")
var ErrUnknownKey = errors.New("the access token was signed by an unknown key")

func Verify(state *state.StarServer, jwt string) (*AccessTokenJWT, error) {
	signature, err := jose.ParseSigned(jwt, database.SignatureAlgorithms())
	if err != nil || len(signature.Signatures) != 1 {
		return nil, ErrMalformedToken
	}

	header := signature.Signatures[0].Protected
	if header.KeyID == "" {
		return nil, ErrMissingKeyID
	}

	keys := state.SigningKeys.Load()
	if keys == nil {
		return nil, ErrUnknownKey
	}

	index := slices.IndexFunc(*keys, func(key database.SigningKey) bool {
		return key.ID == header.KeyID
	})

	if index == -1 {
		return nil, ErrUnknownKey
	}

	signingKey := (*keys)[index]
	if header.Algorithm != string(signingKey.Algorithm) {
		return nil, ErrMalformedToken
	}

	publicKey, err := signingKey.GetPublicKey()
	if err != nil {
		return nil, err
	}

	return verifyJWT(state, signature, publicKey)
}

func verifyJWT(state *state.StarServer, signature *jose.JSONWebSignature, publicKey crypto.PublicKey) (*AccessTokenJWT, error) {
	if signature.Signatures[0].Protected.ExtraHeaders[jose.HeaderType] != string(TypeAccessToken) {
		return nil, errors.New("token is not an access token")
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

	if state.Revocations.Revoked(claims.IssuedAt, claims.SessionID, claims.Subject, database.GlobalRevocationSubject) {
		return nil, errors.New("token has been revoked")
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
	Permissions string   `json:"urn:star:iam:permissions"`
	Roles       []string `json:"urn:star:iam:roles"`
}
