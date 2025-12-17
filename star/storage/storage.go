package storage

// var (
// 	_ op.Storage                  = &Storage{}
// 	_ op.ClientCredentialsStorage = &Storage{}
// )

// type Storage struct {
// 	lock     sync.Mutex
// 	database *database.Database
// }

// func (s *Storage) AuthRequestByCode(_ context.Context, code string) (op.AuthRequest, error) {
// 	request, err := s.database.RequestStore.GetAuthRequestByCode(code)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return request, nil
// }

// func (s *Storage) AuthRequestByID(_ context.Context, id string) (op.AuthRequest, error) {
// 	request, err := s.database.RequestStore.GetAuthRequestByID(id)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return request, nil
// }

// func (s *Storage) AuthorizeClientIDSecret(ctx context.Context, clientID string, clientSecret string) error {
// 	client, err := s.database.ClientStore.GetClientByID(clientID)
// 	if err != nil {
// 		return err
// 	}

// 	if client.Secret != clientSecret {
// 		return fmt.Errorf("invalid secret")
// 	}

// 	return nil
// }

// func (s *Storage) CreateAccessAndRefreshTokens(ctx context.Context, request op.TokenRequest, currentRefreshToken string) (accessTokenID string, newRefreshToken string, expiration time.Time, err error) {
// 	if currentRefreshToken == "" {
// 		user, err := s.database.UserStore.GetUserByID(request.GetSubject())
// 		if err != nil {
// 			return "", "", time.Time{}, err
// 		}

// 		token, expiresAt, jwt, err := s.database.UserStore.CreateRefreshToken(user)
// 		if err != nil {
// 			return "", "", time.Time{}, err
// 		}

// 		return jwt, token.Token, expiresAt, nil
// 	}

// 	jwt, expiresAt, err := s.database.UserStore.RefreshAccessToken(currentRefreshToken)
// 	if err != nil {
// 		return "", "", time.Time{}, err
// 	}

// 	return jwt, currentRefreshToken, expiresAt, nil
// }

// // CreateAccessToken implements op.Storage.
// func (s *Storage) CreateAccessToken(context.Context, op.TokenRequest) (accessTokenID string, expiration time.Time, err error) {
// 	panic("unimplemented")
// }

// // CreateAuthRequest implements op.Storage.
// func (s *Storage) CreateAuthRequest(context.Context, *oidc.AuthRequest, string) (op.AuthRequest, error) {
// 	panic("unimplemented")
// }

// // DeleteAuthRequest implements op.Storage.
// func (s *Storage) DeleteAuthRequest(context.Context, string) error {
// 	panic("unimplemented")
// }

// // GetClientByClientID implements op.Storage.
// func (s *Storage) GetClientByClientID(ctx context.Context, clientID string) (op.Client, error) {
// 	panic("unimplemented")
// }

// // GetKeyByIDAndClientID implements op.Storage.
// func (s *Storage) GetKeyByIDAndClientID(ctx context.Context, keyID string, clientID string) (*jose.JSONWebKey, error) {
// 	panic("unimplemented")
// }

// // GetPrivateClaimsFromScopes implements op.Storage.
// func (s *Storage) GetPrivateClaimsFromScopes(ctx context.Context, userID string, clientID string, scopes []string) (map[string]any, error) {
// 	panic("unimplemented")
// }

// // GetRefreshTokenInfo implements op.Storage.
// func (s *Storage) GetRefreshTokenInfo(ctx context.Context, clientID string, token string) (userID string, tokenID string, err error) {
// 	panic("unimplemented")
// }

// // Health implements op.Storage.
// func (s *Storage) Health(context.Context) error {
// 	return nil
// }

// // KeySet implements op.Storage.
// func (s *Storage) KeySet(context.Context) ([]op.Key, error) {
// 	panic("unimplemented")
// }

// // RevokeToken implements op.Storage.
// func (s *Storage) RevokeToken(ctx context.Context, tokenOrTokenID string, userID string, clientID string) *oidc.Error {
// 	panic("unimplemented")
// }

// // SaveAuthCode implements op.Storage.
// func (s *Storage) SaveAuthCode(context.Context, string, string) error {
// 	panic("unimplemented")
// }

// // SetIntrospectionFromToken implements op.Storage.
// func (s *Storage) SetIntrospectionFromToken(ctx context.Context, userinfo *oidc.IntrospectionResponse, tokenID string, subject string, clientID string) error {
// 	panic("unimplemented")
// }

// // SetUserinfoFromScopes implements op.Storage.
// func (s *Storage) SetUserinfoFromScopes(ctx context.Context, userinfo *oidc.UserInfo, userID string, clientID string, scopes []string) error {
// 	panic("unimplemented")
// }

