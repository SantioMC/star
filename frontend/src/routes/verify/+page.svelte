<script lang="ts">
	import { browser } from '$app/environment';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import * as star from '$lib/star';
	import Button from '@/components/ui/button/button.svelte';
	import * as Card from '@/components/ui/card';
	import { Spinner } from '@/components/ui/spinner';
	import { onMount } from 'svelte';

	let loading = $state(true);
	let error = $state<string | null>(null);

	onMount(async () => {
		if (!browser) return;

		const email = page.url.searchParams.get('email');
		const code = page.url.searchParams.get('code');

		if (!email || !code) {
			error = 'Invalid verification link provided';
			return;
		}

		const data = await star.verifyEmail(email, code);
		if ('error' in data) {
			error = data.message;
			return;
		}

		loading = false;
		localStorage.setItem(
			'star-login-note',
			'Your account has been verified, you may now login'
		);

		goto('/login');
	});
</script>

<div class="h-full w-full flex items-center justify-center">
	<Card.Root class="-my-4 w-full max-w-sm">
		{#if error}
			<Card.Header>
				<Card.Title>Failed to verify your account</Card.Title>
			</Card.Header>
			<Card.Content class="text-red-300 text-center flex flex-col gap-8">
				<span>{error}</span>
				<Button onclick={() => goto('/login')}>Go to Login</Button>
			</Card.Content>
		{:else if loading}
			<Card.Header>
				<Card.Title>Verifying your account</Card.Title>
				<Card.Description>
					One moment while we verify your account...
				</Card.Description>
			</Card.Header>
			<Card.Content>
				<div class="flex items-center justify-center h-48 w-full">
					<Spinner size="2rem" />
				</div>
			</Card.Content>
		{/if}
	</Card.Root>
</div>
