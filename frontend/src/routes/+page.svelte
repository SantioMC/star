<script lang="ts">
	import { browser } from '$app/environment';
	import { goto } from '$app/navigation';
	import Email from '$lib/components/email.svelte';
	import Icon from '$lib/components/icon.svelte';
	import Copy from '@/components/copy.svelte';
	import Modal from '@/components/modal.svelte';
	import Session from '@/components/session.svelte';
	import Sidebar from '@/components/sidebar.svelte';
	import TotpInput from '@/components/totp-input.svelte';
	import * as Avatar from '@/components/ui/avatar';
	import { Badge } from '@/components/ui/badge';
	import * as Item from '@/components/ui/item';
	import { Button } from '@/components/ui/button';
	import * as Card from '@/components/ui/card';
	import { Input } from '@/components/ui/input';
	import { Label } from '@/components/ui/label';
	import { Spinner } from '@/components/ui/spinner';
	import { getJWT, hasAdminAccess, hasPermission } from '@/index';
	import * as star from '@/star';
	import { isExpired } from '@/utils';
	import { onMount } from 'svelte';
	import { toast } from 'svelte-sonner';
	import { renderSVG } from 'uqr';

	let data = $state<star.User | null>(null);
	let sessions = $state<star.Session[]>([]);
	let emails = $state<star.Email[]>([]);
	let loading = $state<boolean>(false);

	let username = $state<string>('');
	let firstName = $state<string>('');
	let lastName = $state<string>('');

	let newEmail = $state<string>('');
	let addingEmail = $state<boolean>(false);

	let currentPassword = $state<string>('');
	let newPassword = $state<string>('');
	let confirmPassword = $state<string>('');

	let totpEnabled = $state<boolean | null>(null);
	let totpSecret = $state<string | null>(null);
	let totpCode = $state<string>('');
	let showManualKey = $state<boolean>(false);
	let enrollingTotp = $state<boolean>(false);
	let totpModal = $state<Modal>();
	let totpModalOpen = $state<boolean>(false);
	let disableTotpModal = $state<Modal>();

	const totpQr = $derived(totpSecret ? renderSVG(totpSecret) : null);
	const totpManualKey = $derived.by(() => {
		if (!totpSecret) return null;

		try {
			return new URL(totpSecret).searchParams.get('secret');
		} catch {
			return null;
		}
	});

	const sortedEmails = $derived(
		[...emails].sort((a, b) => {
			if (a.primary !== b.primary) return a.primary ? -1 : 1;
			return Date.parse(a.created_at) - Date.parse(b.created_at);
		})
	);

	onMount(async () => {
		if (!browser) return;

		data = await star.getSelf();
		if (data == null) return goto('/login');

		username = data.username;
		firstName = data.first_name;
		lastName = data.last_name;

		sessions = await star.getSessions();
		emails = await star.getEmails();

		const factors = await star.getMfaFactors();
		totpEnabled = factors == null ? null : factors.factors.includes('totp');
	});

	$effect(() => {
		if (!totpModalOpen) {
			totpSecret = null;
			totpCode = '';
			showManualKey = false;
		}
	});

	const removeSession = (session: star.Session) => {
		sessions = sessions.filter((s) => s.id != session.id);
	};

	const addEmail = async (e: Event) => {
		e.preventDefault();

		const email = newEmail.trim();
		if (!email) return;

		addingEmail = true;
		const response = await star.addEmail(email);
		addingEmail = false;

		if ('error' in response) {
			return toast.error(response.message);
		}

		emails = [...emails, response];
		newEmail = '';
		toast.success(`A verification email has been sent to ${response.email}`);
	};

	const onPrimaryEmail = (email: star.Email) => {
		emails = emails.map((e) => ({ ...e, primary: e.id == email.id }));
		toast.success(`${email.email} is now your primary email`);
	};

	const onDeleteEmail = (email: star.Email) => {
		emails = emails.filter((e) => e.id != email.id);
		toast.success(`Removed ${email.email} from your account`);
	};

	const invalidateAllSessions = async () => {
		loading = true;
		const response = await star.invalidateAllSessions();

		if ('error' in response) {
			loading = false;
			return toast.error(response.message);
		}

		goto('/logout');
	};

	const changePassword = async (e: Event) => {
		e.preventDefault();

		if (!currentPassword || !newPassword || !confirmPassword) {
			return;
		}

		if (newPassword !== confirmPassword) {
			return toast.error('Your passwords do not match!');
		}

		loading = true;
		const response = await star.changePassword(currentPassword, newPassword);

		if ('error' in response) {
			loading = false;
			return toast.error(response.message);
		}

		goto('/logout');
	};

	const startTotpEnrollment = async () => {
		if (enrollingTotp) return;

		enrollingTotp = true;
		const response = await star.enrollTotp();
		enrollingTotp = false;

		if ('error' in response) {
			return toast.error(response.message);
		}

		totpSecret = response.secret;
		totpCode = '';
		totpModal?.open();
	};

	const confirmTotpEnrollment = async (e: Event | null = null) => {
		e?.preventDefault();

		const code = totpCode.trim();
		if (code.length < 6 || loading) return;

		loading = true;
		const response = await star.verifyTotp(code);
		loading = false;

		if ('error' in response) {
			totpCode = '';
			return toast.error(response.message);
		}

		totpEnabled = true;
		totpModal?.close();
		toast.success('Two-factor authentication is now enabled');
	};

	const disableTotp = async () => {
		if (loading) return;

		loading = true;
		const response = await star.deleteTotp();
		loading = false;
		disableTotpModal?.close();

		if ('error' in response) {
			return toast.error(response.message);
		}

		totpEnabled = false;
		toast.success('Two-factor authentication has been disabled');
	};

	const updateDetails = async (e: Event) => {
		e.preventDefault();

		loading = true;
		const response = await star.updateUser({
			username,
			first_name: firstName,
			last_name: lastName
		});

		loading = false;

		if ('error' in response) {
			return toast.error(response.message);
		} else {
			toast.success('Successfully updated your details!');
		}
	};
