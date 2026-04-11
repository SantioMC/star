<script lang="ts">
	import { browser } from '$app/environment';
	import { page } from '$app/state';
	import * as Card from '@/components/ui/card';
	import { Spinner } from '@/components/ui/spinner';
	import { continueFlow, refreshToken, token } from '@/index';
	import { onMount } from 'svelte';

	onMount(() => {
		if (!browser) return;

		token.set(page.url.searchParams.get('access_token'));
		refreshToken.set(page.url.searchParams.get('refresh_token'));

		continueFlow();
	});
</script>

<div class="h-full w-full flex items-center justify-center">
	<Card.Root class="-my-4 w-full max-w-sm">
		<Card.Header>
			<Card.Title>Getting you signed in</Card.Title>
			<Card.Description>One moment while we sign you in...</Card.Description>
		</Card.Header>
		<Card.Content>
			<div class="flex items-center justify-center h-48 w-full">
				<Spinner size="2rem" />
			</div>
		</Card.Content>
	</Card.Root>
</div>
