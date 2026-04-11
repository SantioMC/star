<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import * as star from '$lib/star';
	import { Button } from '@/components/ui/button';
	import * as Card from '@/components/ui/card';
	import { Spinner } from '@/components/ui/spinner';
	import { isLoggedIn } from '@/index';
	import { onMount } from 'svelte';

	let error = $state<string | null>(null);
	let loading = $state<boolean>(true);
	let loggingIn = $state<boolean>(false);
	let application = $state<star.PublicApplication | null>(null);

	onMount(async () => {
		error = page.url.searchParams.get('error');

		const clientId = page.url.searchParams.get('client_id');
		if (!clientId) return;

		if (!isLoggedIn()) return requestLogin();

		application = await star.getOAuthApplication(clientId);
		if (application?.implicit) {
			return submit();
		}

		loading = false;
	});

	const requestLogin = async () => {
		localStorage.setItem(
			'star-auth-redirect',
			page.url.searchParams.toString()
		);

		return goto('/login');
	};

	const submit = async () => {
		if (!!error || !application) {
			history.back();
			return;
		}

		loading = false;
		loggingIn = true;

		const redirectUri = page.url.searchParams.get('redirect_uri');
		const scope = page.url.searchParams.get('scope') ?? 'openid';

		if (!redirectUri) {
			error = 'Missing redirect_uri parameter';
			loggingIn = false;
			return;
		}

		const authCode = await star.getOAuthAuthCode(
			application.id,
			scope,
			redirectUri,
			page.url.searchParams.get('state'),
			page.url.searchParams.get('code_challenge'),
			page.url.searchParams.get('code_challenge_method'),
			page.url.searchParams.get('nonce'),
		);

		if ('error' in authCode) {
			if (authCode.message === 'Forbidden') {
				return requestLogin();
			}

			error = authCode.message;
			loading = false;
			loggingIn = false;
			return;
		}

		const url = new URL(redirectUri);
		url.searchParams.set('code', authCode.code);

		if (authCode.state) {
			url.searchParams.set('state', authCode.state);
		}

		window.location.href = url.toString();
	};
</script>

<div class="h-full w-full flex items-center justify-center">
	<form onsubmit={submit} class="w-full sm:w-fit">
		<Card.Root class="-my-4 w-full sm:w-sm">
			{#if error}
				<Card.Header>
					<Card.Title>An error has occurred</Card.Title>
					<Card.Description>
						Failed to authorize your request, please try again or contact the
						site administrator if the problem persists.
					</Card.Description>
				</Card.Header>
				<Card.Content
					class="w-96 flex items-center flex-col gap-2 text-red-300"
				>
					{error}
				</Card.Content>
				<Card.Footer class="flex-col gap-2">
					<Button type="submit" class="w-full">Go back</Button>
				</Card.Footer>
			{:else if loading}
				<Card.Header>
					<Card.Title>One Moment...</Card.Title>
					<Card.Description>
						Loading application information, please wait...
					</Card.Description>
				</Card.Header>
				<Card.Content class="w-96 flex items-center flex-col gap-2">
					<Spinner class="my-8" />
				</Card.Content>
			{:else if !application}
				<Card.Header>
					<Card.Title>Failed to get application</Card.Title>
					<Card.Description>
						Failed to get the application information, please try again or
						contact the site administrator if the problem persists.
					</Card.Description>
				</Card.Header>
				<Card.Footer class="flex-col gap-2">
					<Button type="submit" class="w-full">Go back</Button>
				</Card.Footer>
			{:else if loggingIn}
				<Card.Header>
					<Card.Title>One Moment...</Card.Title>
					<Card.Description>
						You are being logged in, please wait...
					</Card.Description>
				</Card.Header>
				<Card.Content class="w-96 flex items-center flex-col gap-2">
					<Spinner class="my-8" />
				</Card.Content>
			{:else}
				<Card.Header>
					<Card.Title>Login to {application.name}</Card.Title>
					<Card.Description>
						Click on the 'Continue' button to continue logging in to {application.name}.
					</Card.Description>
				</Card.Header>
				<Card.Footer class="flex-col gap-2">
					<Button class="w-full" onclick={submit}>Continue</Button>
				</Card.Footer>
			{/if}
		</Card.Root>
	</form>
</div>
