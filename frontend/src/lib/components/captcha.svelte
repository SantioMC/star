<script lang="ts" module>
	let loader: Promise<void> | null = null;

	const loadTurnstile = () => {
		loader ??= new Promise<void>((resolve, reject) => {
			const script = document.createElement('script');
			script.src = 'https://challenges.cloudflare.com/turnstile/v0/api.js?render=explicit';
			script.async = true;
			script.onload = () => resolve();
			script.onerror = () => {
				loader = null;
				script.remove();
				reject(new Error('Failed to load the captcha'));
			};

			document.head.append(script);
		});

		return loader;
	};
</script>

<script lang="ts">
	import { env } from '$env/dynamic/public';
	import { onMount } from 'svelte';

	let captcha = $state<HTMLElement>();

	const {
		sitekey = env.PUBLIC_SITE_KEY,
		onsuccess,
		onerror
	}: {
		sitekey?: string;
		onsuccess?: (token: string) => void;
		onerror?: (error: string) => void;
	} = $props();

	onMount(() => {
		let widget: string | null | undefined = null;
		let destroyed = false;

		loadTurnstile()
			.then(() => {
				if (destroyed) return;

				widget = turnstile.render(captcha!, {
					sitekey,
					theme: 'auto',
					callback: function (token: string) {
						onsuccess?.(token);
					},
					'error-callback': function (errorCode: any) {
						onerror?.('Failed to solve captcha, error code: ' + errorCode);
					},
					'expired-callback': function () {
						onerror?.('The captcha has expired!');
					},
					'timeout-callback': function () {
						onerror?.('The captcha has timed out!');
					}
				});
			})
			.catch((error: Error) => onerror?.(error.message));

		return () => {
			destroyed = true;
			if (widget) turnstile.remove(widget);
		};
	});
</script>

<div bind:this={captcha} class="flex items-center justify-center w-full"></div>
