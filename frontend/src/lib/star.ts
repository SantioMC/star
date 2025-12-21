// Super simple library for working with the Star API

import { env } from "$env/dynamic/public";
import ky, { HTTPError } from "ky";

const API_URL = env.PUBLIC_API_URL ?? "/api";
const APP_ID = "star";

export const api = ky.extend({
	prefixUrl: API_URL,
	retry: {
		limit: 3,
		retryOnTimeout: true,
		jitter: delay => delay + (Math.random() * 200 - 100)
	}
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
	return api.post(`app/${APP_ID}/login`, { json: { username, password } }).json<LoginResponse>()
		.catch(orError)
}

export const loginWithProvider = (provider: string) => {
	location.href = `${API_URL}/app/${APP_ID}/login/${provider}`
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