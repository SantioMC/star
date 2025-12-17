package repository

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/tubblelive/star/oidc"
	"github.com/tubblelive/star/star/config"
	"github.com/tubblelive/star/star/crypto"
	"github.com/tubblelive/star/star/database/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type UserStore struct {
	DB     *gorm.DB
	Config *config.Config
}

func (store *UserStore) CreateSuperUser(password string, role *model.Role) (*model.User, error) {
	user := model.NewUser("admin", "root@star")
	user.EmailVerified = true
	user.Roles = []model.Role{*role}

	account := model.NewAccountWithUserPass(&user, "admin", password)

	if err := store.DB.Create(&account).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func (store *UserStore) CreateUser(username string, email string, password string) (*model.User, error) {
	user := model.NewUser(username, email)
	account := model.NewAccountWithUserPass(&user, username, password)

	if err := store.DB.Create(&account).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func (store *UserStore) GetAllUsers() ([]model.User, error) {
	var users []model.User

	if err := store.DB.Find(&users).Error; err != nil {
		return nil, err
	}

	return users, nil
}

func (store *UserStore) GetUserByID(id string) (*model.User, error) {
	user := model.User{}

	if err := store.DB.Where("id = ?", id).Preload(clause.Associations).First(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func (store *UserStore) GetUserByUsername(username string) (*model.User, error) {
	user := model.User{}

	if err := store.DB.Where("username = ?", username).Preload(clause.Associations).First(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func (store *UserStore) GetUserByEmail(email string) (*model.User, error) {
	user := model.User{}

	if err := store.DB.Where("email = ?", email).Preload(clause.Associations).First(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func (store *UserStore) GetUserByUsernameOrEmail(query string) (*model.User, error) {
	user := model.User{}

	if err := store.DB.Where("email = ? OR username = ?", query, query).Preload(clause.Associations).First(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func (store *UserStore) ValidateLogin(username, password string) (*model.User, error) {
	user, err := store.GetUserByUsernameOrEmail(username)
	if err != nil {
		return nil, err
	}

	var internalAccount *model.Account
	for _, account := range user.Accounts {
		if account.Provider == "internal" {
			internalAccount = &account
			break
		}
	}

	if internalAccount == nil {
		return nil, fmt.Errorf("no internal account found")
	}

	if !crypto.VerifyHash(password, *internalAccount.Password) {
		return nil, fmt.Errorf("invalid password")
	}

	return user, nil
}

func (store *UserStore) CreateAccessToken(
	refreshToken *model.RefreshToken, 
	app *model.Application, 
	claims []oidc.Claim,
	impersonate bool,
) (string, time.Time, error) {
	// todo: a better way of getting the key store
	// todo: impl claims

	expiresAfter := time.Hour * 2
	keyStore := KeyStore{DB: store.DB, Config: store.Config}
	signingKey, err := keyStore.GetSigningKey()
	if err != nil {
		return "", time.Time{}, err
	}

	roles := make([]interface{}, len(refreshToken.User.Roles))
	var groups []string 

	for idx, role := range refreshToken.User.Roles {
		groups = append(groups, role.Name)
		roles[idx] = map[string]any{
			"id": role.ID,
			"name": role.Name,
			"super": role.SuperRole,
		}
	}

	fields := map[string]any{
		"sub": refreshToken.User.ID,
		"aud": app.ID,
		"name": refreshToken.User.Username,
	}

	if slices.Contains(claims, oidc.ClaimEmail) {
		fields["email"] = refreshToken.User.Email
		fields["email_verified"] = refreshToken.User.EmailVerified
	}

	if slices.Contains(claims, oidc.ClaimRoles) {
		fields["roles"] = roles
	}

	if slices.Contains(claims, oidc.ClaimGroups) {
		fields["groups"] = groups
	}

	if impersonate {
		fields["imp"] = true
	}

	privateKey, err := signingKey.GetPrivateKey()
	if err != nil {
		return "", time.Time{}, err
	}

	var jku *string = nil
	if baseUrl := store.Config.Url; baseUrl != "" {
		url := fmt.Sprintf("%s/.well-known/jwks.json", baseUrl)
		jku = &url
	}

	jwt, err := crypto.CreateJWT(crypto.JWTBody{
		Duration: expiresAfter,
		Fields: fields,
	},  jku, privateKey)

	if err != nil {
		return "", time.Time{}, err
	}

	return string(jwt), time.Now().Add(expiresAfter), nil
}

func (store *UserStore) CreateRefreshToken(user *model.User, app *model.Application, claims []oidc.Claim) (*model.RefreshToken, string, time.Time, error) {
	refreshToken := model.NewRefreshToken(time.Hour * 24 * 7, claims)
	refreshToken.User = user
	refreshToken.Application = app

	err := store.DB.Transaction(func(tx *gorm.DB) error {
		gorm.G[model.RefreshToken](store.DB.Unscoped()).Where("user_id = ? AND application_id = ?", user.ID, app.ID).Delete(context.Background())
		
		if err := store.DB.Create(&refreshToken).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, "", time.Time{}, err
	}

	jwt, expiresAt, err := store.CreateAccessToken(&refreshToken, app, claims, false)
	
	if err != nil {
		return nil, "", time.Time{}, err
	}

	return &refreshToken, jwt, expiresAt, nil
}

func (store *UserStore) RefreshAccessToken(refreshToken string) (string, time.Time, error) {
	var token model.RefreshToken
	if err := store.DB.Where("refresh_token = ?", refreshToken).Preload("User").First(&token).Error; err != nil {
		return "", time.Time{}, err
	}

	if token.InvalidatedAt != nil {
		return "", time.Time{}, fmt.Errorf("token is invalidated")
	}

	if token.IsExpired() {
		return "", time.Time{}, fmt.Errorf("token is expired")
	}

	jwt, expiresAt, err := store.CreateAccessToken(&token, token.Application, token.Claims.Data(), false)
	if err != nil {
		return "", time.Time{}, err
	}

	return jwt, expiresAt, nil
}