// // SetUserinfoFromToken implements op.Storage.
// func (s *Storage) SetUserinfoFromToken(ctx context.Context, userinfo *oidc.UserInfo, tokenID string, subject string, origin string) error {
// 	panic("unimplemented")
// }

// // SignatureAlgorithms implements op.Storage.
// func (s *Storage) SignatureAlgorithms(context.Context) ([]jose.SignatureAlgorithm, error) {
// 	panic("unimplemented")
// }

// // SigningKey implements op.Storage.
// func (s *Storage) SigningKey(context.Context) (op.SigningKey, error) {
// 	panic("unimplemented")
// }

// // TerminateSession implements op.Storage.
// func (s *Storage) TerminateSession(ctx context.Context, userID string, clientID string) error {
// 	panic("unimplemented")
// }

// // TokenRequestByRefreshToken implements op.Storage.
// func (s *Storage) TokenRequestByRefreshToken(ctx context.Context, refreshToken string) (op.RefreshTokenRequest, error) {
// 	panic("unimplemented")
// }

// // ValidateJWTProfileScopes implements op.Storage.
// func (s *Storage) ValidateJWTProfileScopes(ctx context.Context, userID string, scopes []string) ([]string, error) {
// 	panic("unimplemented")
// }

// func (s *Storage) CheckUsernamePassword(username, password, id string) error {
// 	s.lock.Lock()
// 	defer s.lock.Unlock()

// 	request, ok := s.authRequests[id]
// 	if !ok {
// 		return fmt.Errorf("request not found")
// 	}

// 	// todo: hash check
// 	user := s.userStore.GetUserByUsername(username)
// 	if user != nil && user.Password == password {
// 		request.UserID = user.ID
// 		request.done = true
// 		request.authTime = time.Now()

// 		return nil
// 	}

// 	return fmt.Errorf("username or password wrong")
// }

// func (s *Storage) CheckUsernamePasswordSimple(username, password string) error {
// 	s.lock.Lock()
// 	defer s.lock.Unlock()

// 	// todo: hash check
// 	user := s.userStore.GetUserByUsername(username)
// 	if user != nil && user.Password == password {
// 		return nil
// 	}

// 	return fmt.Errorf("username or password wrong")
// }

// func (s *Storage) CreateAuthRequest(ctx context.Context, authReq *oidc.AuthRequest, userID string) (op.AuthRequest, error) {
// 	s.lock.Lock()
// 	defer s.lock.Unlock()

// 	if len(authReq.Prompt) == 1 && authReq.Prompt[0] == "none" {
// 		return nil, oidc.ErrLoginRequired()
// 	}

// 	request := authRequestToInternal(authReq, userID)

// 	request.ID = database.CreateId(12)
// 	s.authRequests[request.ID] = request

// 	return request, nil
// }

// func (s *Storage) AuthRequestByID(ctx context.Context, id string) (op.AuthRequest, error) {
// 	s.lock.Lock()
// 	defer s.lock.Unlock()

// 	request, ok := s.authRequests[id]

// 	if !ok {
// 		return nil, fmt.Errorf("request not found")
// 	}

// 	return request, nil
// }

// func (s *Storage) AuthRequestByCode(ctx context.Context, code string) (op.AuthRequest, error) {
// 	requestID, ok := func() (string, bool) {
// 		s.lock.Lock()
// 		defer s.lock.Unlock()
// 		requestID, ok := s.codes[code]
// 		return requestID, ok
// 	}()

// 	if !ok {
// 		return nil, fmt.Errorf("code invalid or expired")
// 	}

// 	return s.AuthRequestByID(ctx, requestID)
// }

// func (s *Storage) SaveAuthCode(ctx context.Context, id string, code string) error {
// 	s.lock.Lock()
// 	defer s.lock.Unlock()

// 	s.codes[code] = id

// 	return nil
// }

// func (s *Storage) DeleteAuthRequest(ctx context.Context, id string) error {
// 	s.lock.Lock()
// 	defer s.lock.Unlock()

// 	delete(s.authRequests, id)

// 	for code, requestID := range s.codes {
// 		if id == requestID {
// 			delete(s.codes, code)
// 			return nil
// 		}
// 	}

// 	return nil
// }

// func (s *Storage) CreateAccessToken(ctx context.Context, request op.TokenRequest) (string, time.Time, error) {
// 	var applicationID string
// 	switch req := request.(type) {
// 	case *AuthRequest:
// 		applicationID = req.ApplicationID
// 	case op.TokenExchangeRequest:
// 		applicationID = req.GetClientID()
// 	}

// 	token, err := s.accessToken(applicationID, "", request.GetSubject(), request.GetAudience(), request.GetScopes())
// 	if err != nil {
// 		return "", time.Time{}, err
// 	}

