// Super simple library for working with the Star API

import { env } from "$env/dynamic/public";
import { getJWT, refreshAccessToken, token } from "@/index";
import type { UsernameRules } from "@/username";
import type {
  AuthenticationResponseJSON,
  PublicKeyCredentialCreationOptionsJSON,
  PublicKeyCredentialRequestOptionsJSON,
  RegistrationResponseJSON,
} from "@simplewebauthn/browser";
import ky, { HTTPError, type KyRequest } from "ky";
import { get } from "svelte/store";

export const API_URL = env.PUBLIC_API_URL ?? "/";

// Permissions in the star management console, purely for typescript typings
export type StarPermission =
  | "*"
  | "user:password:write"
  | "user:password:reset"
  | "user:application:authorize"
  | "user:application:revoke"
  | "user:email:read"
  | "user:email:create"
  | "user:email:delete"
  | "user:email:primary"
  | "user:connection:write"
  | "user:mfa:enroll"
  | "user:mfa:delete"
  | "user:session:delete"
  | "user:data:export"
  | "application:create"
  | "application:read"
  | "application:write"
  | "application:delete"
  | "admin:user:read"
  | "admin:user:write"
  | "admin:user:delete"
  | "admin:user:export"
  | "admin:user:impersonate"
  | "admin:password:require"
  | "admin:role:read"
  | "admin:role:create"
  | "admin:role:write"
  | "admin:role:delete"
  | "admin:provider:read"
  | "admin:provider:create"
  | "admin:provider:write"
  | "admin:provider:delete"
  | "admin:settings:read"
  | "admin:settings:write"
  | "admin:key:read"
  | "admin:key:rotate"
  | "admin:key:delete"
  | "admin:token:read"
  | "admin:token:revoke"
  | "admin:metrics:read"
  | "admin:alert:read"
  | "admin:alert:resolve"
  | "admin:audit:read"
  | "admin:audit:write";

// Scopes issued by star itself, purely for typescript typings
export type StarScope =
  | "openid"
  | "offline_access"
  | "profile"
  | "oidc"
  | "email"
  | "roles"
  | "star:user:read"
  | "star:user:write"
  | "star:user:password:write"
  | "star:user:email:read"
  | "star:user:email:write"
  | "star:user:birthday:read"
  | "star:user:roles:read"
  | "star:user:sessions:read"
  | "star:user:sessions:write"
  | "star:user:mfa:read"
  | "star:user:mfa:write"
  | "star:user:applications:authorize"
  | "star:user:applications:read"
  | "star:user:applications:revoke"
  | "star:applications:read"
  | "star:applications:write"
  | "star:applications:delete"
  | "star:admin:users:read"
  | "star:admin:users:write"
  | "star:admin:users:delete"
  | "star:admin:users:impersonate"
  | "star:admin:roles:read"
  | "star:admin:roles:write"
  | "star:admin:providers:read"
  | "star:admin:providers:write"
  | "star:admin:settings:read"
  | "star:admin:settings:write"
  | "star:admin:keys:read"
  | "star:admin:keys:write"
  | "star:admin:tokens:read"
  | "star:admin:tokens:write"
  | "star:admin:metrics:read"
  | "star:admin:alerts:read"
  | "star:admin:alerts:write"
  | "star:admin:audit:read"
  | "star:admin:audit:write";

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
    afterResponse: [
      async (request, _options, response) => {
        const rejected =
          response.status === 401 &&
          request.headers.has("Authorization") &&
          response.headers.has("WWW-Authenticate");

        if (!rejected) return response;

        const refreshed = await refreshAccessToken();
        if (!refreshed) return response;

        return ky(request, {
          headers: { Authorization: `Bearer ${refreshed}` },
        });
      },
    ],
  },
});

const credentialed = api.extend({ credentials: "include" });

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

export const getProviderLoginUrl = (provider: string) =>
  `${API_URL}/api/user/login/provider/${encodeURIComponent(provider)}`;

export const getProviderCallbackUrl = (provider: string) =>
  `${getProviderLoginUrl(provider)}/callback`;

export interface LoginProvider {
  id: string;
  name: string;
  icon: string;
}

export const getLoginProviders = async () => {
  return api
    .get(`api/user/login/providers`)
    .json<LoginProvider[]>()
    .catch(() => [] as LoginProvider[]);
};

export interface ConnectionAccount {
  id: string;
  name: string | null;
  email: string | null;
  picture: string | null;
  created_at: string;
}

export interface Connection extends LoginProvider {
  account: ConnectionAccount | null;
}

export const getConnections = async () => {
  return api
    .get("api/connection")
    .json<Connection[]>()
    .catch(() => [] as Connection[]);
};

export const startConnection = async (provider: string) => {
  return credentialed
    .post(`api/connection/${encodeURIComponent(provider)}`)
    .json<{ url: string }>()
    .catch(orError);
};

