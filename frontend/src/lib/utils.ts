import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";

export const knownProviders = {
	"discord": {
		icon: "brand-discord",
		color: "#5865F2"
	},
	"twitch": {
		icon: "brand-twitch",
		color: "#9146FF"
	},
	"google": {
		icon: "brand-google",
		color: "#4285f4"
	},
	"slack": {
		icon: "brand-slack",
		color: "#4A154B"
	},
	"github": {
		icon: "brand-github",
		color: "#222222"
	}
} as Record<string, { icon: string, color: string }>

export function cn(...inputs: ClassValue[]) {
	return twMerge(clsx(inputs));
}

export function titleCase(text: string): string {
	return text.charAt(0).toUpperCase() + text.substring(1).toLowerCase()
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
export type WithoutChild<T> = T extends { child?: any } ? Omit<T, "child"> : T;
// eslint-disable-next-line @typescript-eslint/no-explicit-any
export type WithoutChildren<T> = T extends { children?: any } ? Omit<T, "children"> : T;
export type WithoutChildrenOrChild<T> = WithoutChildren<WithoutChild<T>>;
export type WithElementRef<T, U extends HTMLElement = HTMLElement> = T & { ref?: U | null };