// 	return token.ID, token.Expiration, nil
// }

// func (s *Storage) CreateAccessAndRefreshTokens(ctx context.Context, request op.TokenRequest, currentRefreshToken string) (accessTokenID string, newRefreshToken string, expiration time.Time, err error) {
// 	if teReq, ok := request.(op.TokenExchangeRequest); ok {
// 		return s.exchangeRefreshToken(ctx, teReq)
// 	}

// 	applicationID, authTime, amr := getInfoFromRequest(request)

// 	if currentRefreshToken == "" {
// 		refreshTokenID := uuid.NewString()
// 		accessToken, err := s.accessToken(applicationID, refreshTokenID, request.GetSubject(), request.GetAudience(), request.GetScopes())
// 		if err != nil {
// 			return "", "", time.Time{}, err
// 		}
// 		refreshToken, err := s.createRefreshToken(accessToken, amr, authTime)
// 		if err != nil {
// 			return "", "", time.Time{}, err
// 		}
// 		return accessToken.ID, refreshToken, accessToken.Expiration, nil
// 	}

// 	// todo: change to jwt
// 	newRefreshToken = database.CreateId(24)

// 	accessToken, err := s.accessToken(applicationID, newRefreshToken, request.GetSubject(), request.GetAudience(), request.GetScopes())
// 	if err != nil {
// 		return "", "", time.Time{}, err
// 	}

// 	if err := s.renewRefreshToken(currentRefreshToken, newRefreshToken, accessToken.ID); err != nil {
// 		return "", "", time.Time{}, err
// 	}

// 	return accessToken.ID, newRefreshToken, accessToken.Expiration, nil
// }

// func (s *Storage) exchangeRefreshToken(ctx context.Context, request op.TokenExchangeRequest) (accessTokenID string, newRefreshToken string, expiration time.Time, err error) {
// 	applicationID := request.GetClientID()
// 	authTime := request.GetAuthTime()

// 	refreshTokenID := uuid.NewString()
// 	accessToken, err := s.accessToken(applicationID, refreshTokenID, request.GetSubject(), request.GetAudience(), request.GetScopes())
// 	if err != nil {
// 		return "", "", time.Time{}, err
// 	}

// 	refreshToken, err := s.createRefreshToken(accessToken, nil, authTime)
// 	if err != nil {
// 		return "", "", time.Time{}, err
// 	}

// 	return accessToken.ID, refreshToken, accessToken.Expiration, nil
// }

// func (s *Storage) TokenRequestByRefreshToken(ctx context.Context, refreshToken string) (op.RefreshTokenRequest, error) {
// 	s.lock.Lock()
// 	defer s.lock.Unlock()
// 	token, ok := s.refreshTokens[refreshToken]

// 	if !ok {
// 		return nil, fmt.Errorf("invalid refresh_token")
// 	}

// 	return RefreshTokenRequestFromBusiness(token), nil
// }

// func (s *Storage) TerminateSession(ctx context.Context, userID string, clientID string) error {
// 	s.lock.Lock()
// 	defer s.lock.Unlock()

// 	for _, token := range s.tokens {
// 		if token.ApplicationID == clientID && token.Subject == userID {
// 			delete(s.tokens, token.ID)
// 			delete(s.refreshTokens, token.RefreshTokenID)
// 		}
// 	}

// 	return nil
// }

// // GetRefreshTokenInfo looks up a refresh token and returns the token id and user id.
// // If given something that is not a refresh token, it must return error.
// func (s *Storage) GetRefreshTokenInfo(ctx context.Context, clientID string, token string) (userID string, tokenID string, err error) {
// 	refreshToken, ok := s.refreshTokens[token]

// 	if !ok {
// 		return "", "", op.ErrInvalidRefreshToken
// 	}

// 	return refreshToken.UserID, refreshToken.ID, nil
// }

// func (s *Storage) RevokeToken(ctx context.Context, tokenIDOrToken string, userID string, clientID string) *oidc.Error {
// 	s.lock.Lock()
// 	defer s.lock.Unlock()

// 	accessToken, ok := s.tokens[tokenIDOrToken] // tokenID

// 	if ok {
// 		if accessToken.ApplicationID != clientID {
// 			return oidc.ErrInvalidClient().WithDescription("token was not issued for this client")
// 		}

// 		delete(s.tokens, accessToken.ID)
// 		return nil
// 	}

// 	refreshToken, ok := s.refreshTokens[tokenIDOrToken] // token
// 	if !ok {
// 		return nil
// 	}

// 	if refreshToken.ApplicationID != clientID {
// 		return oidc.ErrInvalidClient().WithDescription("token was not issued for this client")
// 	}

