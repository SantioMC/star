// Super simple library for working with the Star API

import { env } from "$env/dynamic/public";
import { token } from "@/index";
import { Thumbmark } from '@thumbmarkjs/thumbmarkjs';
import ky, { HTTPError, type BeforeRequestState, type KyRequest, type NormalizedOptions } from "ky";
import { get } from "svelte/store";

const API_URL = env.PUBLIC_API_URL ?? "/api";
const APP_ID = "star";

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
	| "star:admin:read:applications"
	| "star:admin:update:applications"
	| "star:admin:create:applications"
	| "star:admin:delete:applications"

export type DefaultSocialProvider = 'google'
	| 'github'
	| 'twitch'
	| 'slack'
	| 'discord'

export const api = ky.extend({
	prefixUrl: API_URL,
	retry: {
		limit: 3,
		retryOnTimeout: true,
		jitter: delay => delay + (Math.random() * 200 - 100)
	},
	hooks: {
		beforeRequest: [async (request: KyRequest, options: NormalizedOptions, state: BeforeRequestState) => {
			const authToken = await getAuth()
			if (!!authToken) {
				request.headers.set('Authorization', `Bearer ${authToken}`)
			}
		}]
	}
})

const getAuth = async () => get(token)

const thumbmark = new Thumbmark({
	logging: false,
})

const orError = async (e: Error): Promise<StarError> => {
	if (!(e instanceof HTTPError)) {
		throw new Error("ky didn't provide an http error??")
	}

	return e.response.json<StarError>()
}

export const getLogins = async () => {
	return api.get(`app/${APP_ID}/logins`).json<AppLogins>()
		.catch(_ => null)
}

export const login = async (username: string, password: string) => {
	const fingerprint = await thumbmark.get() as ThumbmarkResult

	return api.post(`app/${APP_ID}/login`, {
		json: {
			username,
			password,
			fingerprint: fingerprint.thumbmark,
			userAgent: navigator.userAgent
		}
	}).json<LoginResponse>().catch(orError)
}

export const loginWithProvider = (provider: string) => {
	location.href = `${API_URL}/app/${APP_ID}/login/${provider}`
}

export const getSelf = async () => {
	return api.get('user/me').json<User>().catch(_ => null)
}

export const getSessions = async () => {
	return api.get('session').json<Session[]>().catch(_ => [] as Session[])
}

export const getSession = async (id: string) => {
	return api.get(`session/${id}`).json<Session>().catch(_ => null)
}

export const invalidateSession = async (id: string) => {
	return api.delete(`session/${id}`).json<{ success: boolean }>().catch(orError)
}

export const invalidateAllSessions = async () => {
	return api.delete(`session`).json<{ success: boolean }>().catch(orError)
}

export const invalidateUserSessions = async (userId: string) => {
	return api.post(`user/${userId}/invalidate`).json<{ success: boolean }>().catch(orError)
}

export const getMetrics = async () => {
	return api.get(`metrics`).json<Metrics>().catch(_ => null)
}

export const getAllUsers = async (page: number = 1) => {
	return api.get(`user?page=${page}`).json<AdminUser[]>().catch(_ => [] as AdminUser[])
}

export const searchUsers = async (query: string, page: number = 1) => {
	return api.get(`user/search?query=${encodeURIComponent(query)}&page=${page}`).json<AdminUser[]>()
		.catch(_ => [] as AdminUser[])
}

export const getUser = async (id: string) => {
	return api.get(`user/${id}`).json<AdminUser>().catch(_ => null)
}

export const deleteUser = async (id: string) => {
	return api.delete(`user/${id}`).json<{ success: boolean }>().catch(orError)
}

export const suspendUser = async (id: string, reason: string, duration: number | null = null) => {
	return api.post(`user/${id}/suspend`, {
		json: {
			reason,
			duration
		}
	}).json<{ success: boolean }>().catch(orError)
}

export const unsuspendUser = async (id: string) => {
	return api.post(`user/${id}/unsuspend`).json<{ success: boolean }>().catch(orError)
}

export const forceChangePassword = async (user: string, newPassword: string) => {
	return api.post(`user/${user}/password`, {
		json: {
			password: newPassword
		}
	}).json<{ success: boolean }>().catch(orError)
}

export const updateUser = (user: string, changes: UserChanges) => {
	return api.patch(`user/${user}`, {
		json: changes
	}).json<User>().catch(orError)
}

export const getAllApps = async () => {
	return api.get(`app`).json<AdminApplication[]>().catch(_ => [] as any[])
}

export const getApplication = async (id: string) => {
	return api.get(`app/${id}`).json<AdminApplication>().catch(_ => null)
}

export const createApplication = async (app: { name: string, slug?: string, url: string }) => {
	return api.post(`app`, { json: app }).json<AdminApplication>().catch(orError)
}

export const deleteApplication = async (app: string) => {
	return api.delete(`app/${app}`).json<{ success: boolean }>().catch(orError)
}

export const updateApplication = async (id: string, app: { name?: string, slug?: string, url?: string, callback_url?: string }) => {
	return api.patch(`app/${id}`, { json: app }).json<AdminApplication>().catch(orError)
}

export const createProvider = async (app: string, kind: string) => {
	return api.post(`app/${app}/provider`, { json: { kind } }).json<AdminProvider>().catch(orError)
}

export const deleteProvider = async (app: string, provider: string) => {
	return api.delete(`app/${app}/provider/${provider}`).json<{ success: boolean }>().catch(orError)
}

export const rotateProvider = async (app: string, provider: string) => {
	return api.post(`app/${app}/provider/${provider}/rotate`).json<{ secret: string }>().catch(orError)
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

export interface AppLogins {
	user_pass_login: boolean
	providers: string[]
}

export interface LoginResponse {
	refresh_token: string
	access_token: string
	expires_at: number
}

export interface UserChanges {
	username?: string
	given_name?: string
	family_name?: string
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
	roles: Role[]
	accounts: Account[]
	sessions: Session[]
}

export interface AdminUser {
	id: string
	created_at: number
	updated_at: number
	username: string
	first_name: string
	last_name: string
	email: string
	email_verified: boolean
	profile_picture: string
	last_login_at: number
	roles: Role[]
	accounts: Account[]
	sessions: Session[]
	suspension: Suspension | null
}

export interface Role {
	id: string
	created_at: number
	updated_at: number
	name: string
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
	providers: AdminProvider[]
}

export interface AdminProvider {
	id: string
	created_at: number
	updated_at: number
	application_id: string
	type: 'oauth2'
	secret: string
	redirect_uris: string[]
	redirect_uri_globs: string[]
	application_type: string
	auth_method: string
	response_types: string[]
	grant_types: string[]
	access_token_type: string
}