package repository

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/tubblelive/star/oidc"
	"github.com/tubblelive/star/star/config"
	"github.com/tubblelive/star/star/crypto"
	"github.com/tubblelive/star/star/database/model"
	"github.com/tubblelive/star/star/state/scopes"
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

	account := model.NewInternalAccount(&user, "admin", password)

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
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		
		return nil, err
	}

	return &user, nil
}

func (store *UserStore) GetUserByUsername(username string) (*model.User, error) {
	user := model.User{}

	if err := store.DB.Where("username = ?", username).Preload(clause.Associations).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		
		return nil, err
	}

	return &user, nil
}

func (store *UserStore) GetUserByEmail(email string) (*model.User, error) {
	user := model.User{}

	if err := store.DB.Where("email = ?", strings.ToLower(email)).Preload(clause.Associations).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		
		return nil, err
	}

	return &user, nil
}

func (store *UserStore) GetUserByUsernameOrEmail(query string) (*model.User, error) {
	user := model.User{}

	if err := store.DB.Where("email = ? OR username = ?", query, query).Preload(clause.Associations).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return &user, nil
}

func (store *UserStore) GetActiveSuspension(id string) (*model.Suspension, error) {
	suspension := model.Suspension{}

	if err := store.DB.Where("user_id = ?", id).First(&suspension).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return &suspension, nil
}

func (store *UserStore) SaveUser(user model.User) error {
	_, err := gorm.G[model.User](store.DB).Where("id = ?", user.ID).Updates(context.Background(), user)
	return err
}

func (store *UserStore) CreateUser(username string, email string, password string) (*model.User, *model.Account, error) {
	var user model.User
	if err := store.DB.Where("email = ?", email).Attrs(model.NewUser(username, email)).FirstOrCreate(&user).Error; err != nil {
		return nil, nil, err
	}

	newAccount := model.NewInternalAccount(&user, username, password)

	var account model.Account
	if err := store.DB.Where("user_id = ? AND provider_id IS NULL", user.ID).Attrs(newAccount).FirstOrCreate(&account).Error; err != nil {
		return &user, nil, err
	}
	
	return &user, &account, nil
}

func (store *UserStore) CreateUserFromSocial(provider *model.SocialProvider, token string) (*model.User, *model.Account, error) {
	userInfo, err := provider.GetUserInfo(token)
	if err != nil {
		return nil, nil, err
	}

	if userInfo.Email == "" {
		return nil, nil, fmt.Errorf("cannot create account without valid email")
	}

	var user model.User
	if err := store.DB.Where("email = ?", userInfo.Email).Attrs(model.NewUserFromInfo(userInfo)).FirstOrCreate(&user).Error; err != nil {
		return nil, nil, err
	}

	newAccount := model.NewProviderAccount(&user, provider, userInfo)

	var account model.Account
	if err := store.DB.
		Where("user_id = ? AND provider_id = ?", user.ID, provider.ID).
		Assign(model.Account{ PlatformAvatar: userInfo.ProfilePicture, Username: userInfo.Username }).
		Attrs(newAccount).
		FirstOrCreate(&account).Error; 
		err != nil {	
			return &user, nil, err
		}

	return &user, &newAccount, nil
}

func (store *UserStore) ValidateLogin(username, password string) (*model.User, error) {
	user, err := store.GetUserByUsernameOrEmail(username)
	if err != nil || user == nil {
		return nil, fmt.Errorf("invalid user")
	}

	var internalAccount *model.Account
	for _, account := range user.Accounts {
		if account.Provider == nil {
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

func (store *UserStore) GetUserPermissions(user *model.User) ([]model.Permission, error) {
	var roleIds []string
	for _, role := range user.Roles {
		roleIds = append(roleIds, role.ID)
	}

	return gorm.G[model.Permission](store.DB).
		Where(
			"application_id IN (?)", 
			store.DB.Model(model.Permission{}).
				Select("application_id").
				Where("default = ? OR id IN ?", true, roleIds),
		).Find(context.Background())
}

func (store *UserStore) CreateAccessToken(
	refreshToken *model.Session, 
	app *model.Application, 
	claims []oidc.Claim,
	impersonate bool,
) (string, time.Time, error) {
	// todo: a better way of getting the stores
	keyStore := KeyStore{DB: store.DB, Config: store.Config}

	expiresAfter := time.Hour * 2
	signingKey, err := keyStore.GetSigningKey()
	if err != nil {
		return "", time.Time{}, err
	}

	roles := make([]interface{}, len(refreshToken.User.Roles))
	isSuperUser := false

	var groups []string 
	var perms []scopes.Scope 

	for idx, role := range refreshToken.User.Roles {
		groups = append(groups, role.Name)
		if role.SuperRole {
			isSuperUser = true
		}

		roles[idx] = map[string]any{
			"id": role.ID,
			"name": role.Name,
			"super": role.SuperRole,
		}
	}

	fields := map[string]any{
		"sub": refreshToken.User.ID,
		"aud": app.ID,
	}

	if slices.Contains(claims, oidc.ClaimEmail) {
		fields["email"] = refreshToken.User.Email
		fields["email_verified"] = refreshToken.User.EmailVerified
		perms = append(perms, scopes.ScopeReadEmail) // Allow apps to read email
	}

	if slices.Contains(claims, oidc.ClaimRoles) {
		fields["roles"] = roles
	}

	if slices.Contains(claims, oidc.ClaimGroups) {
		fields["groups"] = groups
	}

	if slices.Contains(claims, oidc.ClaimPicture) {
		fields["picture"] = refreshToken.User.GenerateProfilePicture()
	}

	if slices.Contains(claims, oidc.ClaimProfile) {
		fields["name"] = refreshToken.User.Username
		fields["nickname"] = refreshToken.User.Username
		fields["given_name"] = refreshToken.User.FirstName
		fields["family_name"] = refreshToken.User.LastName
		fields["picture"] = refreshToken.User.GenerateProfilePicture()
	}

	if impersonate {
		fields["imp"] = true
	}

	if app.Internal {
		// If we're authenticating a management access token, we'll add scopes for it
		if isSuperUser {
			// Give token access admin endpoints
			perms = append(perms, scopes.ScopeAll)
		} else {
			// Give token access to user endpoints
			perms = append(perms, scopes.ScopeUserAll)
		}
	} else {
		permissions, err := store.GetUserPermissions(refreshToken.User)
		if err != nil {
			return "", time.Time{}, err
		}

		for _, perm := range permissions {
			perms = append(perms, scopes.Scope(perm.Permission))
		}
	}

	fields["scopes"] = perms

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

func (store *UserStore) CreateRefreshToken(user *model.User, app *model.Application, claims []oidc.Claim) (*model.Session, string, time.Time, error) {
	refreshToken := model.NewSession(time.Hour * 24 * 7, claims)
	refreshToken.User = user
	refreshToken.Application = app

	err := store.DB.Transaction(func(tx *gorm.DB) error {
		gorm.G[model.Session](store.DB.Unscoped()).Where("user_id = ? AND application_id = ?", user.ID, app.ID).Delete(context.Background())
		
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
	var token model.Session
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