<script lang="ts">
	import { browser } from '$app/environment';
	import { goto } from '$app/navigation';
	import Icon from '$lib/components/icon.svelte';
	import Session from '@/components/session.svelte';
	import Sidebar from '@/components/sidebar.svelte';
	import * as Avatar from '@/components/ui/avatar';
	import { Button } from '@/components/ui/button';
	import * as Card from '@/components/ui/card';
	import { Input } from '@/components/ui/input';
	import { Label } from '@/components/ui/label';
	import { Spinner } from '@/components/ui/spinner';
	import { getJWT, hasPermission } from '@/index';
	import * as star from '@/star';
	import { onMount } from 'svelte';
	import { toast } from 'svelte-sonner';

	let data = $state<star.User | null>(null);
	let sessions = $state<star.Session[]>([]);
	let loading = $state<boolean>(false);

	let username = $state<string>('');
	let firstName = $state<string>('');
	let lastName = $state<string>('');

	let currentPassword = $state<string>('');
	let newPassword = $state<string>('');
	let confirmPassword = $state<string>('');

	onMount(async () => {
		if (!browser) return;

		data = await star.getSelf();
		if (data == null) return goto('/login');

		username = data.username;
		firstName = data.first_name;
		lastName = data.last_name;

		sessions = await star.getSessions();
	});

	const removeSession = (session: star.Session) => {
		sessions = sessions.filter((s) => s.id != session.id);
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
		(!hasPermission('star:server') && !hasPermission('star:admin'))}
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

				{#if hasPermission('star:admin:read:users')}
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

							<div class="grid gap-2">
								<Label for="email">Email</Label>
								<Input id="email" type="email" value={data.email} disabled />
								{#if !data.email_verified}
									<span class="text-xs text-amber-400">Your email is not yet verified</span>
								{/if}
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
								placeholder={'•••���••••••••••••'}
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
					<Card.Title><Icon icon="device-desktop" /> Sessions</Card.Title>
					<Card.Description class="max-w-xl">
						The different devices and sessions currently logged into your
						account. Invalidating a session might take up to an hour to take
						effect.
					</Card.Description>
				</Card.CardHeader>
				<Card.Content class="flex flex-col gap-2">
					{#each sessions
						.filter((i) => !i.expired)
						.sort((a, b) => b.created_at - a.created_at) as session}
						<Session
							{session}
							oninvalidate={() => removeSession(session)}
							currentSession={session.id == getJWT()?.sid}
						/>
					{/each}

					{#if sessions.filter((i) => !i.expired).length === 0}
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
