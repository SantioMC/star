<script lang="ts">
	import { browser } from '$app/environment';
	import * as Card from '@/components/ui/card';
	import { Spinner } from '@/components/ui/spinner';
	import { refreshToken, token } from '@/index';
	import { onMount } from 'svelte';

	onMount(() => {
		if (!browser) return

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
