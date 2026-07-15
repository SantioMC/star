// Super simple library for working with the Star API

import { env } from "$env/dynamic/public";
import { getJWT, refreshAccessToken, token } from "@/index";
import ky, { HTTPError, type KyRequest } from "ky";
import { get } from "svelte/store";

export const API_URL = env.PUBLIC_API_URL ?? "/";

// Permissions in the star management console, purely for typescript typings
export type StarPermission =
  | "star:*"
  | "star:user:*"
  | "star:user:email"
  | "star:user:read:sessions"
  | "star:user:delete:sessions"
  | "star:user:delete"
  | "star:server:metrics"
  | "star:admin:*"
  | "star:admin:read:users"
  | "star:admin:update:users"
  | "star:admin:delete:users"
  | "star:admin:read:roles"
  | "star:admin:create:roles"
  | "star:admin:update:roles"
  | "star:admin:delete:roles"
  | "star:admin:read:applications"
  | "star:admin:update:applications"
  | "star:admin:create:applications"
  | "star:admin:delete:applications"
  | "star:admin:read:providers"
  | "star:admin:create:providers"
  | "star:admin:delete:providers"
  | "star:admin:read:keys"
  | "star:admin:create:keys"
  | "star:admin:delete:keys";

export const api = ky.extend({
  prefixUrl: API_URL,
  retry: {
    limit: 3,
    retryOnTimeout: true,
    jitter: (delay) => delay + (Math.random() * 200 - 100),
  },
  hooks: {
    beforeRequest: [
      async (request: KyRequest, _a, _b) => {
        const authToken = await getAuth();
        if (!!authToken) {
          request.headers.set("Authorization", `Bearer ${authToken}`);
        }
      },
    ],
  },
});

const getAuth = async () => {
  const value = get(token);
  if (!value) return null;

  const data = getJWT(value);
  if (!data) return null;

  // If our token is expired or will be in 5 seconds, we'll refresh it
  if (Math.floor(Date.now() / 1000) + 5 > data.exp) {
    return await refreshAccessToken();
  }

  return value;
};

const orTypedError = async <E>(e: Error): Promise<E> => {
  if (!(e instanceof HTTPError)) {
    throw new Error("ky didn't provide an http error??");
  }

  return e.response.json<E>();
};

const orError = async (e: Error) => orTypedError<StarError>(e);

export const getAuthUrl = () => `${API_URL}/oauth/authorize`;
export const getTokenUrl = () => `${API_URL}/oauth/token`;
export const getUserinfoUrl = () => `${API_URL}/oidc/userinfo`;
export const getDiscoveryURL = () =>
  `${API_URL}/.well-known/openid-configuration`;

// Stubbed — social login flow not yet implemented in the new API
export const getLoginUrl = (provider: string) => {
  return `${API_URL}/api/login/${encodeURIComponent(provider)}`;
};

// Stubbed — login methods discovery not yet implemented in the new API
export interface SocialProvider {
  id: string;
  name: string;
  icon: string;
}

export interface AppLogins {
  user_pass_login: boolean;
  providers: SocialProvider[];
}

export const getLogins = async (): Promise<AppLogins | null> => {
  return { user_pass_login: true, providers: [] };
};

export const login = async (data: LoginRequest) => {
  return api
    .post(`api/user/login`, {
      json: {
        username: data.username,
        password: data.password,
      },
      headers: data.captcha
        ? {
            "X-Captcha-Token": data.captcha,
          }
        : {},
    })
    .json<LoginResponse>()
    .catch(orError);
};

export const register = async (data: RegisterRequest) => {
  return api
    .post(`api/user/register`, {
      json: {
        username: data.username,
        password: data.password,
        email: data.email,
        first_name: data.first_name,
        last_name: data.last_name,
      },
      headers: data.captcha
        ? {
            "X-Captcha-Token": data.captcha,
          }
        : {},
    })
    .json<RegisterResponse>()
    .catch(orError);
};

export const refreshToken = async (refreshToken: string) => {
  const searchParams = new URLSearchParams();
  searchParams.set("grant_type", `refresh_token`);
  searchParams.set("refresh_token", refreshToken);

  return ky
    .post(`${API_URL}/oauth/token`, {
      body: searchParams,
    })
    .json<OAuthTokenResponse>()
    .catch((e) => orTypedError<OAuthTokenError>(e));
};

