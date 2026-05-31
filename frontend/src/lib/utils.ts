import type { Session } from "@/star";
import { clsx, type ClassValue } from "clsx";
import ms from "ms";
import type { ChangeEventHandler } from "svelte/elements";
import { twMerge } from "tailwind-merge";

let debounceTask: number | null = null

export const dateFormatter = new Intl.DateTimeFormat(undefined, {
	month: 'short',
	day: 'numeric',
	year: 'numeric'
});

const timeFormatter = new Intl.DateTimeFormat(undefined, {
	month: 'short',
	day: 'numeric',
	year: 'numeric',
	hour: 'numeric',
	minute: '2-digit',
	hour12: true
});

export function formatTime(time: number | Date): string {
	const date = time instanceof Date ? time : new Date(time * 1000)
	const parts = timeFormatter.formatToParts(date)
	let year = false;

	for (const part of parts) {
		if (part.type == 'dayPeriod') part.value = part.value.replaceAll('.', '');
		if (part.type == 'year') year = true
		if (part.type == 'literal' && year) {
			part.value = ` at `
			year = false
		}
	}

	return parts.map(p => p.value).join('')
}

export function cn(...inputs: ClassValue[]) {
	return twMerge(clsx(inputs));
}

export function titleCase(text: string): string {
	// just some edge cases, we don't care a whole lot tbh
	if (text == 'oauth2') return 'OAuth2';
	if (text.length == 1) return text.toUpperCase()

	return text.charAt(0).toUpperCase() + text.substring(1).toLowerCase();
}

export function relativeTime(time: number | string, long: boolean = true): string {
	const epoch = typeof time === 'string' ? Date.parse(time) : time * 1000
	const diff = Math.abs(Date.now() - epoch)
	if (diff < 1e3) return "a few moments"

	return ms(diff, { long });
}

export const isExpired = (session: Session): boolean => Date.parse(session.expires_at) < Date.now()

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