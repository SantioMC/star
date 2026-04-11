<script lang="ts">
	import { browser } from '$app/environment';
	import { page } from '$app/state';
	import favicon from '$lib/assets/favicon.svg';
	import { Toaster } from '@/components/ui/sonner';
	import { isLoggedIn } from '@/index';
	import './layout.css';

	let { children } = $props();

	const PUBLIC_ROUTES = [
		'/login',
		'/register',
		'/verify',
		'/callback',
		'/oauth/authorize'
	];

	if (browser && !PUBLIC_ROUTES.includes(page.url.pathname) && !isLoggedIn()) {
		location.href = '/login';
	}
</script>

<svelte:head><link rel="icon" href={favicon} /></svelte:head>
<Toaster duration={2000} position="bottom-right" closeButton={true} />

{@render children()}