export const deleteConnection = async (provider: string) => {
  return api
    .delete(`api/connection/${encodeURIComponent(provider)}`)
    .json<{ success: boolean }>()
    .catch(orError);
};

export const completeProviderLogin = async (flow: string) => {
  return credentialed
    .post(`api/user/login/provider/complete`, {
      json: { flow },
    })
    .json<LoginResponse>()
    .catch(orError);
};

export const login = async (data: LoginRequest) => {
  return credentialed
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

export const loginTotp = async (flow: string, code: string) => {
  return credentialed
    .post(`api/user/login/totp`, {
      json: { flow, code },
    })
    .json<LoginResponse>()
    .catch(orError);
};

export const loginBackupCode = async (flow: string, code: string) => {
  return credentialed
    .post(`api/user/login/backup`, {
      json: { flow, code },
    })
    .json<LoginResponse>()
    .catch(orError);
};

export const loginPassword = async (flow: string, password: string) => {
  return credentialed
    .post(`api/user/login/password`, {
      json: { flow, password },
    })
    .json<LoginResponse>()
    .catch(orError);
};

export const loginBirthday = async (flow: string, birthday: string) => {
  return credentialed
    .post(`api/user/login/birthday`, {
      json: { flow, birthday },
    })
    .json<LoginResponse>()
    .catch(orError);
};

export const getPasskeyLoginOptions = async (flow?: string) => {
  return api
    .post(`api/user/login/passkey/options`, {
      json: { flow },
    })
    .json<PasskeyChallenge<PublicKeyCredentialRequestOptionsJSON>>()
    .catch(orError);
};

export const loginPasskey = async (
  challenge: string,
  credential: AuthenticationResponseJSON,
  flow?: string,
) => {
  return credentialed
    .post(`api/user/login/passkey`, {
      json: { challenge, credential, flow },
    })
    .json<LoginResponse>()
    .catch(orError);
};

export const getRegisterRequirements = async () => {
  return api
    .get(`api/user/register`)
    .json<RegisterRequirements>()
    .catch((_) => null);
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
        birthday: data.birthday,
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

export const refreshSession = async () => {
  return ky
    .post("api/user/session", { prefixUrl: API_URL, credentials: "include" })
    .json<SessionRefreshResponse>()
    .catch(orError);
};

export const endSession = async () => {
  return ky
    .delete("api/user/session", { prefixUrl: API_URL, credentials: "include" })
    .catch(() => null);
};

export const getOAuthApplication = (
  clientId: string,
  redirectUri: string,
  scope: string | null = null,
) => {
  const searchParams = new URLSearchParams();
  searchParams.set("redirect_uri", redirectUri);

  if (scope) {
    searchParams.set("scope", scope);
  }

  return api
    .get(`oauth/application/${encodeURIComponent(clientId)}?${searchParams}`)
    .json<PublicApplication>()
    .catch(orError);
};

export const getOAuthAuthCode = async (
  appId: string,
  scope: string,
  redirectUri: string,
  state: string | null,
  codeChallenge: string | null = null,
  codeChallengeMethod: string | null = null,
  nonce: string | null = null,
  consent: boolean = false,
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
        consent,
      },
    })
    .json<AuthorizationCodeResponse>()
    .catch(orError);
};

export const verifyDeviceCode = async (userCode: string) => {
  return api
    .post(`api/device/verify`, {
      json: { user_code: userCode },
    })
    .json<DeviceVerification>()
    .catch(orError);
};

export const decideDeviceAuthorization = async (
  userCode: string,
  approve: boolean,
) => {
  return api
    .post(`api/device/decision`, {
      json: { user_code: userCode, approve },
    })
    .json<{ success: boolean }>()
    .catch(orError);
};

// Stubbed - endpoint not yet implemented in the new API
export const getUserinfo = async (): Promise<UserinfoResponse | null> => {
  return api
    .get("oidc/userinfo")
    .json<UserinfoResponse>()
    .catch((_) => null);
};

// Stubbed - endpoint not yet implemented in the new API
export const getSelf = async (): Promise<User | null> => {
  return api
    .get("api/user/me")
    .json<User>()
    .catch((_) => null);
};

export const changePassword = async (
  currentPassword: string,
  newPassword: string,
) => {
  return api
    .post(`api/user/password/change`, {
      json: { old_password: currentPassword, new_password: newPassword },
    })
    .json<{ success: boolean }>()
    .catch(orError);
};

