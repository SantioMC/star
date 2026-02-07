<script lang="ts">
	import { env } from '$env/dynamic/public';
	import { onMount } from 'svelte';

	let captcha = $state<HTMLElement>();

	const {
		onsuccess,
		onerror
	}: {
		onsuccess?: (token: string) => void;
		onerror?: (error: string) => void;
	} = $props();

	onMount(() => {
		turnstile.render(captcha!, {
			sitekey: env.PUBLIC_SITE_KEY,
			theme: 'auto',
			callback: function (token: string) {
				console.log('Captcha solved!', token);
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
	});
</script>

<div bind:this={captcha} class="flex items-center justify-center w-full"></div>
