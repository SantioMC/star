// Super simple library for working with the Star API

import { env } from "$env/dynamic/public";
import { getJWT, refreshAccessToken, token } from "@/index";
import { Thumbmark } from '@thumbmarkjs/thumbmarkjs';
import ky, { HTTPError, type KyRequest } from "ky";
import { get } from "svelte/store";

export const API_URL = env.PUBLIC_API_URL ?? '/';
export const APP_ID = "star";

// Permissions in the star management console, purely for typescript typings
export type StarPermission = "star:*"
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

export const api = ky.extend({
	prefixUrl: API_URL,
	retry: {
		limit: 3,
		retryOnTimeout: true,
		jitter: delay => delay + (Math.random() * 200 - 100)
	},
	hooks: {
		beforeRequest: [async (request: KyRequest, _a, _b) => {
			const authToken = await getAuth()
			if (!!authToken) {
				request.headers.set('Authorization', `Bearer ${authToken}`)
			}
		}]
	}
})

const getAuth = async () => {
	const value = get(token)
	if (!value) return null

	const data = getJWT(value)
	if (!data) return null

	// If our token is expired or will be in 5 seconds, we'll refresh it
	if (Math.floor(Date.now() / 1000) + 5 > data.exp) {
		return await refreshAccessToken()
	}

	return value
}

const thumbmark = new Thumbmark({
	logging: false,
})

const orTypedError = async <E>(e: Error): Promise<E> => {
	if (!(e instanceof HTTPError)) {
		throw new Error("ky didn't provide an http error??")
	}

	return e.response.json<E>()
}

const orError = async (e: Error) => orTypedError<StarError>(e)

export const getLogins = async (app: string) => {
	return api.get(`api/app/${encodeURIComponent(app)}/logins`).json<AppLogins>()
		.catch(_ => null)
}

export const login = async (app: string, username: string, password: string) => {
	const fingerprint = await thumbmark.get() as ThumbmarkResult

	return api.post(`api/app/${encodeURIComponent(app)}/login`, {
		json: {
			username,
			password,
			fingerprint: fingerprint.thumbmark,
			userAgent: navigator.userAgent
		}
	}).json<LoginResponse>().catch(orError)
}

export const refreshToken = async (refreshToken: string) => {
	const fingerprint = await thumbmark.get() as ThumbmarkResult

	const searchParams = new URLSearchParams();
	searchParams.set('grant_type', `refresh_token`);
	searchParams.set('refresh_token', refreshToken);
	searchParams.set('fp', fingerprint.thumbmark);

	return ky.post(`${API_URL}/oauth/token`, {
		body: searchParams
	}).json<OAuthTokenResponse>().catch(e => orTypedError<OAuthTokenError>(e))
}

export const getLoginUrl = (app: string, provider: string) => {
	return `${API_URL}/api/app/${encodeURIComponent(app)}/login/${encodeURIComponent(provider)}`
}

export const loginWithProvider = (provider: string) => {
	location.href = getLoginUrl(APP_ID, provider)
}

export const getSelf = async () => {
	return api.get('api/user/me').json<User>().catch(_ => null)
}

export const changePassword = async (currentPassword: string, newPassword: string) => {
	return api.post(`api/user/me/password`, {
		json: { current_password: currentPassword, new_password: newPassword }
	}).json<{ success: boolean }>().catch(orError)
}

export const updateUser = (changes: UserChanges) => {
	return api.patch(`api/user/me`, {
		json: changes
	}).json<User>().catch(orError)
}

export const getSessions = async () => {
	return api.get('api/session').json<Session[]>().catch(_ => [] as Session[])
}

export const getSession = async (id: string) => {
	return api.get(`api/session/${encodeURIComponent(id)}`).json<Session>().catch(_ => null)
}

export const invalidateSession = async (id: string) => {
	return api.delete(`api/session/${encodeURIComponent(id)}`).json<{ success: boolean }>().catch(orError)
}

export const invalidateAllSessions = async () => {
	return api.delete(`api/session`).json<{ success: boolean }>().catch(orError)
}

export const invalidateUserSessions = async (userId: string) => {
	return api.post(`api/user/${encodeURIComponent(userId)}/invalidate`).json<{ success: boolean }>().catch(orError)
}

export const getMetrics = async () => {
	return api.get(`api/metrics`).json<Metrics>().catch(_ => null)
}

export const getAllUsers = async (page: number = 1) => {
	return api.get(`api/user?page=${encodeURIComponent(page)}`).json<AdminUser[]>().catch(_ => [] as AdminUser[])
}

export const searchUsers = async (query: string, page: number = 1) => {
	return api.get(`api/user/search?query=${encodeURIComponent(query)}&page=${encodeURIComponent(page)}`).json<AdminUser[]>()
		.catch(_ => [] as AdminUser[])
}

export const getUser = async (id: string) => {
	return api.get(`api/user/${encodeURIComponent(id)}`).json<AdminUser>().catch(_ => null)
}

export const deleteUser = async (id: string) => {
	return api.delete(`api/user/${encodeURIComponent(id)}`).json<{ success: boolean }>().catch(orError)
}

export const suspendUser = async (id: string, reason: string, duration: number | null = null) => {
	return api.post(`api/user/${encodeURIComponent(id)}/suspend`, {
		json: {
			reason,
			duration
		}
	}).json<{ success: boolean }>().catch(orError)
}

export const unsuspendUser = async (id: string) => {
	return api.post(`api/user/${encodeURIComponent(id)}/unsuspend`).json<{ success: boolean }>().catch(orError)
}

export const forceChangePassword = async (user: string, newPassword: string) => {
	return api.post(`api/user/${encodeURIComponent(user)}/password`, {
		json: {
			password: newPassword
		}
	}).json<{ success: boolean }>().catch(orError)
}