export const getOAuthApplication = (clientId: string) => {
  return api
    .get(`oauth/application/${encodeURIComponent(clientId)}`)
    .json<PublicApplication>()
    .catch((_) => null);
};

export const getOAuthAuthCode = async (
  appId: string,
  scope: string,
  redirectUri: string,
  state: string | null,
  codeChallenge: string | null = null,
  codeChallengeMethod: string | null = null,
  nonce: string | null = null,
) => {
  return api
    .post(`api/app/authorization`, {
      json: {
        application_id: appId,
        scope,
        redirect_uri: redirectUri,
        state: state ?? undefined,
        code_challenge: codeChallenge ?? undefined,
        code_challenge_method: codeChallengeMethod ?? undefined,
        nonce: nonce ?? undefined,
      },
    })
    .json<AuthorizationCodeResponse>()
    .catch(orError);
};

// Stubbed — endpoint not yet implemented in the new API
export const getUserinfo = async (): Promise<UserinfoResponse | null> => {
  return api
    .get("oidc/userinfo")
    .json<UserinfoResponse>()
    .catch((_) => null);
};

// Stubbed — endpoint not yet implemented in the new API
export const getSelf = async (): Promise<User | null> => {
  // Use the OIDC userinfo endpoint as a base, then fill in stubs for missing fields
  const info = await getUserinfo();
  if (!info) return null;

  return {
    id: info.sub,
    created_at: 0,
    updated_at: 0,
    username: info.preferred_username ?? info.nickname ?? info.name ?? "",
    first_name: info.given_name ?? "",
    last_name: info.family_name ?? "",
    email: info.email ?? "",
    email_verified: info.email_verified ?? false,
    profile_picture: info.picture ?? "",
    metadata: { public: {}, private: null },
    roles: [],
    accounts: [],
    sessions: [],
  };
};

export const changePassword = async (
  currentPassword: string,
  newPassword: string,
) => {
  return api
    .post(`api/user/me/password`, {
      json: { current_password: currentPassword, new_password: newPassword },
    })
    .json<{ success: boolean }>()
    .catch(orError);
};

export const resendVerification = async (email: string) => {
  return api
    .post(`api/user/me/resend`, {
      json: { email },
    })
    .json<{ success: boolean }>()
    .catch(orError);
};

export const verifyEmail = async (email: string, code: string) => {
  return api
    .post(`api/email/verify`, {
      json: { email, code },
    })
    .json<{ success: boolean }>()
    .catch(orError);
};

export const updateUser = (changes: UserChanges) => {
  return api
    .patch(`api/user/me`, {
      json: changes,
    })
    .json<User>()
    .catch(orError);
};

export const getSessions = async () => {
  return api
    .get("api/session")
    .json<Session[]>()
    .catch((_) => [] as Session[]);
};

export const getSession = async (id: string) => {
  return api
    .get(`api/session/${encodeURIComponent(id)}`)
    .json<Session>()
    .catch((_) => null);
};

export const invalidateSession = async (id: string) => {
  return api
    .delete(`api/session/${encodeURIComponent(id)}`)
    .json<{ success: boolean }>()
    .catch(orError);
};

export const invalidateAllSessions = async () => {
  return api.delete(`api/session`).json<{ success: boolean }>().catch(orError);
};

export const getEmails = async () => {
  return api
    .get("api/email")
    .json<Email[]>()
    .catch((_) => [] as Email[]);
};

export const addEmail = async (email: string) => {
  return api
    .post("api/email", { json: { email } })
    .json<Email>()
    .catch(orError);
};

export const deleteEmail = async (id: string) => {
  return api
    .delete(`api/email/${encodeURIComponent(id)}`)
    .json<{ success: boolean }>()
    .catch(orError);
};

export const setPrimaryEmail = async (id: string) => {
  return api
    .post(`api/email/${encodeURIComponent(id)}/primary`)
    .json<{ success: boolean }>()
    .catch(orError);
};

export const invalidateUserSessions = async (userId: string) => {
  return api
    .post(`api/user/${encodeURIComponent(userId)}/invalidate`)
    .json<{ success: boolean }>()
    .catch(orError);
};

export const getMetrics = async () => {
  return api
    .get(`api/metrics`)
    .json<Metrics>()
    .catch((_) => null);
};

export const getAllUsers = async (page: number = 1) => {
  return api
    .get(`api/user?page=${encodeURIComponent(page)}`)
    .json<AdminUser[]>()
    .catch((_) => [] as AdminUser[]);
};

