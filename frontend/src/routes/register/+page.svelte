<script lang="ts">
	import { env } from '$env/dynamic/public';
	import * as star from '$lib/star';
	import Captcha from '@/components/captcha.svelte';
	import Icon from '@/components/icon.svelte';
	import Button from '@/components/ui/button/button.svelte';
	import * as Card from '@/components/ui/card';
	import { Input } from '@/components/ui/input';
	import { Label } from '@/components/ui/label';
	import { Spinner } from '@/components/ui/spinner';
	import { onMount } from 'svelte';
	import { toast } from 'svelte-sonner';

	let registered = $state(false);
	let registeredEmail = $state('');

	let loading = $state(false);
	let promptCaptcha = $state(false);
	let error = $state('');

	let countdown = $state<number>(0);
	let countdownInterval = $state<number | undefined>();

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

		loading = true;

		const data = await star.register({
			captcha: captchaToken ? captchaToken : undefined,
			username,
			email,
			password
		});

		loading = false;

		if ('error' in data) {
			captchaToken = '';
			error = data.message;
			return;
		}

		resetCountdown();
		registeredEmail = email;
		registered = true;
	};

	const resetCountdown = () => {
		countdown = 60;

		if (countdownInterval) {
			clearInterval(countdownInterval);
		}

		countdownInterval = setInterval(() => {
			if (--countdown <= 0) {
				clearInterval(countdownInterval);
				countdownInterval = undefined;
			}
		}, 1000);
	};

	const captchaCallback = async (token: string) => {
		captchaToken = token;
		await register();
		promptCaptcha = false;
	};

	const resendEmail = async () => {
		if (!registeredEmail) return;

		loading = true;
		const data = await star.resendVerification(registeredEmail);

		resetCountdown();
		loading = false;

		if ('error' in data) {
			toast.error(data.message);
			return;
		}

		if (!data.success) {
			toast.error('Failed to resend the verification email');
			return;
		}

		toast.success('The email has been resent');
	};
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
			{:else if registered}
				<Card.Header>
					<Card.Title>
						<div class="flex items-center gap-2 h-4">
							<Icon icon="mail" class="text-xl" />
							<span>Verify your email</span>
						</div>
					</Card.Title>
				</Card.Header>
				<Card.Content class="w-96 flex items-center flex-col gap-8">
					<p class="text-primary/80">
						We sent a verification email to <strong>{registeredEmail}</strong>.
						Please check your inbox and spam folder. It may take a minute to arrive.
					</p>
					<Button
						onclick={resendEmail}
						disabled={countdown > 0 || loading}
						class="ml-auto"
					>
						{#if countdown > 0}
							Resend email ({countdown}s)
						{:else if loading}
							<Spinner />
							Resend email
						{:else}
							Resend email
						{/if}
					</Button>
				</Card.Content>
			{:else}
				<Card.Header>
					<Card.Title>Create an account</Card.Title>
					<Card.Description>
						Enter your account details below to create your account
					</Card.Description>
				</Card.Header>
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
						<div class="flex w-full items-center justify-center my-2">
							<span class="text-red-300 text-center">{error}</span>
						</div>
					{/if}
				</Card.Content>
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
