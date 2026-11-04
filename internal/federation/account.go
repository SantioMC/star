package federation

import (
	"context"
	"errors"
	"math/rand/v2"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/tubblelive/star/internal/database"
	"github.com/tubblelive/star/internal/email"
	"github.com/tubblelive/star/internal/state"
	"github.com/tubblelive/star/internal/username"
)

var (
	ErrRegistrationDisabled = fiber.NewError(403, "Registration is currently disabled, sign in with an existing account instead")
	ErrEmailInUse           = fiber.NewError(409, "An account already uses this email address, sign in with your password instead")
	ErrAccountConflict      = fiber.NewError(409, "This account was created by another request at the same time, please try signing in again")
)

func ResolveAccount(
	ctx context.Context,
	star *state.StarServer,
	provider *database.Provider,
	identity *Identity,
) (*database.User, *database.Account, bool, error) {
	account := star.Database.GetProviderAccount(ctx, provider.ID, identity.Subject)
	if account == nil {
		user, linked, err := createUser(ctx, star, provider, identity)
		return user, linked, err == nil, err
	}

	account.Name = displayName(identity)
	account.Email = identity.Email
	account.Avatar = identity.Picture

	if err := star.Database.UpdateAccount(ctx, account); err != nil {
		return nil, nil, false, err
	}

	user := star.Database.GetUserByID(ctx, account.UserID)
	if user == nil {
		return nil, nil, false, errors.New("linked provider account has no user")
	}

	return user, account, false, nil
}

func createUser(
	ctx context.Context,
	star *state.StarServer,
	provider *database.Provider,
	identity *Identity,
) (*database.User, *database.Account, error) {
	if !star.Settings.Load().AllowRegistration {
		return nil, nil, ErrRegistrationDisabled
	}

	verified := identity.Email != nil && identity.EmailVerified && provider.TrustEmail
	if verified && len(star.Database.GetUsersByVerifiedEmail(ctx, *identity.Email)) > 0 {
		return nil, nil, ErrEmailInUse
	}

	name, err := availableUsername(ctx, star, identity)
	if err != nil {
		return nil, nil, err
	}

	created := time.Now()
	user := &database.User{
		ID:             database.CreateID(database.UserKey),
		Name:           name,
		FirstName:      identity.GivenName,
		LastName:       identity.FamilyName,
		ProfilePicture: identity.Picture,
		CreatedAt:      created,
	}

	var primary *database.UserEmail
	if verified {
		primary = &database.UserEmail{
			ID:        database.CreateID(database.EmailKey),
			UserID:    user.ID,
			Email:     *identity.Email,
			Verified:  true,
			Primary:   true,
			Code:      email.CreateCode(star.Settings.Load().EmailCodeLength),
			CreatedAt: created,
		}
	}

	account := &database.Account{
		ID:         database.CreateID(database.AccountKey),
		Type:       "provider",
		UserID:     user.ID,
		ProviderID: &provider.ID,
		ExternalID: &identity.Subject,
		Name:       displayName(identity),
		Email:      identity.Email,
		Avatar:     identity.Picture,
		CreatedAt:  created,
	}

	err = star.Database.CreateUserIdentity(ctx, user, primary, account)
	if errors.Is(err, database.ErrDuplicate) {
		return nil, nil, ErrAccountConflict
	} else if err != nil {
		return nil, nil, err
	}

	return user, account, nil
}

func displayName(identity *Identity) *string {
	if identity.Username != nil {
		return identity.Username
	}

	return identity.Name
}

func availableUsername(ctx context.Context, star *state.StarServer, identity *Identity) (string, error) {
	candidates := []string{}

	for _, value := range []*string{identity.Username, identity.Email} {
		if value != nil {
			local, _, _ := strings.Cut(*value, "@")
			candidates = append(candidates, local)
		}
	}

	candidates = append(candidates, "user")

	for _, candidate := range candidates {
		base, err := username.ValidateUsername(candidate, &username.DefaultSettings)
		if err != nil {
			continue
		}

		attempt := base
		for range 5 {
			valid, err := username.ValidateUsername(attempt, &username.DefaultSettings)
			if err == nil && len(star.Database.GetUsersByName(ctx, valid)) == 0 {
				return valid, nil
			}

			attempt = base + strconv.Itoa(rand.IntN(100000))
		}
	}

	return "", errors.New("unable to find an available username for the provider account")
}
