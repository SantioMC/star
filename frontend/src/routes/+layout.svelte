<script lang="ts">
	import { browser } from '$app/environment';
	import { page } from '$app/state';
	import favicon from '$lib/assets/favicon.svg';
	import { isLoggedIn } from '@/index';
	import './layout.css';

	let { children } = $props();

	const PUBLIC_ROUTES = ['/login', '/register', '/callback']

	if (browser && !PUBLIC_ROUTES.includes(page.url.pathname) && !isLoggedIn()) {
		location.href = "/login"
	}
</script>

<svelte:head><link rel="icon" href={favicon} /></svelte:head>
{@render children()}
