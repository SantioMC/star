package service

import (
	"fmt"
	"slices"
	"strconv"
	"time"

	"github.com/tubblelive/star/oidc"
	"github.com/tubblelive/star/star/crypto"
	"github.com/tubblelive/star/star/database/model"
	"github.com/tubblelive/star/star/state"
	"github.com/tubblelive/star/star/state/scopes"
)

var ErrTokenInvalid = fmt.Errorf("the provided token is not valid")
var ErrTokenExpired = fmt.Errorf("the provided token has expired, please login")
var ErrSuspiciousActivity = fmt.Errorf("suspicious activity was detected, please login")

func CreateSuperUser(password string, role *model.Role) (*model.User, error) {
	user := model.NewUser("admin", "root@star")
	user.EmailVerified = true
	user.Roles = []model.Role{*role}
	
	if err := state.Database.UserStore.SaveUser(user); err != nil {
		return nil, err
	}

	account := model.NewInternalAccount(&user, "admin", password)
	if err := state.Database.UserStore.SaveAccount(&account); err != nil {
		return nil, err
	}

	return &user, nil
}

func GetAllUsers() ([]model.User, error) {
	return state.Database.UserStore.GetAllUsers()
}

func GetUserByID(id string) (*model.User, error) {
	return state.Database.UserStore.GetUserByID(id)
}

func GetUserByUsername(username string) (*model.User, error) {
	return state.Database.UserStore.GetUserByUsername(username)
}

func GetUserByEmail(email string) (*model.User, error) {
	return state.Database.UserStore.GetUserByEmail(email)
}

func GetUserByUsernameOrEmail(query string) (*model.User, error) {
	return state.Database.UserStore.GetUserByUsernameOrEmail(query)
}

func GetActiveSuspension(userId string) (*model.Suspension, error) {
	return state.Database.UserStore.GetActiveSuspension(userId)
}

func SaveUser(user model.User) error {
	return state.Database.UserStore.SaveUser(user)
}

func FindOrCreateUser(username string, email string) (model.User, error) {
	user, err := GetUserByEmail(email)
	if err != nil {
		return model.User{}, err
	}

	if user == nil {
		newUser := model.NewUser(username, email)

		if err := state.Database.UserStore.SaveUser(newUser); err != nil {
			return model.User{}, err
		}

		user = &newUser
	}

	return *user, nil
}

func CreateUser(username string, email string, password string) (*model.User, *model.Account, error) {
	user, err := FindOrCreateUser(username, email)
	if err != nil {
		return nil, nil, err
	}

	account := model.NewInternalAccount(&user, username, password)
	if err := state.Database.UserStore.SaveAccount(&account); err != nil {
		return &user, nil, err
	}
	
	return &user, &account, nil
}

func FindUsername(username string, iteration int) (string, error) {
	if iteration >= 10 {
		// Give up, give a random username and don't bother pursuing this further
		return model.CreateId(16), nil
	}

	check := username
	if iteration > 1 {
		check += strconv.Itoa(iteration - 1)
	}

	existing, err := GetUserByUsername(check)
	if err != nil {
		return "", err
	}

	if existing != nil {
		return FindUsername(check, iteration + 1)
	}

	return check, nil
}

func CreateUserFromSocial(provider *model.SocialProvider, token string) (*model.User, *model.Account, error) {
	userInfo, err := provider.GetUserInfo(token)
	if err != nil {
		return nil, nil, err
	}

	if userInfo.Email == "" {
		return nil, nil, fmt.Errorf("cannot create account without valid email")
	}

	if userInfo.Username == "" {
		return nil, nil, fmt.Errorf("cannot create account without valid username")
	}

	username, err := FindUsername(userInfo.Username, 1)
	if err != nil {
		return nil, nil, err
	}

	user, err := FindOrCreateUser(username, userInfo.Email)
	if err != nil {
		return nil, nil, err
	}

	newAccount := model.NewProviderAccount(&user, provider, userInfo)
	if err := state.Database.UserStore.SaveAccount(&newAccount); err != nil {
		return &user, nil, err
	}

	return &user, &newAccount, nil
}

func ValidateLogin(username, password string) (*model.User, error) {
	user, err := GetUserByUsernameOrEmail(username)
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

func GetUserPermissions(user *model.User) ([]model.Permission, error) {
	return state.Database.UserStore.GetUserPermissions(user)
}

func CreateAccessToken(
	refreshToken *model.Session, 
	app *model.Application, 
	claims []oidc.Claim,
	impersonate bool,
) (string, time.Time, error) {
	expiresAfter := time.Hour * 2
	signingKey, err := GetSigningKey()
	if err != nil {
		return "", time.Time{}, err
	}

	roles := make([]interface{}, len(refreshToken.User.Roles))

	var groups []string 
	var perms []scopes.Scope 

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
		// Give ourselves frontend access to user-facing endpoints
		perms = append(perms, scopes.ScopeUserAll)
	}

	permissions, err := GetUserPermissions(refreshToken.User)
	if err != nil {
		return "", time.Time{}, err
	}

	for _, perm := range permissions {
		perms = append(perms, scopes.Scope(perm.Permission))
	}

	fields["scopes"] = perms

	privateKey, err := signingKey.GetPrivateKey()
	if err != nil {
		return "", time.Time{}, err
	}

	var jku *string = nil
	if baseUrl := state.StarState.Config.Url; baseUrl != "" {
		url := baseUrl + "/.well-known/jwks.json"
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

func CreateSession(
	user *model.User,
	app *model.Application,
	fingerprint *string,
	userAgent string,
	claims []oidc.Claim,
) (*model.Session, string, time.Time, error) {
	session := model.NewSession(time.Hour * 24 * 7, claims)
	session.User = user
	session.Application = app
	session.UserAgent = userAgent
	session.Fingerprint = fingerprint

	if err := state.Database.SessionStore.CreateSession(&session); err != nil {
		return nil, "", time.Time{}, err
	}

	jwt, expiresAt, err := CreateAccessToken(&session, app, claims, false)
	if err != nil {
		return nil, "", time.Time{}, err
	}

	return &session, jwt, expiresAt, nil
}

func RefreshAccessToken(
	userId string, 
	refreshToken string,
	fingerprint *string,
	userAgent string,
) (string, time.Time, error) {
	sessions, err := state.Database.SessionStore.GetUserSessions(userId)
	if err != nil {
		return "", time.Time{}, err
	}

	var session *model.Session = nil
	for _, s := range sessions {
		if s.Token == refreshToken {
			session = &s
		}
	}

	if session == nil {
		return "", time.Time{}, ErrTokenInvalid
	}

	if session.InvalidatedAt != nil {
		return "", time.Time{}, ErrTokenInvalid
	}

	if session.IsExpired() {
		return "", time.Time{}, ErrTokenExpired
	}

	if session.Fingerprint != nil && session.Fingerprint != fingerprint {
		// If the fingerprint has changed, we invalidate the session
		state.Database.SessionStore.InvalidateSession(session.ID)
		return "", time.Time{}, ErrSuspiciousActivity
	} else if session.Fingerprint == nil && fingerprint != nil {
		// If we did not have a fingerprint associated but were provided one now, we'll update the session to 
		session.Fingerprint = fingerprint
		state.Database.SessionStore.UpdateSessionFingerprint(session.ID, fingerprint)
	}

	jwt, expiresAt, err := CreateAccessToken(session, session.Application, session.Claims.Data(), false)
	if err != nil {
		return "", time.Time{}, err
	}

	return jwt, expiresAt, nil
}