export const searchUsers = async (query: string, page: number = 1) => {
  return api
    .get(
      `api/user/search?query=${encodeURIComponent(query)}&page=${encodeURIComponent(page)}`,
    )
    .json<AdminUser[]>()
    .catch((_) => [] as AdminUser[]);
};

export const getUser = async (id: string) => {
  return api
    .get(`api/user/${encodeURIComponent(id)}`)
    .json<AdminUser>()
    .catch((_) => null);
};

export const deleteUser = async (id: string) => {
  return api
    .delete(`api/user/${encodeURIComponent(id)}`)
    .json<{ success: boolean }>()
    .catch(orError);
};

export const suspendUser = async (
  id: string,
  reason: string,
  duration: number | null = null,
) => {
  return api
    .post(`api/user/${encodeURIComponent(id)}/suspend`, {
      json: {
        reason,
        duration,
      },
    })
    .json<{ success: boolean }>()
    .catch(orError);
};

export const unsuspendUser = async (id: string) => {
  return api
    .post(`api/user/${encodeURIComponent(id)}/unsuspend`)
    .json<{ success: boolean }>()
    .catch(orError);
};

export const forceChangePassword = async (
  user: string,
  newPassword: string,
) => {
  return api
    .post(`api/user/${encodeURIComponent(user)}/password`, {
      json: {
        password: newPassword,
      },
    })
    .json<{ success: boolean }>()
    .catch(orError);
};

export const modifyUser = (user: string, changes: AdminUserChanges) => {
  return api
    .patch(`api/user/${encodeURIComponent(user)}`, {
      json: changes,
    })
    .json<User>()
    .catch(orError);
};

export const getAllApps = async () => {
  return api
    .get(`api/app`)
    .json<AdminApplication[]>()
    .catch((_) => [] as any[]);
};

export const getApplication = async (id: string) => {
  return api
    .get(`api/app/${encodeURIComponent(id)}`)
    .json<AdminApplication>()
    .catch((_) => null);
};

export const createApplication = async (data: CreateAppRequest) => {
  return api
    .post(`api/app`, { json: data })
    .json<AdminApplication>()
    .catch(orError);
};

export const deleteApplication = async (app: string) => {
  return api
    .delete(`api/app/${encodeURIComponent(app)}`)
    .json<{ success: boolean }>()
    .catch(orError);
};

export const updateApplication = async (app: string, changes: AppChanges) => {
  return api
    .patch(`api/app/${encodeURIComponent(app)}`, { json: changes })
    .json<AdminApplication>()
    .catch(orError);
};

export const createProvider = async (app: string, kind: string) => {
  return api
    .post(`api/app/${encodeURIComponent(app)}/provider`, { json: { kind } })
    .json<OAuthProvider>()
    .catch(orError);
};

export const deleteProvider = async (app: string, provider: string) => {
  return api
    .delete(
      `api/app/${encodeURIComponent(app)}/provider/${encodeURIComponent(provider)}`,
    )
    .json<{ success: boolean }>()
    .catch(orError);
};

export const rotateProvider = async (app: string, provider: string) => {
  return api
    .post(
      `api/app/${encodeURIComponent(app)}/provider/${encodeURIComponent(provider)}/rotate`,
    )
    .json<{ secret: string }>()
    .catch(orError);
};

export const modifyProvider = async (
  app: string,
  provider: string,
  changes: ProviderChanges,
) => {
  return api
    .patch(
      `api/app/${encodeURIComponent(app)}/provider/${encodeURIComponent(provider)}`,
      {
        json: changes,
      },
    )
    .json<OAuthProvider>()
    .catch(orError);
};

export const addRedirectUri = async (
  app: string,
  provider: string,
  redirectUri: string,
) => {
  return api
    .post(
      `api/app/${encodeURIComponent(app)}/provider/${encodeURIComponent(provider)}/redirect`,
      {
        json: {
          redirect_uri: redirectUri,
        },
      },
    )
    .json<OAuthProvider>()
    .catch(orError);
};

export const removeRedirectUri = async (
  app: string,
  provider: string,
  index: number,
) => {
  return api
    .delete(
      `api/app/${encodeURIComponent(app)}/provider/${encodeURIComponent(provider)}/redirect/${encodeURIComponent(index)}`,
    )
    .json<OAuthProvider>()
    .catch(orError);
};