export const resendVerification = async (email: string) => {
  return api
    .post(`api/email/resend`, {
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

export const requestPasswordReset = async (email: string, captcha?: string) => {
  return api
    .post(`api/user/password/reset`, {
      json: { email },
      headers: captcha ? { "X-Captcha-Token": captcha } : {},
    })
    .json<{ success: boolean }>()
    .catch(orError);
};

export const verifyPasswordReset = async (token: string) => {
  return api
    .post(`api/user/password/reset/verify`, {
      json: { token },
    })
    .json<{ success: boolean }>()
    .catch(orError);
};

export const completePasswordReset = async (
  token: string,
  password: string,
) => {
  return api
    .post(`api/user/password/reset/complete`, {
      json: { token, password },
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

export const getAuthorizedApps = async () => {
  return api
    .get("api/grant")
    .json<AuthorizedApplication[]>()
    .catch((_) => null);
};

export const revokeAuthorizedApp = async (app: string) => {
  return api
    .delete(`api/grant/${encodeURIComponent(app)}`)
    .json<{ success: boolean }>()
    .catch(orError);
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

export const getMfaFactors = async () => {
  return api
    .get("api/mfa")
    .json<MfaFactorsResponse>()
    .catch((_) => null);
};

export const enrollTotp = async () => {
  return api.post("api/mfa/totp/enroll").json<TotpEnrollment>().catch(orError);
};

export const verifyTotp = async (code: string) => {
  return api
    .post("api/mfa/totp/verify", { json: { code } })
    .json<{ success: boolean; backup_codes: string[] | null }>()
    .catch(orError);
};

export const deleteTotp = async () => {
  return api.delete("api/mfa/totp").json<{ success: boolean }>().catch(orError);
};

export const getPasskeys = async () => {
  return api
    .get("api/mfa/passkeys")
    .json<Passkey[]>()
    .catch(() => [] as Passkey[]);
};

export const getPasskeySignal = async () => {
  return api
    .get("api/mfa/passkeys/signal")
    .json<PasskeySignal>()
    .catch(() => null);
};

export const enrollPasskey = async () => {
  return api
    .post("api/mfa/passkeys/enroll")
    .json<PasskeyChallenge<PublicKeyCredentialCreationOptionsJSON>>()
    .catch(orError);
};

export const verifyPasskey = async (
  challenge: string,
  name: string,
  credential: RegistrationResponseJSON,
) => {
  return api
    .post("api/mfa/passkeys/verify", {
      json: { challenge, name, credential },
    })
    .json<{ passkey: Passkey; backup_codes: string[] | null }>()
    .catch(orError);
};

export const renamePasskey = async (id: string, name: string) => {
  return api
    .patch(`api/mfa/passkeys/${encodeURIComponent(id)}`, { json: { name } })
    .json<Passkey>()
    .catch(orError);
};

export const deletePasskey = async (id: string) => {
  return api
    .delete(`api/mfa/passkeys/${encodeURIComponent(id)}`)
    .json<{ success: boolean }>()
    .catch(orError);
};

export const getBackupCodeCount = async () => {
  return api
    .get("api/mfa/backup")
    .json<{ remaining: number }>()
    .catch((_) => null);
};

export const regenerateBackupCodes = async () => {
  return api
    .post("api/mfa/backup")
    .json<{ backup_codes: string[] }>()
    .catch(orError);
};

const attachmentName = (response: Response): string => {
  const disposition = response.headers.get("Content-Disposition") ?? "";
  const match = disposition.match(/filename="([^"]+)"/);

  return match?.[1] ?? "star-export.json";
};

const toDownload = async (response: Response) => {
  const blob = await response.blob();
  return { blob, filename: attachmentName(response) };
};

const exportData = async (path: string) => {
  return api.get(path).then(toDownload).catch(orError);
};

export const exportSelf = async () => exportData("api/user/me/export");

export const exportUser = async (userId: string) =>
  exportData(`api/user/${encodeURIComponent(userId)}/export`);

export const getUserSessions = async (userId: string, before?: string) => {
  const user = encodeURIComponent(userId);
  const searchParams = new URLSearchParams({ limit: "10" });
  if (before) searchParams.set("before", before);

  return api
    .get(`api/user/${user}/sessions?${searchParams.toString()}`)
    .json<SessionPage>()
    .catch(orError);
};

export const invalidateUserSession = async (
  userId: string,
  sessionId: string,
) => {
  const user = encodeURIComponent(userId);
  const session = encodeURIComponent(sessionId);

  return api
    .delete(`api/user/${user}/sessions/${session}`)
    .json<{ success: boolean }>()
    .catch(orError);
};

export const invalidateUserSessions = async (userId: string) => {
  return api
    .delete(`api/user/${encodeURIComponent(userId)}/sessions`)
    .json<{ success: boolean }>()
    .catch(orError);
};

export const getMetrics = async () => {
  return api
    .get(`api/metrics`)
    .json<Metrics>()
    .catch((_) => null);
};

export const getAllUsers = async (
  page: number = 1,
  query: string = "",
  type: UserType | "" = "",
) => {
  const searchParams = new URLSearchParams();
  searchParams.set("page", page.toString());

  if (query) {
    searchParams.set("query", query);
  }

  if (type) {
    searchParams.set("type", type);
  }

  return api
    .get(`api/user?${searchParams.toString()}`)
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

export const clearLockout = async (user: string) => {
  return api
    .delete(`api/user/${encodeURIComponent(user)}/lockout`)
    .json<AdminUser>()
    .catch(orError);
};

export const getUserSuspensions = async (user: string, before?: string) => {
  const searchParams = new URLSearchParams({ limit: "10" });
  if (before) searchParams.set("before", before);

  return api
    .get(
      `api/user/${encodeURIComponent(user)}/suspensions?${searchParams.toString()}`,
    )
    .json<SuspensionPage>()
    .catch(orError);
};

export const createInvite = async (data: CreateInviteRequest) => {
  return api
    .post(`api/invite`, { json: data })
    .json<InviteLink>()
    .catch(orError);
};

export const createUser = async (data: CreateUserRequest) => {
  return api.post(`api/user`, { json: data }).json<AdminUser>().catch(orError);
};

export const regenerateInvite = async (user: string, expiresIn: number) => {
  return api
    .post(`api/user/${encodeURIComponent(user)}/invite`, {
      json: { expires_in: expiresIn },
    })
    .json<InviteLink>()
    .catch(orError);
};

export const emailInvite = async (user: string, code: string) => {
  return api
    .post(`api/user/${encodeURIComponent(user)}/invite/email`, {
      json: { code },
    })
    .json<{ success: boolean }>()
    .catch(orError);
};

export const lookupInvite = async (code: string) => {
  return api
    .post(`api/invite/lookup`, { json: { code } })
    .json<InviteDetails>()
    .catch(orError);
};

export const acceptInvite = async (data: AcceptInviteRequest) => {
  return api
    .post(`api/invite/accept`, { json: data })
    .json<{ success: boolean }>()
    .catch(orError);
};

export const suspendUser = async (user: string, data: SuspendUserRequest) => {
  return api
    .post(`api/user/${encodeURIComponent(user)}/suspend`, { json: data })
    .json<AdminUser>()
    .catch(orError);
};

export const unsuspendUser = async (user: string, reason: string) => {
  return api
    .post(`api/user/${encodeURIComponent(user)}/unsuspend`, {
      json: { reason },
    })
    .json<AdminUser>()
    .catch(orError);
};

export const getImpersonationApplications = async (
  user: string,
  page: number,
  query: string,
  limit: number,
) => {
  const searchParams = new URLSearchParams({
    page: page.toString(),
    limit: limit.toString(),
  });

  if (query) {
    searchParams.set("query", query);
  }

  return api
    .get(
      `api/user/${encodeURIComponent(user)}/impersonate/applications?${searchParams.toString()}`,
    )
    .json<SharedApplication[]>()
    .catch(orError);
};

export const impersonateUser = async (user: string, applicationId: string) => {
  return api
    .post(`api/user/${encodeURIComponent(user)}/impersonate`, {
      json: { application_id: applicationId },
    })
    .json<ImpersonationResponse>()
    .catch(orError);
};

export const adminSetPassword = async (user: string, password: string) => {
  return api
    .post(`api/user/${encodeURIComponent(user)}/password`, {
      json: { password },
    })
    .json<{ success: boolean }>()
    .catch(orError);
};

export const requirePasswordChange = async (user: string) => {
  return api
    .post(`api/user/${encodeURIComponent(user)}/password/require-change`)
    .json<AdminUser>()
    .catch(orError);
};

export const requirePasswordChangeForAll = async () => {
  return api
    .post(`api/user/password/require-change`)
    .json<{ affected: number }>()
    .catch(orError);
};

export const modifyUser = (user: string, changes: AdminUserChanges) => {
  return api
    .patch(`api/user/${encodeURIComponent(user)}`, {
      json: changes,
    })
    .json<AdminUser>()
    .catch(orError);
};

export const getAllApps = async () => {
  return api
    .get(`api/app`)
    .json<AdminApplication[]>()
    .catch((_) => [] as AdminApplication[]);
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
    .json<ApplicationSecretResponse>()
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
    .json<ApplicationSecretResponse>()
    .catch(orError);
};

export const getApplicationScopes = async (app: string) => {
  return api
    .get(`api/app/${encodeURIComponent(app)}/scopes`)
    .json<ApplicationScope[]>()
    .catch(() => [] as ApplicationScope[]);
};

export const createApplicationScope = async (
  app: string,
  data: CreateApplicationScopeRequest,
) => {
  return api
    .post(`api/app/${encodeURIComponent(app)}/scopes`, { json: data })
    .json<ApplicationScope>()
    .catch(orError);
};

export const updateApplicationScope = async (
  app: string,
  scope: string,
  changes: UpdateApplicationScopeRequest,
) => {
  return api
    .patch(
      `api/app/${encodeURIComponent(app)}/scopes/${encodeURIComponent(scope)}`,
      { json: changes },
    )
    .json<ApplicationScope>()
    .catch(orError);
};

export const deleteApplicationScope = async (app: string, scope: string) => {
  return api
    .delete(
      `api/app/${encodeURIComponent(app)}/scopes/${encodeURIComponent(scope)}`,
    )
    .json<{ success: boolean }>()
    .catch(orError);
};

export const rotateApplicationSecret = async (app: string) => {
  return api
    .post(`api/app/${encodeURIComponent(app)}/rotate`)
    .json<{ client_secret: string }>()
    .catch(orError);
};

export const getAllRoles = async () => {
  return api
    .get("api/role")
    .json<Role[]>()
    .catch(() => [] as Role[]);
};

export const getAvailablePermissions = async () => {
  return api
    .get("api/role/permissions")
    .json<AvailablePermission[]>()
    .catch(() => [] as AvailablePermission[]);
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

export const reorderRoles = async (roles: string[]) => {
  return api
    .post("api/role/order", { json: { roles } })
    .json<Role[]>()
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

export const getSigningKeys = async () => {
  return api
    .get("api/key")
    .json<SigningKey[]>()
    .catch(() => [] as SigningKey[]);
};

export const rotateSigningKeys = async () => {
  return api.post("api/key/rotate").json<SigningKey[]>().catch(orError);
};

export const deleteSigningKey = async (id: string) => {
  return api
    .delete(`api/key/${encodeURIComponent(id)}`)
    .json<SigningKey[]>()
    .catch(orError);
};

export const getTokenRevocations = async (page: number = 1) => {
  const searchParams = new URLSearchParams();
  searchParams.set("page", page.toString());

  return api
    .get(`api/revocation?${searchParams.toString()}`)
    .json<TokenRevocation[]>()
    .catch(() => [] as TokenRevocation[]);
};

export const revokeTokens = async (request: CreateRevocationRequest) => {
  return api
    .post("api/revocation", { json: request })
    .json<TokenRevocation[]>()
    .catch(orError);
};

export const getSecurityAlerts = async (
  page: number = 1,
  status: "open" | "all" = "open",
) => {
  const searchParams = new URLSearchParams();
  searchParams.set("page", page.toString());
  searchParams.set("status", status);

  return api
    .get(`api/alert?${searchParams.toString()}`)
    .json<SecurityAlert[]>()
    .catch(() => [] as SecurityAlert[]);
};

export const resolveSecurityAlert = async (id: string) => {
  return api
    .post(`api/alert/${encodeURIComponent(id)}/resolve`)
    .json<SecurityAlert>()
    .catch(orError);
};

export const getDiscovery = async () => {
  return ky
    .get(getDiscoveryURL())
    .json<DiscoveryDocument>()
    .catch(() => null);
};

export const getAuditLogs = async (filter: AuditLogFilter) => {
  const searchParams = new URLSearchParams();

  for (const [key, value] of Object.entries(filter)) {
    if (value) searchParams.set(key, value);
  }

  return api
    .get(`api/audit?${searchParams.toString()}`)
    .json<AuditLogPage>()
    .catch(orError);
};

export const getAuditActions = async () => {
  return api
    .get("api/audit/actions")
    .json<AuditActionDefinition[]>()
    .catch(() => [] as AuditActionDefinition[]);
};

export const getAuditSettings = async () => {
  return api
    .get("api/audit/settings")
    .json<AuditSettings>()
    .catch(() => null);
};

export const updateAuditSettings = async (
  changes: UpdateAuditSettingsRequest,
) => {
  return api
    .patch("api/audit/settings", { json: changes })
    .json<AuditSettings>()
    .catch(orError);
};

export const getServerSettings = async () => {
  return api
    .get("api/settings")
    .json<ServerSettings>()
    .catch((_) => null);
};

export const updateServerSettings = async (
  changes: UpdateServerSettingsRequest,
) => {
  return api
    .patch("api/settings", { json: changes })
    .json<ServerSettings>()
    .catch(orError);
};

export const getDisposableDomains = async () => {
  const response = await api
    .get("api/settings/disposable-domains")
    .json<DisposableDomains>()
    .catch((_) => null);

  return response?.domains ?? null;
};

export const getBranding = async () => {
  return api
    .get("api/branding")
    .json<Branding>()
    .catch((_) => null);
};

export const updateBranding = async (changes: UpdateBrandingRequest) => {
  return api
    .patch("api/branding", { json: changes })
    .json<Branding>()
    .catch(orError);
};

export const uploadBrandingLogo = async (logo: Blob) => {
  return api
    .put("api/branding/logo", { body: logo })
    .json<Branding>()
    .catch(orError);
};

export const deleteBrandingLogo = async () => {
  return api.delete("api/branding/logo").json<Branding>().catch(orError);
};

export const createEmailPreview = async (request: EmailPreviewRequest) => {
  return api
    .post("api/branding/preview", { json: request })
    .json<{ ticket: string }>()
    .catch(orError);
};

export const getEmailPreviewUrl = (ticket: string) =>
  `${API_URL}/api/branding/preview?ticket=${encodeURIComponent(ticket)}`;

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
  birthday?: string;
  captcha?: string;
}

export interface RegisterRequirements {
  username: UsernameRules;
  birthday_tracked: boolean;
  birthday_required: boolean;
}

export interface RegisterResponse {
  success: boolean;
  verification_required: boolean;
}

export interface PublicApplication {
  id: string;
  name: string;
  url: string;
  scopes: PublicScope[];
  invalid_scopes: string[];
}

export interface DeviceVerification {
  user_code: string;
  device: string;
  geolocation: string | null;
  expires_at: string;
  application: PublicApplication;
}

export interface PublicScope {
  scope: string;
  description: string;
  application: string;
}

export interface GrantedScope {
  scope: string;
  description: string;
}

export interface AuthorizedApplication {
  id: string;
  name: string;
  url: string;
  scopes: GrantedScope[];
  authorized_at: string;
  updated_at: string;
  last_used_at: string | null;
}

export interface ApplicationScope {
  id: string;
  application_id: string;
  name: string;
  description: string;
  requestable: boolean;
  created_at: string;
}

export interface CreateApplicationScopeRequest {
  name: string;
  description: string;
  requestable?: boolean;
}

export interface UpdateApplicationScopeRequest {
  name?: string;
  description?: string;
  requestable?: boolean;
}

export interface BuiltinScope {
  name: string;
  description: string;
  aliases?: string[];
  restricted: boolean;
  privileged: boolean;
}

export interface PasswordSettings {
  min_length: number;
  max_length: number;
  min_symbols: number;
  min_upper: number;
  min_lower: number;
}

export interface DisposableDomains {
  domains: string[];
}

export type EmailDomainMode = "disabled" | "whitelist" | "blacklist";

export interface ServerSettings {
  allow_registration: boolean;
  allow_logins: boolean;
  email_code_length: number;
  email_verification: boolean;
  email_domain_mode: EmailDomainMode;
  email_domains: string[];
  block_disposable_emails: boolean;
  track_birthdays: boolean;
  minimum_age: number;
  password_settings: PasswordSettings;
}

export interface UpdateServerSettingsRequest {
  allow_registration?: boolean;
  allow_logins?: boolean;
  email_code_length?: number;
  email_verification?: boolean;
  email_domain_mode?: EmailDomainMode;
  email_domains?: string[];
  block_disposable_emails?: boolean;
  track_birthdays?: boolean;
  minimum_age?: number;
  password_settings?: Partial<PasswordSettings>;
}

export type BrandingLayout = "centered" | "split" | "minimal";

export interface Branding {
  name: string;
  accent_color: string;
  heading_font: string;
  body_font: string;
  layout: BrandingLayout;
  logo_url: string | null;
  support_email: string | null;
  privacy_url: string | null;
  terms_url: string | null;
}

export type UpdateBrandingRequest = Partial<
  Omit<Branding, "logo_url" | "support_email" | "privacy_url" | "terms_url">
> & {
  support_email?: string;
  privacy_url?: string;
  terms_url?: string;
};

export type EmailTemplate =
  | "verification"
  | "password-reset"
  | "password-changed"
  | "invite";

export interface EmailPreviewRequest {
  template: EmailTemplate;
  branding?: UpdateBrandingRequest;
  remove_logo?: boolean;
}

export interface AuthorizationCodeResponse {
  code: string;
  state: string | null;
  scopes: string[];
  scoped_modified: boolean;
}

export type MFAMethod = "totp" | "passkey" | "backup_code";

export interface LoginSuccessResponse {
  next: "complete";
  access_token: string;
}

export interface SessionRefreshResponse {
  access_token: string;
  expires_in: number;
}

export interface LoginMfaResponse {
  next: "mfa_required";
  flow: string;
  factors: MFAMethod[];
}

export interface LoginPasswordChangeResponse {
  next: "pass_change";
  flow: string;
}

export interface LoginBirthdayResponse {
  next: "birthday";
  flow: string;
}

export type LoginResponse =
  | LoginSuccessResponse
  | LoginMfaResponse
  | LoginPasswordChangeResponse
  | LoginBirthdayResponse;

export interface MfaFactorsResponse {
  factors: MFAMethod[];
}

export interface PasskeyChallenge<T> {
  challenge: string;
  options: T;
}

export interface Passkey {
  id: string;
  name: string;
  aaguid: string | null;
  backup_eligible: boolean;
  synced: boolean;
  last_used_at: string | null;
  created_at: string;
}

export interface PasskeySignal {
  rp_id: string;
  user_id: string;
  name: string;
  display_name: string;
  credential_ids: string[];
}

export interface TotpEnrollment {
  secret: string;
  uri: string;
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
  kind: "public" | "confidential";
  url: string;
  client_id?: string;
  namespace?: string;
}

export interface CreateRoleRequest {
  name: string;
  description?: string;
  permissions?: string[];
}

export interface UpdateRoleRequest {
  name?: string;
  description?: string;
  permissions?: string[];
}

export interface Role {
  id: string;
  name: string;
  description: string;
  internal: boolean;
  default: boolean;
  weight: number;
  manageable: boolean;
  permissions: string[];
  created_at: string;
}

export interface AvailablePermission {
  id: string;
  name: string;
  description: string;
  created_at: string;
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
  kind?: "public" | "confidential";
  consent?: ApplicationConsent;
  url?: string;
  client_id?: string;
  namespace?: string;
  token_duration?: number;
  session_duration?: number;
  redirect_uris?: string[];
  development_mode?: boolean;
  device_flow?: boolean;
  privileged_scopes?: boolean;
  signing_algorithm?: SigningAlgorithm;
  role_restricted?: boolean;
  roles?: string[];
}

export interface ApplicationSecretResponse {
  id: string;
  client_secret?: string;
}

export interface UserChanges {
  username?: string;
  first_name?: string;
  last_name?: string;
  birthday?: string;
}

export interface AdminUserChanges extends UserChanges {
  roles?: string[];
  type?: UserType;
}

export interface User {
  id: string;
  username: string;
  first_name: string;
  last_name: string;
  picture: string;
  birthday?: string;
  email: string;
  email_verified: boolean;
  roles: string[];
  permissions: StarPermission[];
  sessions: Session[];
  accounts: Account[];
  emails: Email[];
  created_at: string;
}

export type UserType = "regular" | "service";

export interface CreateUserRequest {
  username: string;
  type: UserType;
}

export interface AdminUser {
  id: string;
  username: string;
  type: UserType;
  first_name: string | null;
  last_name: string | null;
  profile_picture: string;
  birthday: string | null;
  email: string | null;
  email_verified: boolean;
  roles: Role[];
  lockout: Lockout | null;
  suspension: UserSuspension | null;
  invite: UserInvite | null;
  has_password: boolean;
  password_change_required: boolean;
  manageable: boolean;
  created_at: string;
}

export interface SharedApplication {
  id: string;
  name: string;
  url: string;
  authorized: boolean;
}

export interface ImpersonationResponse {
  url: string;
  expires_at: string;
}

export interface UserSuspension {
  id: string;
  user_id: string;
  reason: string;
  suspended_by: string | null;
  suspended_by_name: string | null;
  expires_at: string | null;
  created_at: string;
  lifted_at: string | null;
  lifted_by: string | null;
  lifted_by_name: string | null;
  lift_reason: string | null;
}

export interface SuspensionPage {
  items: UserSuspension[];
  next: string | null;
}

export interface UserInvite {
  id: string;
  user_id: string;
  invited_by: string | null;
  invited_by_name: string | null;
  expires_at: string;
  created_at: string;
}

export interface CreateInviteRequest {
  username: string;
  email: string;
  first_name?: string;
  last_name?: string;
  birthday?: string;
  expires_in: number;
}

export interface InviteLink {
  user_id: string;
  code: string;
  url: string;
  expires_at: string;
  email_enabled: boolean;
}

export interface InviteDetails {
  username: string;
  email: string;
  first_name: string | null;
  last_name: string | null;
  birthday: string | null;
  birthday_tracked: boolean;
  birthday_required: boolean;
  expires_at: string;
}

export interface AcceptInviteRequest {
  code: string;
  password: string;
  first_name?: string;
  last_name?: string;
  birthday?: string;
}

export interface SuspendUserRequest {
  reason: string;
  expires_at: string | null;
}

export interface Lockout {
  locked: boolean;
  attempts: number;
  last_attempt_at: string;
  locked_until: string | null;
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
  users: number;
  active_sessions: number;
  applications: number;
  open_alerts: number;
  version: string;
  started_at: string;
}

export interface Session {
  id: string;
  ip_address: string;
  user_agent: string;
  geolocation: string | null;
  browser: string | "Unknown";
  status: "active" | "rotated";
  device: string;
  impersonator_id: string | null;
  created_at: string;
  expires_at: string;
}

export interface SessionPage {
  items: Session[];
  next: string | null;
}

export interface Email {
  id: string;
  email: string;
  verified: boolean;
  primary: boolean;
  created_at: string;
}

export type ApplicationConsent = "default" | "explicit" | "implicit";

export type SigningAlgorithm = "ES512" | "RS256";

export interface AdminApplication {
  id: string;
  kind: "public" | "confidential";
  name: string;
  namespace: string;
  application_url: string;
  client_id: string;
  has_client_secret: boolean;
  internal: boolean;
  consent: ApplicationConsent;
  token_duration: number;
  session_duration: number;
  redirect_uris: string[];
  development_mode: boolean;
  device_flow: boolean;
  privileged_scopes: boolean;
  signing_algorithm: SigningAlgorithm;
  role_restricted: boolean;
  roles: string[];
  scopes: ApplicationScope[];
  builtin_scopes?: BuiltinScope[];
  created_at: string;
}

export interface SocialProviderConfig {
  id: string;
  name: string;
  icon: string;
  order: number;
  client_id: string;
  has_client_secret: boolean;
  scopes: string[];
  oidc_url: string | null;
  authorization_url: string | null;
  token_url: string | null;
  userinfo_url: string | null;
  trust_email: boolean;
  enabled: boolean;
  created_at: string;
}

export interface SocialProviderPreset {
  name: string;
  icon: string;
  scopes?: string[];
  oidc_url?: string;
  authorization_url?: string;
  token_url?: string;
  userinfo_url?: string;
  trust_email: boolean;
}

export interface CreateSocialProviderRequest {
  name: string;
  icon?: string;
  order?: number;
  client_id: string;
  client_secret: string;
  scopes?: string[];
  oidc_url?: string;
  authorization_url?: string;
  token_url?: string;
  userinfo_url?: string;
  trust_email?: boolean;
}

export interface UpdateSocialProviderRequest {
  name?: string;
  icon?: string;
  order?: number;
  client_id?: string;
  client_secret?: string;
  scopes?: string[];
  oidc_url?: string;
  authorization_url?: string;
  token_url?: string;
  userinfo_url?: string;
  trust_email?: boolean;
  enabled?: boolean;
}

export interface SigningKey {
  id: string;
  status: "active" | "next" | "retiring";
  algorithm: string;
  fingerprint: string;
  created_at: string;
  expires_at: string | null;
}

export type SecurityAlertSeverity = "high" | "medium" | "low";

export type RevocationKind = "session" | "user" | "global";

export type RevocationReason =
  | "signed_out"
  | "sessions_cleared"
  | "grant_revoked"
  | "token_replayed"
  | "code_replayed"
  | "roles_changed"
  | "password_reset"
  | "password_changed"
  | "password_change_required"
  | "user_deleted"
  | "user_suspended"
  | "user_type_changed"
  | "application_deleted"
  | "application_access_changed"
  | "connection_removed"
  | "manual";

export interface TokenRevocation {
  subject: string;
  kind: RevocationKind;
  user_id: string | null;
  user_name: string | null;
  reason: RevocationReason;
  revoked_at: string;
  expires_at: string;
}

export type CreateRevocationRequest =
  { kind: "user"; user: string } | { kind: "global" };

export interface SecurityAlert {
  id: string;
  kind: string;
  severity: SecurityAlertSeverity;
  message: string;
  data: Record<string, unknown> | null;
  user_id: string | null;
  application: string | null;
  created_at: string;
  resolved_at: string | null;
  resolved_by: string | null;
}

export interface DiscoveryDocument {
  issuer: string;
  jwks_uri: string;
  authorization_endpoint: string;
  token_endpoint: string;
  device_authorization_endpoint: string;
  userinfo_endpoint: string;
  revocation_endpoint: string;
  introspection_endpoint: string;
}

export type AuditSeverity = "info" | "notice" | "critical";

export interface AuditActionDefinition {
  name: string;
  label: string;
  category: string;
  severity: AuditSeverity;
}

export interface AuditLog {
  id: string;
  action: string;
  actor_id: string | null;
  actor_name: string | null;
  target_type: string | null;
  target_id: string | null;
  target_name: string | null;
  application_id: string | null;
  provider_id: string | null;
  ip_address: string | null;
  user_agent: string | null;
  metadata: Record<string, unknown> | null;
  created_at: string;
}

export interface AuditLogPage {
  items: AuditLog[];
  next: string | null;
}

export interface AuditLogFilter {
  user?: string;
  application?: string;
  provider?: string;
  query?: string;
  category?: string;
  action?: string;
  since?: string;
  until?: string;
  before?: string;
}

export interface AuditSettings {
  retention_days: number | null;
  anonymize: boolean;
}

export interface UpdateAuditSettingsRequest {
  retention_days?: number;
  anonymize?: boolean;
}