// 	delete(s.refreshTokens, refreshToken.ID)
// 	delete(s.tokens, refreshToken.AccessToken)

// 	return nil
// }

// func (s *Storage) SigningKey(ctx context.Context) (op.SigningKey, error) {
// 	return &s.signingKey, nil
// }

// func (s *Storage) SignatureAlgorithms(context.Context) ([]jose.SignatureAlgorithm, error) {
// 	return []jose.SignatureAlgorithm{s.signingKey.SignatureAlgorithm()}, nil
// }

// func (s *Storage) KeySet(ctx context.Context) ([]op.Key, error) {
// 	return []op.Key{&model.PublicKey{SigningKey: s.signingKey}}, nil
// }

// func (s *Storage) GetClientByClientID(ctx context.Context, clientID string) (op.Client, error) {
// 	s.lock.Lock()
// 	defer s.lock.Unlock()

// 	client, ok := s.clients[clientID]

// 	if !ok {
// 		return nil, fmt.Errorf("client not found")
// 	}

// 	return RedirectGlobsClient(client), nil
// }

// func (s *Storage) AuthorizeClientIDSecret(ctx context.Context, clientID, clientSecret string) error {
// 	s.lock.Lock()
// 	defer s.lock.Unlock()

// 	client, ok := s.clients[clientID]
// 	if !ok {
// 		return fmt.Errorf("client not found")
// 	}

// 	if client.secret != clientSecret {
// 		return fmt.Errorf("invalid secret")
// 	}

// 	return nil
// }

// func (s *Storage) SetUserinfoFromScopes(ctx context.Context, userinfo *oidc.UserInfo, userID, clientID string, scopes []string) error {
// 	return nil
// }

// func (s *Storage) SetUserinfoFromRequest(ctx context.Context, userinfo *oidc.UserInfo, token op.IDTokenRequest, scopes []string) error {
// 	return s.setUserinfo(ctx, userinfo, token.GetSubject(), token.GetClientID(), scopes)
// }

// func (s *Storage) SetUserinfoFromToken(ctx context.Context, userinfo *oidc.UserInfo, tokenID, subject, origin string) error {
// 	token, ok := func() (*Token, bool) {
// 		s.lock.Lock()
// 		defer s.lock.Unlock()
// 		token, ok := s.tokens[tokenID]
// 		return token, ok
// 	}()

// 	if !ok {
// 		return fmt.Errorf("token is invalid or has expired")
// 	}

// 	if token.Expiration.Before(time.Now()) {
// 		return fmt.Errorf("token is expired")
// 	}

// 	return s.setUserinfo(ctx, userinfo, token.Subject, token.ApplicationID, token.Scopes)
// }

// func (s *Storage) SetIntrospectionFromToken(ctx context.Context, introspection *oidc.IntrospectionResponse, tokenID, subject, clientID string) error {
// 	token, ok := func() (*Token, bool) {
// 		s.lock.Lock()
// 		defer s.lock.Unlock()
// 		token, ok := s.tokens[tokenID]
// 		return token, ok
// 	}()
// 	if !ok {
// 		return fmt.Errorf("token is invalid or has expired")
// 	}

// 	for _, aud := range token.Audience {
// 		if aud == clientID {
// 			userInfo := new(oidc.UserInfo)
// 			err := s.setUserinfo(ctx, userInfo, subject, clientID, token.Scopes)
// 			if err != nil {
// 				return err
// 			}

// 			introspection.SetUserInfo(userInfo)
// 			introspection.Scope = token.Scopes
// 			introspection.ClientID = token.ApplicationID

// 			return nil
// 		}
// 	}
// 	return fmt.Errorf("token is not valid for this client")
// }

// func (s *Storage) GetPrivateClaimsFromScopes(ctx context.Context, userID, clientID string, scopes []string) (claims map[string]any, err error) {
// 	return s.getPrivateClaimsFromScopes(ctx, userID, clientID, scopes)
// }

// func (s *Storage) getPrivateClaimsFromScopes(ctx context.Context, userID, clientID string, scopes []string) (claims map[string]any, err error) {
// 	// for _, scope := range scopes {
// 	// 	switch scope {
// 	// 	case CustomScope:
// 	// 		claims = appendClaim(claims, CustomClaim, customClaim(clientID))
// 	// 	}
// 	// }
// 	return claims, nil
// }

// // GetKeyByIDAndClientID implements the op.Storage interface
// // it will be called to validate the signatures of a JWT (JWT Profile Grant and Authentication)
// func (s *Storage) GetKeyByIDAndClientID(ctx context.Context, keyID, clientID string) (*jose.JSONWebKey, error) {
// 	s.lock.Lock()
// 	defer s.lock.Unlock()

