package crypto

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/tubblelive/star/server/state/claims"
)

const clock_skew = time.Second * 5

type JWT string
type JWTBody struct {
	Duration time.Duration
	Fields map[string]any
}

func CreateJWT(body JWTBody, privateKey *ecdsa.PrivateKey) (JWT, error) {
	signer, err := jose.NewSigner(
		jose.SigningKey{
			Algorithm: jose.ES512, 
			Key: privateKey,
		}, 
		(&jose.SignerOptions{}).WithType("JWT"),
	)

	if err != nil {
		return "", err
	}

	data := map[string]any{
		"iss": "star",
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(body.Duration).Unix(),
		"nbf": time.Now().Add(-clock_skew).Unix(),
		"naf": time.Now().Add(body.Duration + clock_skew).Unix(),
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

func VerifyJWT(jwt string, publicKey *ecdsa.PublicKey) (*claims.UserJWT, error) {
	signature, err := jose.ParseSigned(jwt, []jose.SignatureAlgorithm{jose.ES512})
	if err != nil {
		return nil, err
	}

	data, err := signature.Verify(publicKey)
	if err != nil {
		return nil, err
	}

	var claims claims.UserJWT
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