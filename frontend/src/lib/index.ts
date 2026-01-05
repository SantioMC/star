import { browser } from "$app/environment";
import type { StarPermission } from "@/star";
import { get, writable, type Writable } from "svelte/store";

const persistentWritable = (key: string): Writable<string|null> => {
	if (!browser) return writable<string|null>(null);

	const store = writable<string|null>(localStorage.getItem(key))
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
export const hasPermission = (permission: StarPermission | (string & {})) => {
	const jwt = getJWT();
	if (!jwt) return false;

	for (const scope of jwt.scopes) {
		if (scope === permission) {
			return true;
		}

		if (scope.endsWith('*')) {
			if (permission.startsWith(scope.slice(0, -1))) {
				return true;
			}
		}
	}

	return false;
}

token.subscribe(newValue => {
	const data = getJWT(newValue)
	if (!data) return

	// todo: refresh token
})

export interface JWTDetails {
	sub: string
	sid: string
	aud: string
	name: string
	nickname: string
	given_name: string
	family_name: string
	picture: string
	email: string
	email_verified: boolean
	exp: number
	iat: number
	iss: string
	naf: number
	nbf: number
	scopes: string[]
	roles: [
		{
			id: string,
			name: string,
			super: boolean
		}
	]
}