// 	service, ok := s.Services[clientID]
// 	if !ok {
// 		return nil, fmt.Errorf("clientID not found")
// 	}

// 	key, ok := service.keys[keyID]
// 	if !ok {
// 		return nil, fmt.Errorf("key not found")
// 	}

// 	return &jose.JSONWebKey{
// 		KeyID: keyID,
// 		Use:   "sig",
// 		Key:   key,
// 	}, nil
// }

// // ValidateJWTProfileScopes implements the op.Storage interface
// // it will be called to validate the scopes of a JWT Profile Authorization Grant request
// func (s *Storage) ValidateJWTProfileScopes(ctx context.Context, userID string, scopes []string) ([]string, error) {
// 	allowedScopes := make([]string, 0)
// 	for _, scope := range scopes {
// 		if scope == oidc.ScopeOpenID {
// 			allowedScopes = append(allowedScopes, scope)
// 		}
// 	}
// 	return allowedScopes, nil
// }

// // Health implements the op.Storage interface
// func (s *Storage) Health(ctx context.Context) error {
// 	return nil
// }

// // createRefreshToken will store a refresh_token in-memory based on the provided information
// func (s *Storage) createRefreshToken(accessToken *Token, amr []string, authTime time.Time) (string, error) {
// 	s.lock.Lock()
// 	defer s.lock.Unlock()
// 	token := &RefreshToken{
// 		ID:            accessToken.RefreshTokenID,
// 		Token:         accessToken.RefreshTokenID,
// 		AuthTime:      authTime,
// 		AMR:           amr,
// 		ApplicationID: accessToken.ApplicationID,
// 		UserID:        accessToken.Subject,
// 		Audience:      accessToken.Audience,
// 		Expiration:    time.Now().Add(5 * time.Hour),
// 		Scopes:        accessToken.Scopes,
// 		AccessToken:   accessToken.ID,
// 	}
// 	s.refreshTokens[token.ID] = token
// 	return token.Token, nil
// }

// // renewRefreshToken checks the provided refresh_token and creates a new one based on the current
// //
// // [Refresh Token Rotation] is implemented.
// //
// // [Refresh Token Rotation]: https://www.rfc-editor.org/rfc/rfc6819#section-5.2.2.3
// func (s *Storage) renewRefreshToken(currentRefreshToken, newRefreshToken, newAccessToken string) error {
// 	s.lock.Lock()
// 	defer s.lock.Unlock()
// 	refreshToken, ok := s.refreshTokens[currentRefreshToken]
// 	if !ok {
// 		return fmt.Errorf("invalid refresh token")
// 	}
// 	// deletes the refresh token
// 	delete(s.refreshTokens, currentRefreshToken)

// 	// delete the access token which was issued based on this refresh token
// 	delete(s.tokens, refreshToken.AccessToken)

// 	if refreshToken.Expiration.Before(time.Now()) {
// 		return fmt.Errorf("expired refresh token")
// 	}

// 	// creates a new refresh token based on the current one
// 	refreshToken.Token = newRefreshToken
// 	refreshToken.ID = newRefreshToken
// 	refreshToken.Expiration = time.Now().Add(5 * time.Hour)
// 	refreshToken.AccessToken = newAccessToken
// 	s.refreshTokens[newRefreshToken] = refreshToken
// 	return nil
// }

// // accessToken will store an access_token in-memory based on the provided information
// func (s *Storage) accessToken(applicationID, refreshTokenID, subject string, audience, scopes []string) (*Token, error) {
// 	s.lock.Lock()
// 	defer s.lock.Unlock()
// 	token := &Token{
// 		ID:             uuid.NewString(),
// 		ApplicationID:  applicationID,
// 		RefreshTokenID: refreshTokenID,
// 		Subject:        subject,
// 		Audience:       audience,
// 		Expiration:     time.Now().Add(5 * time.Minute),
// 		Scopes:         scopes,
// 	}
// 	s.tokens[token.ID] = token
// 	return token, nil
// }

