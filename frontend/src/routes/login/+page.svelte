<script lang="ts">
	import { env } from '$env/dynamic/public';
	import * as star from '$lib/star';
	import Captcha from '@/components/captcha.svelte';
	import Icon from '@/components/icon.svelte';
	import Button from '@/components/ui/button/button.svelte';
	import * as Card from '@/components/ui/card';
	import { Input } from '@/components/ui/input';
	import { Label } from '@/components/ui/label';
	import { Separator } from '@/components/ui/separator';
	import { Spinner } from '@/components/ui/spinner';
	import { continueFlow, refreshToken, token } from '@/index';
	import { onMount } from 'svelte';

	let notice = $state<string | null>(null);

	let loading = $state(false);
	let promptCaptcha = $state(false);
	let error = $state('');

	let username = $state('');
	let password = $state('');

	let captchaToken = $state('');
	let captchaError = $state('');

	const login = async (event: Event | null = null) => {
		if (env.PUBLIC_SITE_KEY && !captchaToken) {
			promptCaptcha = true;
			return;
		}

		event?.preventDefault();
		error = '';
		loading = true;

		const data = await star.login({
			captcha: captchaToken ? captchaToken : undefined,
			username,
			password
		});

		loading = false;

		if ('error' in data) {
			captchaToken = '';
			error = data.message;
			return;
		}

		token.set(data.access_token);
		refreshToken.set(data.refresh_token);

		continueFlow();
	};

	const captchaCallback = async (token: string) => {
		captchaToken = token;
		await login();
		promptCaptcha = false;
	};

	onMount(() => {
		notice = localStorage.getItem('star-login-note');

		if (notice) {
			localStorage.removeItem('star-login-note');
		}
	});
</script>

<div class="h-full w-full flex items-center justify-center">
	<form
		onsubmit={login}
		class="w-full sm:w-fit flex flex-col gap-4 items-center"
	>
		{#if notice}
			<div class="mb-4 flex items-center gap-2 rounded-lg border border-blue-400/30 bg-blue-500/10 px-4 py-3 text-sm text-blue-200">
				<Icon icon="alert-circle" class="text-lg shrink-0" />
				<span>{notice}</span>
			</div>
		{/if}

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
					<Card.Title>Login to your account</Card.Title>
					<Card.Description>
						Enter your account details below to login to your account
					</Card.Description>
				</Card.Header>
				<Card.Content class="w-96">
					<div class="flex flex-col gap-6">
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
							<div class="flex items-center">
								<Label for="password">Password</Label>
							</div>
							<Input
								bind:value={password}
								id="password"
								type="password"
								required
							/>
						</div>
					</div>

					{#if error}
						<span class="text-red-300 mt-2 block">{error}</span>
					{/if}
				</Card.Content>
				<Card.Footer class="flex-col gap-2">
					<Button
						type="submit"
						class="w-full"
						disabled={loading}
						onclick={login}
					>
						{#if loading}
							<Spinner />
						{/if}
						Login
					</Button>

					<div class="flex items-center gap-1 text-sm text-primary/60">
						New user?
						<a href="/register" class="link">Create an account</a>
					</div>
				</Card.Footer>
			{/if}
		</Card.Root>

		{#if env.PUBLIC_TOS || env.PUBLIC_SITE_KEY}
			<div
				class="flex w-full justify-center h-5 gap-4 mt-8 items-center text-sm text-primary/70"
			>
				{#if env.PUBLIC_TOS}
					<a
						href={env.PUBLIC_TOS}
						class="transition-colors cursor-pointer hover:text-primary"
					>
						Terms of Service
					</a>

					{#if env.PUBLIC_SITE_KEY}
						<Separator orientation="vertical" />
					{/if}
				{/if}
				{#if env.PUBLIC_PRIVACY}
					<a
						href={env.PUBLIC_PRIVACY}
						class="transition-colors cursor-pointer hover:text-primary"
					>
						Privacy Policy
					</a>
				{/if}
			</div>
		{/if}
	</form>
</div>