export const modifyUser = (user: string, changes: AdminUserChanges) => {
	return api.patch(`api/user/${encodeURIComponent(user)}`, {
		json: changes
	}).json<User>().catch(orError)
}

export const getAllApps = async () => {
	return api.get(`api/app`).json<AdminApplication[]>().catch(_ => [] as any[])
}

export const getApplication = async (id: string) => {
	return api.get(`api/app/${encodeURIComponent(id)}`).json<AdminApplication>().catch(_ => null)
}

export const createApplication = async (data: CreateAppRequest) => {
	return api.post(`api/app`, { json: data }).json<AdminApplication>().catch(orError)
}

export const deleteApplication = async (app: string) => {
	return api.delete(`api/app/${encodeURIComponent(app)}`).json<{ success: boolean }>().catch(orError)
}

export const updateApplication = async (app: string, changes: AppChanges) => {
	return api.patch(`api/app/${encodeURIComponent(app)}`, { json: changes }).json<AdminApplication>().catch(orError)
}

export const createProvider = async (app: string, kind: string) => {
	return api.post(`api/app/${encodeURIComponent(app)}/provider`, { json: { kind } }).json<OAuthProvider>().catch(orError)
}

export const deleteProvider = async (app: string, provider: string) => {
	return api.delete(`api/app/${encodeURIComponent(app)}/provider/${encodeURIComponent(provider)}`).json<{ success: boolean }>().catch(orError)
}

export const rotateProvider = async (app: string, provider: string) => {
	return api.post(`api/app/${encodeURIComponent(app)}/provider/${encodeURIComponent(provider)}/rotate`).json<{ secret: string }>().catch(orError)
}

export const getPermissions = async (app: string | null = null) => {
	if (app) return await api.get(`api/app/${encodeURIComponent(app)}/permission`).json<Permission[]>().catch(() => [] as Permission[]);
	return [] as Permission[]
}

export const getAllRoles = async () => {
	return api.get('api/role').json<Role[]>().catch(() => [] as Role[])
}

export const getRole = async (id: string) => {
	return api.get(`api/role/${encodeURIComponent(id)}`).json<Role>().catch(_ => null)
}

export const createRole = async (data: CreateRoleRequest) => {
	return api.post('api/role', { json: data }).json<Role>().catch(orError)
}

export const deleteRole = async (id: string) => {
	return api.delete(`api/role/${id}`).json<{ success: boolean }>().catch(orError)
}

interface ThumbmarkResult {
	thumbmark: string;
	components: Record<string, any>;
	info: {
		uniqueness: {
			score: string;
		}
	};
	version: string;
}

export interface StarError {
	error: boolean
	status: number
	message: string
}

export interface SocialProvider {
	id: string
	name: string
	icon: string
}

export interface AppLogins {
	user_pass_login: boolean
	providers: SocialProvider[]
}

export interface LoginResponse {
	refresh_token: string
	access_token: string
	expires_at: number
}

export interface CreateAppRequest {
	name: string
	slug?: string
	url: string
}

export interface CreateRoleRequest {
	name: string
	permissions?: string[]
}

export interface Role {
	id: string
	created_at: number
	updated_at: number
	name: string
	super: boolean
	global: boolean
	permissions: string[]
}

export interface Metadata {
	public: {
		[key: string]: any
	},
	private: {
		[key: string]: any
	} | null
}

export interface AppChanges {
	name?: string,
	slug?: string,
	url?: string,
	callback_url?: string
}

export interface UserChanges {
	username?: string
	first_name?: string
	last_name?: string
}

export interface AdminUserChanges extends UserChanges {
	profile_picture?: string
	metadata?: Metadata
	email?: string
	roles?: string[]
}

export interface User {
	id: string
	created_at: number
	updated_at: number
	username: string
	first_name: string
	last_name: string
	email: string
	email_verified: boolean
	profile_picture: string
	metadata: Metadata
	roles: Role[]
	accounts: Account[]
	sessions: Session[]
}

export interface AdminUser extends User {
	last_login_at: number | null
	suspension: Suspension | null
}

export interface Role {
	id: string
	created_at: number
	updated_at: number
	name: string
	description: string
	super: boolean
	global: boolean
}

export interface Suspension {
	id: string
	created_at: number
	updated_at: number
	user_id: string
	issued_by: string
	reason: string
	expires_at: number | null
}

export interface Account {
	id: string
	created_at: number
	updated_at: number
	user_id: string
	provider_id: string | null
	username: string
	platform_id: string
	platform_avatar: string
}

export interface Metrics {
	total_users: number
	total_sessions: number
	started_at: number
	version: string
	node_id: string
}

export interface Session {
	id: string
	device: string
	operating_system: string
	ip_address: string
	created_at: number
	invalidated_at: number | null
	expires_at: number
	expired: boolean
}

export interface AdminApplication {
	id: string
	created_at: number
	updated_at: number
	name: string
	slug: string
	url: string
	callback: string
	internal: boolean
	oauth_provider: OAuthProvider | null
}

export interface OAuthProvider {
	id: string
	created_at: number
	updated_at: number
	application_id: string
	client_id: string
	client_secret: string
	rotate_refresh_token: boolean
	redirect_uris: string[]
	redirect_uri_globs: string[]
	application_type: 'public' | 'confidential'
	response_types: string[]
	grant_types: string[]
	access_token_type: string
}

export interface Permission {
	id: string
	created_at: number
	updated_at: number
	scope: string
	description: string
}

export interface OAuthTokenResponse {
	access_token: string
	token_type: string
	expires_in: number
	refresh_token?: string
}

export interface OAuthTokenError {
	error: string
	error_description?: string
	error_uri?: string
}