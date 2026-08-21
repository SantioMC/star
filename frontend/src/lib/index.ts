import { browser } from "$app/environment";
import { goto } from "$app/navigation";
import * as star from "$lib/star";
import { get, writable, type Writable } from "svelte/store";

const persistentWritable = (key: string): Writable<string | null> => {
	if (!browser) return writable<string | null>(null);

	const store = writable<string | null>(localStorage.getItem(key))
	store.subscribe(newValue => {
		if (!newValue) localStorage.removeItem(key)
		else localStorage.setItem(key, newValue)
	})

	return store
}

export const token = persistentWritable('access_token')
export const refreshToken = persistentWritable('refresh_token')

export const isLoggedIn = (): boolean => {
	return get(token) != null && get(refreshToken) != null
}

// Continues the login, register, or oauth flow
export const continueFlow = async () => {
	const redirect = localStorage.getItem('star-auth-redirect');
	if (!redirect) return goto('/');

	localStorage.removeItem('star-auth-redirect');
	await goto('/oauth/authorize?' + redirect);
}

const base64UrlDecode = (str: string): string => {
	str = str.replace(/-/g, '+').replace(/_/g, '/');
	while (str.length % 4) {
		str += '=';
	}

	return atob(str);
}

export const getJWT = (accessToken: string | null = get(token)): JWTDetails | null => {
	if (!accessToken) return null;

	const parts = accessToken.split('.');
	if (parts.length != 3) return null;

	// No reason for us to bother verifying it, the frontend doesn't do anything special
	return JSON.parse(base64UrlDecode(parts[1])) as JWTDetails
}

// amazing hack to take in any string, but also still have intellisense on the arg, idek
export const hasPermission = (permission: star.StarPermission | (string & {}), accessToken: string | null = get(token)) => {
	if (!accessToken) return false;

	const jwt = getJWT(accessToken);
	if (!jwt) return false;

	for (const perm of jwt.permissions ?? []) {
		if (perm === permission) {
			return true;
		}

		if (perm.endsWith('*')) {
			if (permission.startsWith(perm.slice(0, -1))) {
				return true;
			}
		}
	}

	return false;
}

const adminPermissions: star.StarPermission[] = [
	'admin:user:read',
	'admin:role:read',
	'application:read',
	'admin:provider:read',
	'admin:settings:read',
]

export const hasAdminAccess = (accessToken: string | null = get(token)) => {
	return adminPermissions.some((permission) => hasPermission(permission, accessToken))
}

export const refreshAccessToken = async (): Promise<string | null> => {
	const value = get(refreshToken)
	if (!value) {
		console.error("No refresh token found, unable to refresh access token");
		return null
	}

	const result = await star.refreshToken(value)
	if ('error' in result) {
		console.error(`Failed to refresh access token (${result.error}): ${result.error_description ?? 'No description provided'}`)

		token.set(null)
		refreshToken.set(null)

		return null
	}

	token.set(result.access_token)
	if (result.refresh_token) {
		refreshToken.set(result.refresh_token)
	}

	return result.access_token
}

export interface JWTDetails {
	sub: string
	sid: string
	aud: string[]
	exp: number
	iat: number
	iss: string
	naf: number
	nbf: number
	scope?: string
	permissions: string[]
}