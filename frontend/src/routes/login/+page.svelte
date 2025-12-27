<script lang="ts">
	import * as star from '$lib/star';
	import Icon from '@/components/icon.svelte';
	import Button from '@/components/ui/button/button.svelte';
	import * as Card from '@/components/ui/card';
	import { Input } from '@/components/ui/input';
	import { Label } from '@/components/ui/label';
	import { Spinner } from '@/components/ui/spinner';
	import { refreshToken, token } from '@/index';
	import { knownProviders, titleCase } from '@/utils';
	import { onMount } from 'svelte';

	let logins = $state<star.AppLogins | null>(null);

	let loading = $state(false);
	let error = $state('')

	let username = $state('');
	let password = $state('');

	const login = async (event: Event) => {
		event.preventDefault();
		error = '';
		
		const data = await star.login(username, password);
		
		if ('error' in data) {
			error = data.message 
		} else {
			token.set(data.access_token)
			refreshToken.set(data.refresh_token)
			location.href = "/"
		}
	}

	onMount(async () => {
		logins = await star.getLogins();
	});
</script>

<div class="h-full w-full flex items-center justify-center">
	<form onsubmit={login}>
		<Card.Root class="-my-4 w-full max-w-sm">
			<Card.Header>
				<Card.Title>Login to your account</Card.Title>
				<Card.Description>
					Enter your account details below to login to your account
				</Card.Description>
			</Card.Header>
			{#if !logins}
				<div class="flex items-center justify-center h-48 w-full">
					<Spinner size='2rem' />
				</div>
			{:else if logins?.user_pass_login}
				<Card.Content class='w-96'>
					<div class="flex flex-col gap-6">
						<div class="grid gap-2">
							<Label for="username">Username</Label>
							<Input bind:value={username} id="username" type="text" required />
						</div>
						<div class="grid gap-2">
							<div class="flex items-center">
								<Label for="password">Password</Label>
							</div>
							<Input bind:value={password} id="password" type="password" required />
						</div>
					</div>

					{#if error}
						<span class="text-red-300">{error}</span>
					{/if}
				</Card.Content>
			{/if}
			<Card.Footer class="flex-col gap-2">
				{#if logins?.user_pass_login}
					<Button type="submit" class="w-full" disabled={loading} onclick={login}>
						{#if loading}
							<Spinner />
						{/if}
						Login
					</Button>
				{/if}

				{#if logins?.user_pass_login && logins?.providers?.length > 0}
					<hr />
				{/if}

				{#each logins?.providers as provider}
					<Button type="submit" variant='outline' class="w-full" onclick={() => star.loginWithProvider(provider)}>
						{#if !!knownProviders[provider]}
							<Icon icon={knownProviders[provider].icon} class='text-lg' />
						{/if}
						Login with {titleCase(provider)}
					</Button>
				{/each}
			</Card.Footer>
		</Card.Root>
	</form>
</div>