export const getPermissions = async (app: string | null = null) => {
  if (app)
    return api
      .get(`api/app/${encodeURIComponent(app)}/permission`)
      .json<Permission[]>()
      .catch(() => [] as Permission[]);
  return [] as Permission[];
};

export const createPermission = async (data: CreatePermissionRequest) => {
  return api
    .post(`api/app/${encodeURIComponent(data.app)}/permission`, {
      json: {
        name: data.permission,
        description: data.description,
      },
    })
    .json<Permission>()
    .catch(orError);
};

export const deletePermission = async (app: string, permission: string) => {
  return api
    .delete(
      `api/app/${encodeURIComponent(app)}/permission/${encodeURIComponent(permission)}`,
    )
    .json<{ success: boolean }>()
    .catch(orError);
};

export const getAllRoles = async () => {
  return api
    .get("api/role")
    .json<Role[]>()
    .catch(() => [] as Role[]);
};

export const getRole = async (id: string) => {
  return api
    .get(`api/role/${encodeURIComponent(id)}`)
    .json<Role>()
    .catch((_) => null);
};

export const createRole = async (data: CreateRoleRequest) => {
  return api.post("api/role", { json: data }).json<Role>().catch(orError);
};

export const updateRole = async (id: string, changes: UpdateRoleRequest) => {
  return api
    .patch(`api/role/${encodeURIComponent(id)}`, { json: changes })
    .json<Role>()
    .catch(orError);
};

export const deleteRole = async (id: string) => {
  return api
    .delete(`api/role/${encodeURIComponent(id)}`)
    .json<{ success: boolean }>()
    .catch(orError);
};

export const getSocialProviders = async () => {
  return api
    .get("api/provider")
    .json<SocialProviderConfig[]>()
    .catch(() => [] as SocialProviderConfig[]);
};

export const getSocialProviderPresets = async () => {
  return api
    .get("api/provider/presets")
    .json<SocialProviderPreset[]>()
    .catch(() => [] as SocialProviderPreset[]);
};

export const getSocialProvider = async (id: string) => {
  return api
    .get(`api/provider/${encodeURIComponent(id)}`)
    .json<SocialProviderConfig>()
    .catch((_) => null);
};

export const createSocialProvider = async (
  data: CreateSocialProviderRequest,
) => {
  return api
    .post("api/provider", { json: data })
    .json<SocialProviderConfig>()
    .catch(orError);
};

export const updateSocialProvider = async (
  id: string,
  changes: UpdateSocialProviderRequest,
) => {
  return api
    .patch(`api/provider/${encodeURIComponent(id)}`, { json: changes })
    .json<SocialProviderConfig>()
    .catch(orError);
};

export const deleteSocialProvider = async (id: string) => {
  return api
    .delete(`api/provider/${encodeURIComponent(id)}`)
    .json<{ success: boolean }>()
    .catch(orError);
};

export const getAllKeys = async () => {
  return api
    .get("api/key")
    .json<SigningKey[]>()
    .catch(() => [] as SigningKey[]);
};

export const getKey = async (id: string) => {
  return api
    .get(`api/key/${encodeURIComponent(id)}`)
    .json<SigningKey>()
    .catch((_) => null);
};

export const rotateSigningKey = async () => {
  return api.post("api/key/rotate").json<SigningKey>().catch(orError);
};

export const deleteKey = async (id: string) => {
  return api
    .delete(`api/key/${encodeURIComponent(id)}`)
    .json<{ success: boolean }>()
    .catch(orError);
};

export interface StarError {
  error: boolean;
  status: number;
  message: string;
}

export interface LoginRequest {
  username: string;
  password: string;
  captcha?: string;
}

export interface RegisterRequest {
  username: string;
  email: string;
  password: string;
  first_name?: string;
  last_name?: string;
  captcha?: string;
}

export interface RegisterResponse {
  success: boolean;
}

export interface PublicApplication {
  id: string;
  name: string;
  implicit: boolean;
}

export interface AuthorizationCodeResponse {
  code: string;
  state: string | null;
  scopes: string[];
  scoped_modified: boolean;
}

export interface LoginResponse {
  next: string;
  access_token: string;
  refresh_token: string;
}

export interface UserinfoResponse {
  sub: string;
  name?: string;
  given_name?: string;
  family_name?: string;
  nickname?: string;
  preferred_username?: string;
  picture?: string;
  email?: string;
  email_verified?: boolean;
}

