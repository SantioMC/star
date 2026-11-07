package federation

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"net/http"
	"strings"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
	"github.com/tubblelive/star/internal/birthday"
	"github.com/tubblelive/star/internal/database"
	"github.com/tubblelive/star/internal/database/models"
)

type Identity struct {
	Subject       string
	Email         *string
	EmailVerified bool
	Username      *string
	Name          *string
	GivenName     *string
	FamilyName    *string
	Picture       *string
	Birthday      *models.Date
}

func FetchIdentity(
	ctx context.Context,
	endpoints *Endpoints,
	provider *database.Provider,
	tokens *Tokens,
	nonce string,
) (*Identity, error) {
	claims := map[string]any{}

	if tokens.IDToken != "" && endpoints.JWKSURL != "" {
		idClaims, err := parseIDToken(ctx, tokens.IDToken, endpoints, provider, nonce)
		if err != nil {
			return nil, err
		}

		maps.Copy(claims, idClaims)
	}

	if endpoints.UserinfoURL != "" {
		userinfo, err := fetchUserinfo(ctx, endpoints.UserinfoURL, tokens.AccessToken)
		if err != nil {
			return nil, err
		}

		idSubject := claimString(claims, "sub")
		userinfoSubject := claimString(userinfo, "sub")

		if idSubject != nil && userinfoSubject != nil && *idSubject != *userinfoSubject {
			return nil, errors.New("userinfo subject does not match the id token subject")
		}

		maps.Copy(claims, userinfo)
	}

	subject := claimString(claims, "sub", "id")
	if subject == nil {
		return nil, errors.New("provider did not return a subject identifier")
	}

	picture := claimString(claims, "picture", "avatar_url")
	if picture != nil && !strings.HasPrefix(*picture, "https://") {
		picture = nil
	}

	claimedBirthday, err := birthday.Parse(claimString(claims, "birthdate"))
	if err != nil {
		claimedBirthday = nil
	}

	return &Identity{
		Subject:       *subject,
		Email:         claimString(claims, "email"),
		EmailVerified: claimBool(claims, "email_verified", "verified"),
		Username:      claimString(claims, "preferred_username", "login", "username", "nickname"),
		Name:          claimString(claims, "name", "global_name"),
		GivenName:     claimString(claims, "given_name", "first_name"),
		FamilyName:    claimString(claims, "family_name", "last_name"),
		Picture:       picture,
		Birthday:      claimedBirthday,
	}, nil
}

func fetchUserinfo(ctx context.Context, userinfoURL string, accessToken string) (map[string]any, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, userinfoURL, nil)
	if err != nil {
		return nil, err
	}

	request.Header.Set("Authorization", "Bearer "+accessToken)

	userinfo := map[string]any{}
	if err := doJSON(request, &userinfo); err != nil {
		return nil, fmt.Errorf("failed to fetch userinfo: %w", err)
	}

	return userinfo, nil
}

func parseIDToken(ctx context.Context, raw string, endpoints *Endpoints, provider *database.Provider, nonce string) (map[string]any, error) {
	token, err := jwt.ParseSigned(raw, []jose.SignatureAlgorithm{
		jose.RS256, jose.RS384, jose.RS512,
		jose.PS256, jose.PS384, jose.PS512,
		jose.ES256, jose.ES384, jose.ES512,
		jose.EdDSA,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse id token: %w", err)
	}

	var registered jwt.Claims
	claims := map[string]any{}

	if err := verifyIDToken(ctx, token, endpoints, &registered, &claims); err != nil {
		return nil, err
	}

	tenant, _ := claims["tid"].(string)
	issuer := strings.ReplaceAll(endpoints.Issuer, "{tenantid}", tenant)

	expected := jwt.Expected{
		Issuer:      issuer,
		AnyAudience: jwt.Audience{provider.ClientID},
		Time:        time.Now(),
	}

	if err := registered.ValidateWithLeeway(expected, time.Minute); err != nil {
		return nil, fmt.Errorf("id token failed validation: %w", err)
	}

	authorizedParty, _ := claims["azp"].(string)
	if authorizedParty != "" && authorizedParty != provider.ClientID {
		return nil, errors.New("id token was issued to a different authorized party")
	}

	if len(registered.Audience) > 1 && authorizedParty == "" {
		return nil, errors.New("id token has multiple audiences but no authorized party")
	}

	tokenNonce, _ := claims["nonce"].(string)
	if subtle.ConstantTimeCompare([]byte(tokenNonce), []byte(nonce)) != 1 {
		return nil, errors.New("id token nonce does not match the authorization request")
	}

	return claims, nil
}

func verifyIDToken(
	ctx context.Context,
	token *jwt.JSONWebToken,
	endpoints *Endpoints,
	registered *jwt.Claims,
	claims *map[string]any,
) error {
	keySet, err := fetchKeySet(ctx, endpoints.JWKSURL, time.Hour)
	if err != nil {
		return err
	}

	if err := token.Claims(keySet, registered, claims); err == nil {
		return nil
	}

	keySet, err = fetchKeySet(ctx, endpoints.JWKSURL, time.Minute)
	if err != nil {
		return err
	}

	if err := token.Claims(keySet, registered, claims); err != nil {
		return fmt.Errorf("id token signature could not be verified: %w", err)
	}

	return nil
}

func claimString(claims map[string]any, keys ...string) *string {
	for _, key := range keys {
		var value string

		switch typed := claims[key].(type) {
		case string:
			value = strings.TrimSpace(typed)
		case json.Number:
			value = typed.String()
		case float64:
			value = fmt.Sprintf("%.0f", typed)
		}

		if value != "" && len(value) <= 255 {
			return &value
		}
	}

	return nil
}

func claimBool(claims map[string]any, keys ...string) bool {
	for _, key := range keys {
		switch typed := claims[key].(type) {
		case bool:
			return typed
		case string:
			return typed == "true"
		}
	}

	return false
}