// // setUserinfo sets the info based on the user, scopes and if necessary the clientID
// func (s *Storage) setUserinfo(ctx context.Context, userInfo *oidc.UserInfo, userID, clientID string, scopes []string) (err error) {
// 	s.lock.Lock()
// 	defer s.lock.Unlock()
// 	user := s.userStore.GetUserByID(userID)
// 	if user == nil {
// 		return fmt.Errorf("user not found")
// 	}
// 	for _, scope := range scopes {
// 		switch scope {
// 		case oidc.ScopeOpenID:
// 			userInfo.Subject = user.ID
// 		case oidc.ScopeEmail:
// 			userInfo.Email = user.Email
// 			userInfo.EmailVerified = oidc.Bool(user.EmailVerified)
// 		case oidc.ScopeProfile:
// 			userInfo.PreferredUsername = user.Username
// 			userInfo.Name = user.FirstName + " " + user.LastName
// 			userInfo.FamilyName = user.LastName
// 			userInfo.GivenName = user.FirstName
// 			userInfo.Locale = oidc.NewLocale(user.PreferredLanguage)
// 		case oidc.ScopePhone:
// 			userInfo.PhoneNumber = user.Phone
// 			userInfo.PhoneNumberVerified = user.PhoneVerified
// 		case CustomScope:
// 			// you can also have a custom scope and assert public or custom claims based on that
// 			userInfo.AppendClaims(CustomClaim, customClaim(clientID))
// 		}
// 	}
// 	return nil
// }

// // ValidateTokenExchangeRequest implements the op.TokenExchangeStorage interface
// // it will be called to validate parsed Token Exchange Grant request
// func (s *Storage) ValidateTokenExchangeRequest(ctx context.Context, request op.TokenExchangeRequest) error {
// 	if request.GetRequestedTokenType() == "" {
// 		request.SetRequestedTokenType(oidc.RefreshTokenType)
// 	}

// 	// Just an example, some use cases might need this use case
// 	if request.GetExchangeSubjectTokenType() == oidc.IDTokenType && request.GetRequestedTokenType() == oidc.RefreshTokenType {
// 		return errors.New("exchanging id_token to refresh_token is not supported")
// 	}

// 	// Check impersonation permissions
// 	if request.GetExchangeActor() == "" && !s.userStore.GetUserByID(request.GetExchangeSubject()).IsAdmin {
// 		return errors.New("user doesn't have impersonation permission")
// 	}

// 	allowedScopes := make([]string, 0)
// 	for _, scope := range request.GetScopes() {
// 		if scope == oidc.ScopeAddress {
// 			continue
// 		}

// 		if strings.HasPrefix(scope, CustomScopeImpersonatePrefix) {
// 			subject := strings.TrimPrefix(scope, CustomScopeImpersonatePrefix)
// 			request.SetSubject(subject)
// 		}

// 		allowedScopes = append(allowedScopes, scope)
// 	}

// 	request.SetCurrentScopes(allowedScopes)

// 	return nil
// }

// // ValidateTokenExchangeRequest implements the op.TokenExchangeStorage interface
// // Common use case is to store request for audit purposes. For this example we skip the storing.
// func (s *Storage) CreateTokenExchangeRequest(ctx context.Context, request op.TokenExchangeRequest) error {
// 	return nil
// }

// // GetPrivateClaimsFromScopesForTokenExchange implements the op.TokenExchangeStorage interface
// // it will be called for the creation of an exchanged JWT access token to assert claims for custom scopes
// // plus adding token exchange specific claims related to delegation or impersonation
// func (s *Storage) GetPrivateClaimsFromTokenExchangeRequest(ctx context.Context, request op.TokenExchangeRequest) (claims map[string]any, err error) {
// 	claims, err = s.getPrivateClaimsFromScopes(ctx, "", request.GetClientID(), request.GetScopes())
// 	if err != nil {
// 		return nil, err
// 	}

// 	for k, v := range s.getTokenExchangeClaims(ctx, request) {
// 		claims = appendClaim(claims, k, v)
// 	}

// 	return claims, nil
// }

// // SetUserinfoFromScopesForTokenExchange implements the op.TokenExchangeStorage interface
// // it will be called for the creation of an id_token - we are using the same private function as for other flows,
// // plus adding token exchange specific claims related to delegation or impersonation
// func (s *Storage) SetUserinfoFromTokenExchangeRequest(ctx context.Context, userinfo *oidc.UserInfo, request op.TokenExchangeRequest) error {
// 	err := s.setUserinfo(ctx, userinfo, request.GetSubject(), request.GetClientID(), request.GetScopes())
// 	if err != nil {
// 		return err
// 	}

// 	for k, v := range s.getTokenExchangeClaims(ctx, request) {
// 		userinfo.AppendClaims(k, v)
// 	}

// 	return nil
// }

// func (s *Storage) getTokenExchangeClaims(ctx context.Context, request op.TokenExchangeRequest) (claims map[string]any) {
// 	for _, scope := range request.GetScopes() {
// 		switch {
// 		case strings.HasPrefix(scope, CustomScopeImpersonatePrefix) && request.GetExchangeActor() == "":
// 			// Set actor subject claim for impersonation flow
// 			claims = appendClaim(claims, "act", map[string]any{
// 				"sub": request.GetExchangeSubject(),
// 			})
// 		}
// 	}

