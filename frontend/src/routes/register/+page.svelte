<script lang="ts">
	import { goto } from '$app/navigation';
	import { env } from '$env/dynamic/public';
	import * as star from '$lib/star';
	import Captcha from '@/components/captcha.svelte';
	import Button from '@/components/ui/button/button.svelte';
	import * as Card from '@/components/ui/card';
	import { Input } from '@/components/ui/input';
	import { Label } from '@/components/ui/label';
	import { Spinner } from '@/components/ui/spinner';
	import { continueFlow, refreshToken, token } from '@/index';
	import { onMount } from 'svelte';
	import { toast } from 'svelte-sonner';

	let logins = $state<star.AppLogins | null>(null);

	let loading = $state(false);
	let promptCaptcha = $state(false);
	let error = $state('');

	let email = $state('');
	let username = $state('');
	let password = $state('');
	let confirmPassword = $state('');

	let captchaToken = $state('');
	let captchaError = $state('');

	const register = async (event: Event | null = null) => {
		if (env.PUBLIC_SITE_KEY && !captchaToken) {
			promptCaptcha = true;
			return;
		}

		event?.preventDefault();
		error = '';

		if (password !== confirmPassword) {
			return (error = 'Your passwords do not match!');
		}

		const data = await star.register({
			application_id: star.APP_ID,
			captcha: captchaToken ? captchaToken : undefined,
			username,
			email,
			password
		});

		if ('error' in data) {
			captchaToken = '';
			error = data.message;
			return;
		}

		if (data.refresh_token && data.access_token) {
			token.set(data.access_token);
			refreshToken.set(data.refresh_token);
			return continueFlow();
		}

		toast.success('Successfully created your account!');
		await goto('/login');
	};

	const captchaCallback = async (token: string) => {
		captchaToken = token;
		await register();
		promptCaptcha = false;
	};

	onMount(async () => {
		logins = await star.getLogins(star.APP_ID);
		if (!logins?.user_pass_login) return goto('/login');
	});
</script>

<div class="h-full w-full flex items-center justify-center">
	<form onsubmit={register} class="w-full sm:w-fit">
		<Card.Root class="-my-4 w-full sm:w-sm">
			{#if promptCaptcha}
				<Card.Header>
					<Card.Title>Solve the Captcha</Card.Title>
					<Card.Description>
						Please wait or manually solve the provided captcha to prove you are
						a human
					</Card.Description>
				</Card.Header>
				<Card.Content class="w-96 flex items-center flex-col gap-2">
					<Captcha
						onsuccess={captchaCallback}
						onerror={(e) => (captchaError = e)}
					/>
					{#if captchaError}
						<span class="text-red-400 text-center">{captchaError}</span>
					{/if}
				</Card.Content>
			{:else}
				<Card.Header>
					<Card.Title>Create an account</Card.Title>
					<Card.Description>
						Enter your account details below to create your account
					</Card.Description>
				</Card.Header>
				{#if !logins}
					<div class="flex items-center justify-center h-48 w-full">
						<Spinner size="2rem" />
					</div>
				{:else}
					<Card.Content class="w-96">
						<div class="flex flex-col gap-4">
							<div class="grid gap-2">
								<Label for="username">Username</Label>
								<Input
									bind:value={username}
									id="username"
									type="text"
									required
								/>
							</div>
							<div class="grid gap-2">
								<Label for="email">Email</Label>
								<Input bind:value={email} id="email" type="email" required />
							</div>
							<div class="grid gap-2">
								<Label for="password">Password</Label>
								<Input
									bind:value={password}
									id="password"
									type="password"
									required
								/>
							</div>
							<div class="grid gap-2">
								<Label for="confirm-password">Confirm Password</Label>
								<Input
									bind:value={confirmPassword}
									id="confirm-password"
									type="password"
									required
								/>
							</div>
						</div>

						{#if error}
							<span class="text-red-300">{error}</span>
						{/if}
					</Card.Content>
				{/if}
				<Card.Footer class="flex-col gap-2">
					<Button
						type="submit"
						class="w-full"
						disabled={loading}
						onclick={register}
					>
						{#if loading}
							<Spinner />
						{/if}
						Create Account
					</Button>

					<div class="flex items-center gap-1 text-sm text-primary/60">
						Already have an account?
						<a href="/login" class="link">Log in</a>
					</div>
				</Card.Footer>
			{/if}
		</Card.Root>
	</form>
</div>
