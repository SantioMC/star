package oauth

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"
	"unicode"

	"github.com/gofiber/fiber/v3"
	"github.com/lib/pq"
	"github.com/tubblelive/star/internal/cryptography"
	"github.com/tubblelive/star/internal/database"
	"github.com/tubblelive/star/internal/openid"
	"github.com/tubblelive/star/internal/server/common"
	"github.com/tubblelive/star/internal/session"
	"github.com/tubblelive/star/internal/state"
	"github.com/tubblelive/star/internal/useragent"
)

const DeviceCodeGrantType = "urn:ietf:params:oauth:grant-type:device_code"

const userCodeAlphabet = "BCDFGHJKLMNPQRSTVWXZ"

type deviceAuthorizationResponse struct {
	DeviceCode              string `json:"device_code"`
	UserCode                string `json:"user_code"`
	VerificationURI         string `json:"verification_uri"`
	VerificationURIComplete string `json:"verification_uri_complete"`
	ExpiresIn               int64  `json:"expires_in"`
	Interval                int    `json:"interval"`
}

func NormalizeUserCode(value string) string {
	return strings.Map(func(character rune) rune {
		upper := unicode.ToUpper(character)
		if strings.ContainsRune(userCodeAlphabet, upper) {
			return upper
		}

		return -1
	}, value)
}

func FormatUserCode(value string) string {
	if len(value) != 8 {
		return value
	}

	return value[:4] + "-" + value[4:]
}

func devicePollCode(ctx fiber.Ctx) string {
	body := ctx.Request().Body()
	if len(body) > 4096 {
		return ""
	}

	form, err := url.ParseQuery(string(body))
	if err != nil || form.Get("grant_type") != DeviceCodeGrantType {
		return ""
	}

	return form.Get("device_code")
}

func IsDevicePoll(ctx fiber.Ctx) bool {
	return devicePollCode(ctx) != ""
}

func DevicePollAddress(ctx fiber.Ctx) string {
	if !IsDevicePoll(ctx) {
		return ""
	}

	return common.GetState(ctx).ClientIP(ctx)
}

func DevicePollKey(ctx fiber.Ctx) string {
	deviceCode := devicePollCode(ctx)
	if deviceCode == "" {
		return ""
	}

	return cryptography.HMAC(deviceCode, common.GetState(ctx).Config.AppToken)
}

func SlowDown(ctx fiber.Ctx) error {
	return writeTokenError(ctx, 400, Oauth2ErrorSlowDown, "Polling too frequently, increase the polling interval by 5 seconds")
}

func DeviceAuthorization(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	form, ok, err := parseForm(ctx)
	if !ok {
		return err
	}

	application, ok, err := authenticateClient(ctx, form)
	if !ok {
		return err
	}

	if application == nil {
		return writeTokenError(ctx, 400, Oauth2ErrorInvalidRequest, "Missing client_id, clients must identify themselves")
	}

	if !application.DeviceFlow {
		return writeTokenError(ctx, 400, Oauth2ErrorUnauthorizedClient, "This client is not permitted to use the device authorization grant")
	}

	scopes := strings.Fields(form.Get("scope"))
	if len(scopes) == 0 {
		scopes = []string{openid.ScopeOpenID}
	}

	resolution := openid.ResolveScopes(ctx, star.Database, application, scopes)
	if len(resolution.Granted) == 0 {
		return writeTokenError(ctx, 400, Oauth2ErrorInvalidScope, "None of the requested scopes are supported")
	}

	authorization, err := createDeviceAuthorization(ctx, star, application, resolution.Names())
	if err != nil {
		return err
	}

	verificationURI := star.Config.ServerURL + "/device"
	userCode := FormatUserCode(authorization.UserCode)

	return writeTokenResponse(ctx, deviceAuthorizationResponse{
		DeviceCode:              authorization.DeviceCode,
		UserCode:                userCode,
		VerificationURI:         verificationURI,
		VerificationURIComplete: fmt.Sprintf("%s?user_code=%s", verificationURI, url.QueryEscape(userCode)),
		ExpiresIn:               int64(time.Until(authorization.ExpiresAt).Round(time.Second).Seconds()),
		Interval:                authorization.PollInterval,
	})
}