// 	// Set actor subject claim for delegation flow
// 	// if request.GetExchangeActor() != "" {
// 	// 	claims = appendClaim(claims, "act", map[string]any{
// 	// 		"sub": request.GetExchangeActor(),
// 	// 	})
// 	// }

// 	return claims
// }

// // getInfoFromRequest returns the clientID, authTime and amr depending on the op.TokenRequest type / implementation
// func getInfoFromRequest(req op.TokenRequest) (clientID string, authTime time.Time, amr []string) {
// 	authReq, ok := req.(*AuthRequest) // Code Flow (with scope offline_access)
// 	if ok {
// 		return authReq.ApplicationID, authReq.authTime, authReq.GetAMR()
// 	}
// 	refreshReq, ok := req.(*RefreshTokenRequest) // Refresh Token Request
// 	if ok {
// 		return refreshReq.ApplicationID, refreshReq.AuthTime, refreshReq.AMR
// 	}
// 	return "", time.Time{}, nil
// }

// // customClaim demonstrates how to return custom claims based on provided information
// func customClaim(clientID string) map[string]any {
// 	return map[string]any{
// 		"client": clientID,
// 		"other":  "stuff",
// 	}
// }

// func appendClaim(claims map[string]any, claim string, value any) map[string]any {
// 	if claims == nil {
// 		claims = make(map[string]any)
// 	}
// 	claims[claim] = value
// 	return claims
// }

// type deviceAuthorizationEntry struct {
// 	deviceCode string
// 	userCode   string
// 	state      *op.DeviceAuthorizationState
// }

// func (s *Storage) StoreDeviceAuthorization(ctx context.Context, clientID, deviceCode, userCode string, expires time.Time, scopes []string) error {
// 	s.lock.Lock()
// 	defer s.lock.Unlock()

// 	if _, ok := s.clients[clientID]; !ok {
// 		return errors.New("client not found")
// 	}

// 	if _, ok := s.userCodes[userCode]; ok {
// 		return op.ErrDuplicateUserCode
// 	}

// 	s.deviceCodes[deviceCode] = deviceAuthorizationEntry{
// 		deviceCode: deviceCode,
// 		userCode:   userCode,
// 		state: &op.DeviceAuthorizationState{
// 			ClientID: clientID,
// 			Scopes:   scopes,
// 			Expires:  expires,
// 		},
// 	}

// 	s.userCodes[userCode] = deviceCode
// 	return nil
// }

// func (s *Storage) GetDeviceAuthorizatonState(ctx context.Context, clientID, deviceCode string) (*op.DeviceAuthorizationState, error) {
// 	if ctx.Err() != nil {
// 		return nil, ctx.Err()
// 	}

// 	s.lock.Lock()
// 	defer s.lock.Unlock()

// 	entry, ok := s.deviceCodes[deviceCode]
// 	if !ok || entry.state.ClientID != clientID {
// 		return nil, errors.New("device code not found for client") // is there a standard not found error in the framework?
// 	}

// 	return entry.state, nil
// }

// func (s *Storage) GetDeviceAuthorizationByUserCode(ctx context.Context, userCode string) (*op.DeviceAuthorizationState, error) {
// 	s.lock.Lock()
// 	defer s.lock.Unlock()

// 	entry, ok := s.deviceCodes[s.userCodes[userCode]]
// 	if !ok {
// 		return nil, errors.New("user code not found")
// 	}

// 	return entry.state, nil
// }

// func (s *Storage) CompleteDeviceAuthorization(ctx context.Context, userCode, subject string) error {
// 	s.lock.Lock()
// 	defer s.lock.Unlock()

// 	entry, ok := s.deviceCodes[s.userCodes[userCode]]
// 	if !ok {
// 		return errors.New("user code not found")
// 	}

// 	entry.state.Subject = subject
// 	entry.state.Done = true
// 	return nil
// }

// func (s *Storage) DenyDeviceAuthorization(ctx context.Context, userCode string) error {
// 	s.lock.Lock()
// 	defer s.lock.Unlock()

// 	s.deviceCodes[s.userCodes[userCode]].state.Denied = true
// 	return nil
// }

// // AuthRequestDone is used by testing and is not required to implement op.Storage
// func (s *Storage) AuthRequestDone(id string) error {
// 	s.lock.Lock()
// 	defer s.lock.Unlock()

// 	if req, ok := s.authRequests[id]; ok {
// 		req.done = true
// 		return nil
// 	}

// 	return errors.New("request not found")
// }

// func (s *Storage) ClientCredentials(ctx context.Context, clientID, clientSecret string) (op.Client, error) {
// 	s.lock.Lock()
// 	defer s.lock.Unlock()

