import type { Session } from "@/star";
import { clsx, type ClassValue } from "clsx";
import ms from "ms";
import type { ChangeEventHandler } from "svelte/elements";
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

let debounceTask: number | null = null

export function cn(...inputs: ClassValue[]) {
	return twMerge(clsx(inputs));
}

export function titleCase(text: string): string {
	return text.charAt(0).toUpperCase() + text.substring(1).toLowerCase();
}

export function relativeTime(time: number, long: boolean = true): string {
	return ms(Math.abs(Date.now() - (time * 1000)), { long });
}

export const debounced = <T extends EventTarget>(func: ChangeEventHandler<T>, delay: number = 250) => (event: Event & { currentTarget: Element & T }) => {
	if (debounceTask != null) clearTimeout(debounceTask);
	debounceTask = setTimeout(() => {
		func(event)
	}, delay)
}

export const getIcon = (session: Session): string => {
	switch (session.device.toLowerCase()) {
		case "desktop":
			return "device-desktop"
		case "mobile":
			return "device-mobile"
		case "tablet":
			return "device-tablet"
		case "tv":
			return "device-tv"
		case "bot":
			return "robot"
		default:
			return "device-unknown"
	}
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
export type WithoutChild<T> = T extends { child?: any } ? Omit<T, "child"> : T;
// eslint-disable-next-line @typescript-eslint/no-explicit-any
export type WithoutChildren<T> = T extends { children?: any } ? Omit<T, "children"> : T;
export type WithoutChildrenOrChild<T> = WithoutChildren<WithoutChild<T>>;
export type WithElementRef<T, U extends HTMLElement = HTMLElement> = T & { ref?: U | null };