func createDeviceAuthorization(ctx fiber.Ctx, star *state.StarServer, application *database.Application, scopes []string) (*database.DeviceAuthorization, error) {
	geolocation := session.Locate(ctx, star, star.ClientIP(ctx))
	now := time.Now()

	for range 5 {
		authorization := &database.DeviceAuthorization{
			ID:            database.CreateID(database.DeviceAuthorizationKey),
			ApplicationID: application.ID,
			DeviceCode:    cryptography.CreateSecureString(64),
			UserCode:      cryptography.CreateSecureStringFrom(userCodeAlphabet, 8),
			Scopes:        pq.StringArray(scopes),
			Status:        database.DeviceAuthorizationPending,
			Device:        string(useragent.GetDevice(ctx.UserAgent())),
			Geolocation:   geolocation,
			PollInterval:  5,
			LastPolledAt:  now,
			CreatedAt:     now,
			ExpiresAt:     now.Add(time.Minute * 10),
		}

		err := star.Database.CreateDeviceAuthorization(ctx, authorization)
		if errors.Is(err, database.ErrDuplicate) {
			continue
		} else if err != nil {
			return nil, err
		}

		return authorization, nil
	}

	return nil, errors.New("failed to generate a unique user code")
}

func deviceCodeGrant(ctx fiber.Ctx, form url.Values, application *database.Application) error {
	star := common.GetState(ctx)

	if application == nil {
		return writeTokenError(ctx, 400, Oauth2ErrorInvalidRequest, "Missing client_id, clients must identify themselves")
	}

	if !application.DeviceFlow {
		return writeTokenError(ctx, 400, Oauth2ErrorUnauthorizedClient, "This client is not permitted to use the device authorization grant")
	}

	deviceCode := form.Get("device_code")
	if deviceCode == "" {
		return writeTokenError(ctx, 400, Oauth2ErrorInvalidRequest, "Missing device_code")
	}

	authorization := star.Database.GetDeviceAuthorizationByDeviceCode(ctx, deviceCode)
	if authorization == nil {
		return writeTokenError(ctx, 400, Oauth2ErrorInvalidGrant, "Unknown or expired device code")
	}

	if authorization.ApplicationID != application.ID {
		return writeTokenError(ctx, 400, Oauth2ErrorInvalidGrant, "This device code was not issued to this client")
	}

	if authorization.RedeemedAt != nil {
		return revokeReplayedDeviceCode(ctx, star, authorization)
	}

	if authorization.ExpiresAt.Before(time.Now()) {
		return writeTokenError(ctx, 400, Oauth2ErrorExpiredToken, "This device code has expired")
	}

	if authorization.Status == database.DeviceAuthorizationDenied {
		return writeTokenError(ctx, 400, Oauth2ErrorAccessDenied, "The authorization request was denied")
	}

	paced, err := star.Database.PollDeviceAuthorization(ctx, authorization.ID)
	if err != nil {
		return err
	} else if !paced {
		return SlowDown(ctx)
	}

	if authorization.Status != database.DeviceAuthorizationApproved {
		return writeTokenError(ctx, 400, Oauth2ErrorAuthorizationPending, "The user has not completed the authorization request yet")
	}

	return redeemGrant(ctx, application, &issuedGrant{
		label:           "device code",
		userID:          *authorization.UserID,
		scopes:          authorization.Scopes,
		authenticatedAt: *authorization.AuthenticatedAt,
		redeem: func(familyID string) (bool, error) {
			return star.Database.RedeemDeviceAuthorization(ctx, authorization.ID, familyID)
		},
		replayed: func() error {
			return revokeReplayedDeviceCode(ctx, star, star.Database.GetDeviceAuthorizationByID(ctx, authorization.ID))
		},
	})
}

func revokeReplayedDeviceCode(ctx fiber.Ctx, star *state.StarServer, authorization *database.DeviceAuthorization) error {
	if authorization == nil || authorization.UserID == nil {
		return writeTokenError(ctx, 400, Oauth2ErrorInvalidGrant, "This device code has already been redeemed")
	}

	return revokeReplayedGrant(ctx, star, &grantReplay{
		kind:          database.SecurityAlertDeviceCodeUsed,
		label:         "device code",
		message:       "A device code was redeemed more than once, the session issued from it has been revoked",
		id:            authorization.ID,
		userID:        *authorization.UserID,
		applicationID: authorization.ApplicationID,
		familyID:      authorization.FamilyID,
	})
}