// 	client, ok := s.serviceUsers[clientID]
// 	if !ok {
// 		return nil, errors.New("wrong service user or password")
// 	}
// 	if client.secret != clientSecret {
// 		return nil, errors.New("wrong service user or password")
// 	}

// 	return client, nil
// }

// func (s *Storage) ClientCredentialsTokenRequest(ctx context.Context, clientID string, scopes []string) (op.TokenRequest, error) {
// 	client, ok := s.serviceUsers[clientID]
// 	if !ok {
// 		return nil, errors.New("wrong service user or password")
// 	}

// 	return &oidc.JWTTokenRequest{
// 		Subject:  client.id,
// 		Audience: []string{clientID},
// 		Scopes:   scopes,
// 	}, nil
// }

// func (s *Storage) StoreDeviceAuthorization(ctx context.Context, clientID, deviceCode, userCode string, expires time.Time, scopes []string) error {
// 	s.lock.Lock()
// 	defer s.lock.Unlock()

// 	if _, ok := s.clients[clientID]; !ok {
// 		return errors.New("client not found")
// 	}

// 	if _, ok := s.userCodes[userCode]; ok {
// 		return op.ErrDuplicateUserCode
// 	}

// 	s.deviceCodes[deviceCode] = model.DeviceAuthorizationEntry{
// 		DeviceCode: deviceCode,
// 		UserCode:   userCode,
// 		State: &op.DeviceAuthorizationState{
// 			ClientID: clientID,
// 			Scopes:   scopes,
// 			Expires:  expires,
// 		},
// 	}

// 	s.userCodes[userCode] = deviceCode
// 	return nil
// }

// func (s *Storage) GetDeviceAuthorizatonState(ctx context.Context, clientID, deviceCode string) (*op.DeviceAuthorizationState, error) {
// 	if ctx.Err() != nil {
// 		return nil, ctx.Err()
// 	}

// 	s.lock.Lock()
// 	defer s.lock.Unlock()

// 	entry, ok := s.deviceCodes[deviceCode]
// 	if !ok || entry.State.ClientID != clientID {
// 		return nil, errors.New("device code not found for client") // is there a standard not found error in the framework?
// 	}

// 	return entry.State, nil
// }

// func (s *Storage) GetDeviceAuthorizationByUserCode(ctx context.Context, userCode string) (*op.DeviceAuthorizationState, error) {
// 	s.lock.Lock()
// 	defer s.lock.Unlock()

// 	entry, ok := s.deviceCodes[s.userCodes[userCode]]
// 	if !ok {
// 		return nil, errors.New("user code not found")
// 	}

// 	return entry.State, nil
// }

// func (s *Storage) CompleteDeviceAuthorization(ctx context.Context, userCode, subject string) error {
// 	s.lock.Lock()
// 	defer s.lock.Unlock()

// 	entry, ok := s.deviceCodes[s.userCodes[userCode]]
// 	if !ok {
// 		return errors.New("user code not found")
// 	}

// 	entry.State.Subject = subject
// 	entry.State.Done = true
// 	return nil
// }

// func (s *Storage) DenyDeviceAuthorization(ctx context.Context, userCode string) error {
// 	s.lock.Lock()
// 	defer s.lock.Unlock()

// 	s.deviceCodes[s.userCodes[userCode]].State.Denied = true
// 	return nil
// }

// // AuthRequestDone is used by testing and is not required to implement op.Storage
// func (s *Storage) AuthRequestDone(id string) error {
// 	s.lock.Lock()
// 	defer s.lock.Unlock()

// 	if req, ok := s.authRequests[id]; ok {
// 		// todo: req.done = true
// 		req.State = "done"
// 		return nil
// 	}

// 	return errors.New("request not found")
// }

// func (s *Storage) ClientCredentials(ctx context.Context, clientID, clientSecret string) (op.Client, error) {
// 	s.lock.Lock()
// 	defer s.lock.Unlock()

// 	client, ok := s.serviceUsers[clientID]
// 	if !ok {
// 		return nil, errors.New("wrong service user or password")
// 	}
// 	if client.secret != clientSecret {
// 		return nil, errors.New("wrong service user or password")
// 	}

// 	return client, nil
// }

// func (s *Storage) ClientCredentialsTokenRequest(ctx context.Context, clientID string, scopes []string) (op.TokenRequest, error) {
// 	client, ok := s.serviceUsers[clientID]
// 	if !ok {
// 		return nil, errors.New("wrong service user or password")
// 	}

// 	return &oidc.JWTTokenRequest{
// 		Subject:  client.id,
// 		Audience: []string{clientID},
// 		Scopes:   scopes,
// 	}, nil
// }

// func (s *Storage) Health(context.Context) error {
// 	return nil
// }
