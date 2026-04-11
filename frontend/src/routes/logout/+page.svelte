<script lang="ts">
	import { browser } from '$app/environment';
	import * as star from '$lib/star';
	import * as Card from '@/components/ui/card';
	import { Spinner } from '@/components/ui/spinner';
	import { getJWT, refreshToken, token } from '@/index';
	import { onMount } from 'svelte';
	import { get } from 'svelte/store';

	onMount(async () => {
		if (!browser) return

		const currentToken = get(token);
		if (currentToken != null) {
			const data = await getJWT(currentToken)
			if (data) await star.invalidateSession(data.sid)
		}

		token.set(null)
		refreshToken.set(null)

		location.href = '/';
	})
</script>

<div class="h-full w-full flex items-center justify-center">
	<Card.Root class="-my-4 w-full max-w-sm">
		<Card.Header>
			<Card.Title>Signing out</Card.Title>
			<Card.Description>
				One moment while we sign you out...
			</Card.Description>
		</Card.Header>
		<Card.Content>
			<div class="flex items-center justify-center h-48 w-full">
				<Spinner size='2rem' />
			</div>
		</Card.Content>
	</Card.Root>
</div>