</script>

<Sidebar
	hidden={data == null ||
		!hasAdminAccess()}
>
	{#if data != null}
		<div class="flex flex-col md:w-3xl mx-auto py-12 px-4 gap-8">
			<div class="flex items-center">
				<div>
					<h1 class="leading-8 text-2xl font-semibold">Manage your account</h1>
					<h2 class="leading-6 text-primary/60 text-sm mt-1">
						Manage your account settings and preferences
					</h2>
				</div>

				{#if hasPermission('admin:user:read')}
					<Button
						onclick={() => goto(`/admin/users/${data!.id}`)}
						variant="secondary"
						class="ml-auto"
					>
						View Admin Page
					</Button>
				{/if}
			</div>

			<Card.Root class="min-w-1/3 w-full md:w-auto">
				<Card.CardHeader>
					<Card.Title><Icon icon="user" /> Profile Information</Card.Title>
				</Card.CardHeader>

				<form onsubmit={updateDetails}>
					<Card.Content class="flex flex-col gap-8">
						<div class="flex gap-3 items-center">
							<Avatar.Root class="size-12">
								<Avatar.Image
									src={data.profile_picture}
									alt="{username}'s Profile Picture"
								/>
								<Avatar.Fallback class="text-lg">
									{username.charAt(0).toUpperCase()}
								</Avatar.Fallback>
							</Avatar.Root>
							<div class="flex flex-col gap-0.5">
								<span class="text-base font-medium">
									{#if firstName && lastName}
										{firstName} {lastName}
									{:else}
										{username}
									{/if}
								</span>
								<span class="text-xs text-primary/50">
									Not you? <a class="text-blue-300/80 hover:text-blue-300 transition-colors" href="/logout">
										Log out
									</a>
								</span>
							</div>
						</div>

						<div class="flex flex-col gap-5">
							<div class="grid gap-2">
								<Label for="username">Username</Label>
								<Input
									bind:value={username}
									id="username"
									type="text"
									required
								/>
							</div>

							<div class="flex gap-2 sm:gap-6">
								<div class="flex-1/2 gap-2 grid">
									<Label for="givenName">Given Name</Label>
									<Input bind:value={firstName} id="givenName" type="text" />
								</div>

								<div class="flex-1/2 gap-2 grid">
									<Label for="familyName">Family Name</Label>
									<Input bind:value={lastName} id="familyName" type="text" />
								</div>
							</div>
						</div>
					</Card.Content>
					<Card.Footer class="mt-4">
						<Button type="submit" class="ml-auto" onclick={updateDetails}>
							Update
						</Button>
					</Card.Footer>
				</form>
			</Card.Root>

			<Card.Root class="min-w-1/3 w-full md:w-auto">
				<Card.CardHeader>
					<Card.Title><Icon icon="mail" /> Email Addresses</Card.Title>
					<Card.Description class="max-w-xl">
						The email addresses associated with your account. Your primary email
						is used for account notifications and signing in.
					</Card.Description>
				</Card.CardHeader>
				<Card.Content class="flex flex-col gap-2">
					{#each sortedEmails as email (email.id)}
						<Email
							{email}
							canDelete={!email.primary && emails.length > 1}
							onprimary={() => onPrimaryEmail(email)}
							ondelete={() => onDeleteEmail(email)}
						/>
					{/each}

					{#if emails.length === 0}
						<p class="text-sm text-primary/50 text-center py-4">No email addresses found</p>
					{/if}
				</Card.Content>
				<Card.Footer>
					<form class="flex w-full items-end gap-2" onsubmit={addEmail}>
						<div class="grid gap-2 flex-1">
							<Label for="newEmail">Add a new email</Label>
							<Input
								bind:value={newEmail}
								id="newEmail"
								type="email"
								placeholder="you@example.com"
								required
							/>
						</div>
						<Button type="submit" disabled={addingEmail || !newEmail.trim()}>
							{#if addingEmail}
								<Spinner />
							{:else}
								<Icon icon="plus" />
							{/if}

							Add Email
						</Button>
					</form>
				</Card.Footer>
			</Card.Root>

			<Card.Root class="min-w-1/3 w-full md:w-auto" id="change-password">
				<Card.CardHeader>
					<Card.Title><Icon icon="lock" /> Change Password</Card.Title>
				</Card.CardHeader>
				<form onsubmit={changePassword}>
					<Card.Content class="flex flex-col gap-5">
						<div class="grid gap-2">
							<Label for="currentPassword">Current Password</Label>
							<Input
								bind:value={currentPassword}
								id="currentPassword"
								type="password"
								required
								placeholder={'••••••••••••'}
							/>
						</div>

						<div class="grid gap-2">
							<Label for="password">New Password</Label>
							<Input bind:value={newPassword} id="password" type="password" />
						</div>

						<div class="grid gap-2">
							<Label for="confirmPassword">Confirm Password</Label>
							<Input
								bind:value={confirmPassword}
								id="confirmPassword"
								type="password"
							/>
						</div>
					</Card.Content>
					<Card.Footer class="mt-4">
						<Button class="ml-auto" onclick={changePassword}>
							Change Password
						</Button>
					</Card.Footer>
				</form>
			</Card.Root>

			<Card.Root class="min-w-1/3 w-full md:w-auto">
				<Card.CardHeader>
					<Card.Title><Icon icon="shield-lock" /> Two-Factor Authentication</Card.Title>
					<Card.Description class="max-w-xl">
						Add another step of verification to your login, increasing your account security.
					</Card.Description>
				</Card.CardHeader>
				<Card.Content>
					<Item.Root variant="outline">
						<Item.Content class="flex flex-row items-center gap-4">
							<Icon class="text-2xl" icon="device-mobile" />
							<div class="min-w-0">
								<p class="text-base flex items-center gap-2 flex-wrap">
									<span>Authenticator app</span>

									{#if totpEnabled == null}
										<Badge variant="outline">Unknown</Badge>
									{:else if totpEnabled}
										<Badge variant="outline" class="border-emerald-600/60 text-emerald-500">Enabled</Badge>
									{:else}
										<Badge variant="outline" class="border-amber-500/60 text-amber-400">Disabled</Badge>
									{/if}
								</p>

								<Item.Description class="text-xs">
									Time-based one-time codes (TOTP)
								</Item.Description>
							</div>

							<div class="ml-auto flex items-center gap-1.5">
								{#if totpEnabled == true && hasPermission('user:mfa:delete')}
									<Button
										variant="destructive"
										size="sm"
										onclick={() => disableTotpModal?.open()}
										disabled={loading}
									>
										Disable
									</Button>
								{:else if totpEnabled == false && hasPermission('user:mfa:enroll')}
									<Button
										variant="secondary"
										size="sm"
										onclick={startTotpEnrollment}
										disabled={enrollingTotp}
									>
										{#if enrollingTotp}
											<Spinner />
										{/if}

										Enable
									</Button>
								{/if}
							</div>
						</Item.Content>
					</Item.Root>
				</Card.Content>
			</Card.Root>

			<Card.Root class="min-w-1/3 w-full md:w-auto">
				<Card.CardHeader>
					<Card.Title><Icon icon="device-desktop" /> Sessions</Card.Title>
					<Card.Description class="max-w-xl">
						The different devices and sessions currently logged into your
						account. Invalidating a session might take up to an hour to take
						effect.
					</Card.Description>
				</Card.CardHeader>
				<Card.Content class="flex flex-col gap-2">
					{#each sessions
						.filter((i) => !isExpired(i))
						.sort((a, b) => Date.parse(b.created_at) - Date.parse(a.created_at)) as session (session.id)}
						<Session
							{session}
							oninvalidate={() => removeSession(session)}
							currentSession={session.id == getJWT()?.sid}
						/>
					{/each}

					{#if sessions.filter((i) => !isExpired(i)).length === 0}
						<p class="text-sm text-primary/50 text-center py-4">No active sessions found</p>
					{/if}
				</Card.Content>
				<Card.Footer>
					<Button
					class="ml-auto"
						variant="destructive"
						onclick={invalidateAllSessions}
						disabled={loading}
					>
						{#if loading}
							<Spinner />
						{/if}

						Sign out all sessions
					</Button>
				</Card.Footer>
			</Card.Root>
		</div>
	{/if}
</Sidebar>

<Modal bind:this={totpModal} bind:showModal={totpModalOpen} button class="xl:w-xl!">
	<h1 class="text-xl font-semibold">Set up two-factor authentication</h1>
	<h2 class="text-card-foreground/60 sm:max-w-10/12">
		Scan the QR code below with your authenticator app, then enter the 6-digit
		code it generates to confirm the setup.
	</h2>

	<div class="flex flex-col items-center gap-4 mt-6">
		{#if totpQr}
			<div class="w-48 rounded-lg bg-white p-3">
				{@html totpQr}
			</div>
		{/if}

		{#if totpManualKey}
			<div class="flex flex-col items-center gap-1">
				{#if showManualKey}
					<span class="text-xs text-card-foreground/60">
						Enter this key manually in your authenticator app
					</span>
					<code class="text-sm break-all text-center flex items-center gap-1">
						{totpManualKey}
						<Copy content={totpManualKey} />
					</code>
				{:else}
					<span class="text-xs text-card-foreground/60">
						Can't scan the code?
						<button
							type="button"
							class="link"
							onclick={() => (showManualKey = true)}
						>
							Click here
						</button>
					</span>
				{/if}
			</div>
		{/if}
	</div>

	<form
		class="flex flex-col items-center gap-4 mt-6"
		onsubmit={confirmTotpEnrollment}
	>
		<div class="grid gap-3 justify-items-center">
			<Label>Authentication code</Label>
			<TotpInput
				bind:value={totpCode}
				disabled={loading}
			/>
		</div>

		<Button
			type="submit"
			class="w-full mt-8"
			disabled={loading || totpCode.trim().length < 6}
		>
			{#if loading}
				<Spinner />
			{/if}

			Verify
		</Button>
	</form>
</Modal>

<Modal bind:this={disableTotpModal} button class="xl:w-xl!">
	<h1 class="text-xl font-semibold">Disable two-factor authentication</h1>
	<h2 class="text-card-foreground/60 sm:max-w-10/12">
		Are you sure you want to disable two-factor authentication? Your account
		will no longer require a one-time code when signing in, making it less
		secure.
	</h2>

	<div class="flex items-center gap-2 justify-end mt-4">
		<Button variant="secondary" onclick={() => disableTotpModal?.close()}>
			Cancel
		</Button>
		<Button variant="destructive" onclick={disableTotp} disabled={loading}>
			{#if loading}
				<Spinner />
			{/if}

			Disable
		</Button>
	</div>
</Modal>
