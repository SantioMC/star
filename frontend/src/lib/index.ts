import { goto } from "$app/navigation";
import * as star from "$lib/star";
import { get, writable } from "svelte/store";
import { persistentStore, typedPersistentStore } from "./store";

export const token = writable<string | null>(null);
export const user = typedPersistentStore<star.User | null>("star_user", null);

const signedIn = persistentStore("star_signed_in");

export const isLoggedIn = (): boolean => {
  return get(token) != null;
};

export const startSession = (accessToken: string) => {
  token.set(accessToken);
  signedIn.set("true");
};

export const clearSession = () => {
  token.set(null);
  signedIn.set(null);
  user.set(null);
};

export const restoreSession = async () => {
  if (get(signedIn) == null) return;
  await refreshAccessToken();
};

export type LoginReason = "authorize" | "reauthenticate";

export interface LoginIntent {
  returnTo: string;
  reason?: LoginReason;
  label?: string;
}

const isSafePath = (path: string): boolean =>
  path.startsWith("/") && !path.startsWith("//") && !path.startsWith("/\\");

export const getLoginIntent = (): LoginIntent | null => {
  const stored = localStorage.getItem("star-auth-redirect");
  if (!stored) return null;

  try {
    const intent = JSON.parse(stored) as LoginIntent;
    return isSafePath(intent.returnTo) ? intent : null;
  } catch {
    return null;
  }
};

export const requestLogin = (intent: LoginIntent) => {
  localStorage.setItem("star-auth-redirect", JSON.stringify(intent));
  return goto("/login");
};

export const continueFlow = async () => {
  const intent = getLoginIntent();
  localStorage.removeItem("star-auth-redirect");

  await goto(intent?.returnTo ?? "/");
};

const base64UrlDecode = (str: string): string => {
  str = str.replace(/-/g, "+").replace(/_/g, "/");
  while (str.length % 4) {
    str += "=";
  }

  return atob(str);
};

export const getJWT = (
  accessToken: string | null = get(token),
): JWTDetails | null => {
  if (!accessToken) return null;

  const parts = accessToken.split(".");
  if (parts.length != 3) return null;

  // No reason for us to bother verifying it, the frontend doesn't do anything special
  return JSON.parse(base64UrlDecode(parts[1])) as JWTDetails;
};

// amazing hack to take in any string, but also still have intellisense on the arg, idek
export const hasPermission = (
  permission: star.StarPermission | (string & {}),
  accessToken: string | null = get(token),
) => {
  if (!accessToken) return false;

  const jwt = getJWT(accessToken);
  if (!jwt) return false;

  const permissions = jwt["urn:star:iam:permissions"]?.split(" ") ?? [];
  return permissions.includes(permission);
};

const adminPermissions: star.StarPermission[] = [
  "admin:user:read",
  "admin:role:read",
  "application:read",
  "admin:provider:read",
  "admin:settings:read",
  "admin:key:read",
  "admin:token:read",
  "admin:metrics:read",
  "admin:alert:read",
];

export const hasAdminAccess = (accessToken: string | null = get(token)) => {
  return adminPermissions.some((permission) =>
    hasPermission(permission, accessToken),
  );
};

let refreshing: Promise<string | null> | null = null;

export const refreshAccessToken = (): Promise<string | null> => {
  if (refreshing) return refreshing;

  const pending = lockedRefresh().finally(() => {
    refreshing = null;
  });

  refreshing = pending;
  return pending;
};

const lockedRefresh = async (): Promise<string | null> => {
  if (!navigator.locks) return requestAccessToken();
  return await navigator.locks.request("star-session", requestAccessToken);
};

const requestAccessToken = async (): Promise<string | null> => {
  try {
    const result = await star.refreshSession();
    if ("error" in result) {
      console.error(`Failed to refresh access token: ${result.message}`);
      clearSession();

      return null;
    }

    token.set(result.access_token);
    return result.access_token;
  } catch (error) {
    console.error(
      "Failed to reach the server to refresh the access token",
      error,
    );
    return null;
  }
};

export interface JWTDetails {
  sub: string;
  sid: string;
  aud: string[];
  exp: number;
  iat: number;
  iss: string;
  naf: number;
  nbf: number;
  scope?: string;
  imp?: string;
  "urn:star:iam:permissions": string;
}
