import { browser } from "$app/environment";
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

export const getJWT = (): JWTDetails | null => {
	const accessToken = get(token);
	if (!accessToken) return null;

	const parts = accessToken.split('.');
	if (parts.length != 3) return null;

	// No reason for us to bother verifying it, the frontend doesn't do anything special
	return JSON.parse(base64UrlDecode(parts[1])) as JWTDetails
}

export interface JWTDetails {
	sub: string
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
	roles: [
		{
			id: string,
			name: string,
			super: boolean
		}
	]
}