export interface CreateAppRequest {
  name: string;
  slug?: string;
  url: string;
}

export interface CreatePermissionRequest {
  app: string;
  permission: string;
  description?: string;
}

export interface CreateRoleRequest {
  name: string;
  description?: string;
  scopes?: string[];
}

export interface UpdateRoleRequest {
  name?: string;
  description?: string;
  scopes?: string[];
}

export interface Role {
  id: string;
  created_at: number;
  updated_at: number;
  name: string;
  description: string;
  super: boolean;
  global: boolean;
  permissions: string[];
}

export interface Metadata {
  public: {
    [key: string]: any;
  };
  private: {
    [key: string]: any;
  } | null;
}

export interface AppChanges {
  name?: string;
  slug?: string;
  url?: string;
  callback_url?: string;
}

export interface UserChanges {
  username?: string;
  first_name?: string;
  last_name?: string;
}

export interface AdminUserChanges extends UserChanges {
  profile_picture?: string;
  metadata?: Metadata;
  email?: string;
  email_verified?: boolean;
  roles?: string[];
}

export interface ProviderChanges {
  fingerprinting?: boolean;
  application_type?: "public" | "confidential";
}

export interface User {
  id: string;
  created_at: number;
  updated_at: number;
  username: string;
  first_name: string;
  last_name: string;
  email: string;
  email_verified: boolean;
  profile_picture: string;
  metadata: Metadata;
  roles: Role[];
  accounts: Account[];
  sessions: Session[];
}

export interface AdminUser extends User {
  last_login_at: number | null;
  suspension: Suspension | null;
}

export interface Suspension {
  id: string;
  created_at: number;
  updated_at: number;
  user_id: string;
  issued_by: string;
  reason: string;
  expires_at: number | null;
}

export interface Account {
  id: string;
  created_at: number;
  updated_at: number;
  user_id: string;
  provider_id: string | null;
  username: string;
  platform_id: string;
  platform_avatar: string;
}

export interface Metrics {
  total_users: number;
  total_sessions: number;
  started_at: number;
  version: string;
  node_id: string;
}

export interface Session {
  id: string;
  ip_address: string;
  user_agent: string;
  geolocation: string | null;
  browser: string | "Unknown";
  status: "active" | "rotated";
  device: string;
  created_at: string;
  expires_at: string;
}

export interface Email {
  id: string;
  email: string;
  verified: boolean;
  primary: boolean;
  created_at: string;
}

export interface AdminApplication {
  id: string;
  created_at: number;
  updated_at: number;
  name: string;
  slug: string;
  url: string;
  callback: string;
  internal: boolean;
  oauth_provider: OAuthProvider | null;
}

export interface OAuthProvider {
  id: string;
  created_at: number;
  updated_at: number;
  application_id: string;
  client_id: string;
  client_secret: string;
  rotate_refresh_token: boolean;
  fingerprinting: boolean;
  redirect_uris: string[];
  application_type: "public" | "confidential";
  response_types: string[];
  grant_types: string[];
  access_token_type: string;
}

export interface Permission {
  id: string;
  created_at: number;
  updated_at: number;
  application_id: string;
  scope: string;
  description: string;
}

export interface OAuthTokenResponse {
  access_token: string;
  token_type: string;
  expires_in: number;
  refresh_token?: string;
}

export interface OAuthTokenError {
  error: string;
  error_description?: string;
  error_uri?: string;
}

export interface SocialProviderConfig {
  id: string;
  created_at: number;
  updated_at: number;
  name: string;
  order: number;
  client_id: string;
  client_secret: string;
  scopes: string[];
  icon: string;
  oidc_url: string | null;
  auth_url: string;
  token_url: string;
  userinfo_url: string;
}

export interface SocialProviderPreset {
  name: string;
  icon: string;
}

export interface CreateSocialProviderRequest {
  name: string;
  client_id: string;
  client_secret: string;
  discovery_uri?: string;
  order?: number;
}

export interface UpdateSocialProviderRequest {
  client_id?: string;
  client_secret?: string;
  icon?: string;
  order?: number;
  auth_url?: string;
  token_url?: string;
  userinfo_url?: string;
}

export interface SigningKey {
  id: string;
  created_at: number;
  updated_at: number;
  name: string;
  kind: "signing" | "next" | "retired";
  expires_at: number | null;
  fingerprint: string;
}
