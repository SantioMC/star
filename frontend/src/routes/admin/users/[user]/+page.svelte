<script lang="ts">
	import { browser } from '$app/environment';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import * as star from '$lib/star';
	import Icon from '@/components/icon.svelte';
	import Modal from '@/components/modal.svelte';
	import RoleSelector from '@/components/role-selector.svelte';
	import Session from '@/components/session.svelte';
	import Sidebar from '@/components/sidebar.svelte';
	import SocialAccount from '@/components/social-account.svelte';
	import { Badge } from '@/components/ui/badge';
	import { Button } from '@/components/ui/button';
	import * as Card from '@/components/ui/card';
	import * as DropdownMenu from '@/components/ui/dropdown-menu';
	import { Input } from '@/components/ui/input';
	import * as Item from '@/components/ui/item';
	import { Label } from '@/components/ui/label';
	import { Spinner } from '@/components/ui/spinner';
	import { hasPermission } from '@/index';
	import { relativeTime, titleCase } from '@/utils';
	import { onMount } from 'svelte';
	import { toast } from 'svelte-sonner';

	const dateFormatter = new Intl.DateTimeFormat(undefined, {
		month: 'short',
		day: '2-digit',
		year: 'numeric'
	});

	let loading = $state<boolean>(false);

	let deleteUserModal = $state<Modal>();
	let suspendUserModal = $state<Modal>();
	let unsuspendUserModal = $state<Modal>();
	let changePasswordModal = $state<Modal>();
	let changeRolesModal = $state<Modal>();
	let verifyUserModal = $state<Modal>();
	let unverifyUserModal = $state<Modal>();
	let resendEmailModal = $state<Modal>();

	let reason = $state<string>('');
	let password = $state<string>('');
	let username = $state<string>('');
	let profilePicture = $state<string>('');
	let firstName = $state<string>('');
	let lastName = $state<string>('');
	let roles = $state<star.Role[]>([]);

	let logins = $state<star.AppLogins | null>(null);
	let user = $state<star.AdminUser | null>(null);
	let suspender = $state<star.AdminUser | null>(null);

	let metadata = $derived(
		!user ? '' : JSON.stringify(user!.metadata, null, 4).trim()
	);
	let sessions = $derived(
		user?.sessions?.toSorted((a, b) => b.created_at - a.created_at) ??
			([] as star.Session[])
	);

	onMount(async () => {
		if (!browser) return;
		if (!hasPermission('star:admin:read:users')) return goto('/');

		user = await star.getUser(page.params.user!);
		if (user == null) return goto('/admin/users');

		logins = await star.getLogins(star.APP_ID);

		username = user.username;
		profilePicture = user.profile_picture;
		firstName = user.first_name;
		lastName = user.last_name;
		roles = Array.from(user.roles);

		if (user.suspension?.issued_by) {
			suspender = await star.getUser(user.suspension.issued_by);
		}
	});

	const deleteUser = async () => {
		deleteUserModal?.close();
		const result = await star.deleteUser(user!.id);
		if ('success' in result) {
			goto('/admin/users');
		} else {
			toast.error(result.message);
		}
	};

	const suspendUser = async () => {
		suspendUserModal?.close();
		const result = await star.suspendUser(user!.id, reason);
		reason = '';

		if ('success' in result) {
			location.reload();
		} else {
			toast.error(result.message);
		}
	};

	const unsuspendUser = async () => {
		unsuspendUserModal?.close();
		const result = await star.unsuspendUser(user!.id);
		if ('success' in result) {
			location.reload();
		} else {
			toast.error(result.message);
		}
	};

	const verifyUser = async () => {
		verifyUserModal?.close();

		const result = await star.modifyUser(user!.id, {
			email_verified: true
		});

		if ('error' in result) {
			toast.error(result.message);
			return;
		}

		user!.email_verified = true;
		toast.success('Successfully verified user');
	};

	const unverifyUser = async () => {
		unverifyUserModal?.close();

		const result = await star.modifyUser(user!.id, {
			email_verified: false
		});

		if ('error' in result) {
			toast.error(result.message);
			return;
		}

		user!.email_verified = false;
		toast.success('Removed email verification from user');
	};

	const resendEmail = async () => {
		loading = true;
		const result = await star.resendVerification(user!.email);

		loading = false;
		resendEmailModal?.close();

		if ('error' in result) {
			toast.error(result.message);
			return;
		}

		toast.success('Successfully resent email verification');
	};

	const changePassword = async () => {
		changePasswordModal?.close();

		const result = await star.forceChangePassword(user!.id, password);
		password = '';

		if ('success' in result) {
			toast.success('Successfully changed password');
		} else {
			toast.error(result.message);
		}
	};

	const updateUser = async () => {
		const result = await star.modifyUser(user!.id, {
			last_name: lastName,
			first_name: firstName,
			username: username
		});

		if ('error' in result) {
			firstName = user!.first_name;
			lastName = user!.last_name;
			username = user!.username;
			toast.error(result.message);
		} else {
			user!.first_name = firstName;
			user!.last_name = lastName;
			user!.username = username;
			toast.success('Successfully updated user');
		}
	};

	const invalidateSession = (session: star.Session) => {
		user!.sessions = user!.sessions.filter((s) => s.id != session.id);
		toast.success('You have invalidated the session');
	};

	const invalidateAllSessions = async () => {
		const result = await star.invalidateUserSessions(user!.id);

		if ('error' in result) {
			toast.error(result.message);
		} else {
			user!.sessions = user!.sessions.map((session) => {
				session.expired = true;
				session.expires_at = Date.now(); // not perfect, but this is perfectly fine for UI
				return session;
			});

			toast.success('Successfully invalidated all user sessions');
		}
	};

	const updateRoles = async () => {
		loading = true;
		const result = await star.modifyUser(user!.id, {
			roles: roles.map((r) => r.id)
		});

		loading = false;
		if ('error' in result) {
			roles = user!.roles;
			toast.error(result.message);
		} else {
			user!.roles = result.roles;
			toast.success("Successfully updated the user's roles");
			changeRolesModal?.close();
		}
	};

	const getAccount = (id: string) => {
		return user!.accounts.find((acc) => acc.id == id);
	};

	const updateMetadata = async () => {
		let data: star.Metadata | null;
		try {
			data = JSON.parse(metadata) as star.Metadata;
		} catch {
			return toast.error('invalid json for metadata');
		}

		const result = await star.modifyUser(user!.id, {
			metadata: data
		});

		if ('error' in result) {
			return toast.error(result.message);
		}

		metadata = JSON.stringify(result.metadata, null, 4).trim();
		toast.success('Successfully updated metadata');
	};
</script>

<Sidebar class="md:p-8 max-w-6xl mx-auto">
	{#if user}
		<div class="flex flex-row items-center">
			<div>
				<h1 class="text-2xl font-semibold">User Management</h1>
				<h2 class="text-lg text-primary/60">
					Viewing profile for {user?.username}
				</h2>
			</div>

			<Button
				variant="ghost"
				class="ml-auto"
				onclick={() => goto('/admin/users')}
			>
				<Icon icon="chevron-left" />Go back
			</Button>
		</div>

		<div class="mt-24 flex flex-col gap-8">
			<div class="flex flex-col xl:flex-row items-center gap-4">
				<img
					src={user.profile_picture}
					alt="{user.username}'s profile picture"
					class="rounded-full w-16 h-16"
				/>
				<div
					class="flex flex-col items-center xl:items-start justify-center gap-2"
				>
					<h3 class="text-lg leading-4">
						{#if user.first_name}
							{user.first_name} {user.last_name}
						{:else}
							{user.username}
						{/if}
					</h3>
					<div
						class="flex flex-col xl:flex-row items-center xl:gap-3 text-sm text-primary/60 flex-wrap"
					>
						<span>
							<Icon icon="id" />
							{user.id}
						</span>
						<span>
							<Icon icon="mail" />
							{user.email}
						</span>
						<span>
							<Icon icon="calendar-event" />
							Joined {dateFormatter.format(new Date(user.created_at * 1000))}
						</span>
						<span>
							<Icon icon="lock-open" />
							{#if user.last_login_at}
								Last logged in {dateFormatter.format(
									new Date(user.last_login_at * 1000)
								)}
							{:else}
								Never logged in
							{/if}
						</span>
					</div>
					{#if user.roles.length > 0}
						<div class="flex flex-col xl:flex-row items-center gap-4">
							{#each user.roles as role (role.id)}
								<Badge variant="secondary">{titleCase(role.name)}</Badge>
							{/each}
						</div>
					{/if}
				</div>

				<div class="xl:ml-auto flex items-center gap-2">
					{#if hasPermission('star:admin:update:users')}
						<DropdownMenu.Root>
							<DropdownMenu.Trigger class="ml-auto">
								<Button variant="ghost" size="icon">
									<Icon icon="dots" class="text-2xl" />
								</Button>
							</DropdownMenu.Trigger>
							<DropdownMenu.Content>
								<DropdownMenu.Group>
									{#if user.email_verified}
										<DropdownMenu.Item
											onclick={() => unverifyUserModal?.open()}
										>
											<Icon icon="mail-cancel" />
											Unverify Email
										</DropdownMenu.Item>
									{:else}
										<DropdownMenu.Item onclick={() => verifyUserModal?.open()}>
											<Icon icon="mail-check" />
											Verify Email
										</DropdownMenu.Item>
										<DropdownMenu.Item onclick={() => resendEmailModal?.open()}>
											<Icon icon="mail" />
											Resend Verification
										</DropdownMenu.Item>
									{/if}

									<DropdownMenu.Item
										onclick={() => changePasswordModal?.open()}
									>
										<Icon icon="password" />
										Change Password
									</DropdownMenu.Item>
									<DropdownMenu.Item disabled>
										<Icon icon="photo" />
										Update Profile Picture
									</DropdownMenu.Item>
									<DropdownMenu.Item disabled>
										<Icon icon="spy" />
										Impersonate User
									</DropdownMenu.Item>
								</DropdownMenu.Group>

								<DropdownMenu.Separator />
								<DropdownMenu.Group>
									{#if !user.suspension}
										<DropdownMenu.Item
											class="text-orange-400"
											onclick={() => suspendUserModal?.open()}
										>
											<Icon icon="hammer" />
											Suspend User
										</DropdownMenu.Item>
									{:else}
										<DropdownMenu.Item
											class="text-orange-400"
											onclick={() => unsuspendUserModal?.open()}
										>
											<Icon icon="hammer-off" />
											Revoke Suspension
										</DropdownMenu.Item>
									{/if}
									{#if hasPermission('star:admin:delete:users')}
										<DropdownMenu.Item
											variant="destructive"
											onclick={() => deleteUserModal?.open()}
										>
											<Icon icon="trash" />
											Delete
										</DropdownMenu.Item>
									{/if}
								</DropdownMenu.Group>
							</DropdownMenu.Content>
						</DropdownMenu.Root>
					{/if}
				</div>
			</div>

			{#if user.suspension}
				<Card.Root class="w-full">
					<Card.Content
						class="flex flex-col overflow-hidden flex-wrap text-orange-100"
					>
						<Card.Title class="text-orange-200 mb-4">
							<Icon icon="gavel" /> User is suspended
						</Card.Title>
						<span>
							This user is currently suspended for the reason "{user.suspension
								.reason}"
						</span>

						<div class="my-4 flex flex-col">
							{#if user.suspension.expires_at}
								<span>
									<Icon icon="clock" />
									This suspension expires on {dateFormatter.format(
										user.suspension.expires_at * 1000
									)} (in {relativeTime(user.suspension.expires_at)})
								</span>
							{/if}

							{#if suspender}
								<span>
									<Icon icon="hammer" />
									This suspension was issued by {suspender?.username}
								</span>
							{/if}
						</div>

						{#if hasPermission('star:admin:update:users') || suspender}
							<div class="ml-auto flex items-center gap-2">
								{#if hasPermission('star:admin:update:users')}
									<Button
										onclick={() => unsuspendUserModal?.open()}
										class="bg-transparent text-orange-400 hover:bg-orange-400/30 hover:border-transparent hover:text-primary"
									>
										<Icon icon="hammer-off" />
										Revoke Suspension
									</Button>
								{/if}

								{#if suspender}
									<Button
										variant="outline"
										onclick={() =>
											(location.href = `/admin/users/${suspender!.id}`)}
									>
										View Issuer
									</Button>
								{/if}
							</div>
						{/if}
					</Card.Content>
				</Card.Root>
			{/if}

			{#if !user.email_verified}
				<Card.Root class="w-full">
					<Card.Content
						class="flex flex-col overflow-hidden flex-wrap text-red-100"
					>
						<Card.Title class="text-red-200 mb-4">
							<Icon icon="mail" /> Email not verified
						</Card.Title>
						<span>
							This user has not yet verified their email address, they need to
							do so before they are allowed to login.
						</span>

						{#if hasPermission('star:admin:update:users')}
							<div class="ml-auto flex items-center gap-2 mt-8">
								{#if hasPermission('star:admin:update:users')}
									<Button
										onclick={() => verifyUserModal?.open()}
										class="bg-transparent text-green-500 hover:bg-green-400/30 hover:border-transparent hover:text-primary"
									>
										<Icon icon="mail-check" />
										Manually Verify
									</Button>

									<Button
										variant="outline"
										class="text-green-100"
										onclick={() => resendEmailModal?.open()}
									>
										Resend Email
									</Button>
								{/if}
							</div>
						{/if}
					</Card.Content>
				</Card.Root>
			{/if}

			<Card.Root class="min-w-1/3 w-full md:w-auto">
				<Card.CardHeader>
					<Card.Title><Icon icon="gavel" /> Roles</Card.Title>
					{#if user.roles.length > 0}
						<Card.Description>
							The roles the user has, roles are typically used to permit access
							to resources.
						</Card.Description>
					{/if}
				</Card.CardHeader>
				<Card.Content class="flex flex-col gap-1">
					{#if user.roles.length == 0}
						<div class="flex flex-col gap-4 items-center w-full my-16">
							<span class="text-xl font-semibold leading-2">No Roles</span>
							<span class="text-primary/60 leading-2">
								This user currently has no assigned roles
							</span>
							<Button class="mt-1" onclick={() => changeRolesModal?.open()}>
								Add Role
							</Button>
						</div>
					{:else}
						<div class="flex gap-2 items-center">
							{#each user.roles as role (role.id)}
								<Item.Root variant="outline" class="w-full">
									<Item.Content class="flex flex-row items-center gap-2">
										<Icon class="text-xl" icon={'folder'} />
										<div class="flex flex-col">
											<p class="text-md flex items-center gap-2">
												{titleCase(role.name)}
											</p>
											<p
												class="text-primary/60 text-md flex items-center gap-2"
											>
												{role.description}
											</p>
										</div>
									</Item.Content>
								</Item.Root>
							{/each}
						</div>
					{/if}
				</Card.Content>
				{#if user.roles.length > 0}
					<Card.Footer>
						<Button class="ml-auto" onclick={() => changeRolesModal?.open()}>
							Modify Roles
						</Button>
					</Card.Footer>
				{/if}
			</Card.Root>

			<Card.Root class="min-w-1/3 w-full md:w-auto">
				<Card.CardHeader>
					<Card.Title><Icon icon="user" /> Profile Information</Card.Title>
				</Card.CardHeader>
				<Card.Content class="flex flex-col gap-8">
					<form class="flex flex-col gap-6">
						<div class="grid gap-2">
							<Label for="username">Username</Label>
							<Input id="username" type="text" required bind:value={username} />
						</div>

						<div class="grid gap-2">
							<Label for="email">Email</Label>
							<Input id="email" type="email" value={user.email} disabled />
						</div>

						<div class="flex gap-2 sm:gap-8">
							<div class="flex-1/2 gap-2 grid">
								<Label for="givenName">Given Name</Label>
								<Input id="givenName" type="text" bind:value={firstName} />
							</div>

							<div class="flex-1/2 gap-2 grid">
								<Label for="familyName">Family Name</Label>
								<Input id="familyName" type="text" bind:value={lastName} />
							</div>
						</div>
					</form>
				</Card.Content>
				<Card.Footer>
					<Button class="ml-auto" onclick={updateUser}>Update</Button>
				</Card.Footer>
			</Card.Root>

			<Card.Root class="min-w-1/3 w-full md:w-auto">
				<Card.CardHeader>
					<Card.Title><Icon icon="code" /> User Metadata</Card.Title>
					<Card.Description>
						View and modify the custom metadata attached to this user. The
						metadata must be valid JSON. The <code>public</code>
						key is visible to users, and is what's set in the
						<code>meta</code>
						claim in the access token.
					</Card.Description>
				</Card.CardHeader>
				<Card.Content>
					<textarea
						class="w-full! p-4! resize-none"
						rows={metadata.split('\n').length}
						bind:value={metadata}
					></textarea>
				</Card.Content>
				<Card.Footer>
					<Button class="ml-auto" onclick={updateMetadata}>Update</Button>
				</Card.Footer>
			</Card.Root>

			{#if logins && logins.providers.length > 0}
				<Card.Root class="min-w-1/3 w-full md:w-auto">
					<Card.CardHeader>
						<Card.Title>
							<Icon icon="plug-connected" /> Connected Accounts
						</Card.Title>
						<Card.Description class="max-w-xl">
							The social accounts connected to the user
						</Card.Description>
					</Card.CardHeader>
					<Card.Content class="flex flex-row gap-2 flex-wrap">
						{#each logins?.providers as provider}
							<SocialAccount {provider} account={getAccount(provider.id)} />
						{/each}
					</Card.Content>
				</Card.Root>
			{/if}

			<Card.Root class="min-w-1/3 w-full md:w-auto">
				<Card.CardHeader>
					<Card.Title><Icon icon="device-desktop" /> Sessions</Card.Title>
					<Card.Description class="max-w-xl">
						The different devices and sessions currently logged into the
						account, invalidating a session might take up to an hour to take
						effect.
					</Card.Description>
				</Card.CardHeader>
				<Card.Content class="flex flex-col gap-2">
					{#each sessions as session}
						<Session
							{session}
							oninvalidate={() => invalidateSession(session)}
						/>
					{/each}

					{#if sessions.length == 0}
						<div class="flex flex-col gap-4 items-center w-full my-24">
							<span class="text-xl font-semibold leading-2">No Sessions</span>
							<span class="text-primary/60 leading-2">
								This user currently has no sessions
							</span>
						</div>
					{/if}
				</Card.Content>
				{#if sessions.length != 0}
					<Card.Footer>
						<Button
							class="ml-auto"
							variant="destructive"
							onclick={invalidateAllSessions}
						>
							Sign out all sessions
						</Button>
					</Card.Footer>
				{/if}
			</Card.Root>
		</div>
	{/if}
</Sidebar>

<Modal bind:this={deleteUserModal} button class="xl:w-xl!">
	<h1 class="text-xl font-semibold">Are you sure?</h1>
	<h2 class="text-card-foreground/60 sm:max-w-10/12">
		This action can not be reversed! It is recommended to instead suspend a user
		over deleting.
	</h2>

	<div class="flex items-center gap-2 justify-end mt-4">
		<Button variant="secondary" onclick={() => deleteUserModal?.close()}>
			Cancel
		</Button>
		<Button variant="destructive" onclick={deleteUser}>Delete User</Button>
	</div>
</Modal>

<Modal bind:this={suspendUserModal} button class="xl:w-xl!">
	<h1 class="text-xl font-semibold">Suspend {user?.username}</h1>
	<h2 class="text-card-foreground/60 sm:max-w-10/12">
		Suspending a user will immediately invalidate all sessions and prevent the
		user from signing in again.
	</h2>

	<form onsubmit={suspendUser} class="mt-6">
		<div class="flex flex-col gap-6 w-full mx-auto">
			<div class="grid gap-2">
				<Label for="reason">Reason</Label>
				<Input
					id="username"
					type="text"
					placeholder="Suspicious account behaviour"
					bind:value={reason}
					required
				/>
			</div>
		</div>

		<div class="flex items-center gap-2 justify-end mt-6">
			<Button variant="secondary" onclick={() => suspendUserModal?.close()}>
				Cancel
			</Button>
			<Button variant="destructive" type="submit">Suspend User</Button>
		</div>
	</form>
</Modal>

<Modal bind:this={unsuspendUserModal} button class="xl:w-xl!">
	<h1 class="text-xl font-semibold">Revoke Suspension</h1>
	<h2 class="text-card-foreground/60 sm:max-w-10/12">
		Revoking {user?.username}'s suspension will allow them to continue using
		their account
	</h2>

	<form onsubmit={unsuspendUser} class="mt-6">
		<div class="flex items-center gap-2 justify-end mt-6">
			<Button variant="secondary" onclick={() => unsuspendUserModal?.close()}>
				Cancel
			</Button>
			<Button type="submit">Revoke Suspension</Button>
		</div>
	</form>
</Modal>

<Modal bind:this={verifyUserModal} button class="xl:w-xl!">
	<h1 class="text-xl font-semibold">Verify {user?.username}</h1>
	<h2 class="text-card-foreground/60 sm:max-w-10/12">
		Verifying this user manually will mark their email as verified allowing them
		to login without needing to verify their email themselves.
	</h2>

	<form onsubmit={verifyUser} class="mt-6">
		<div class="flex items-center gap-2 justify-end mt-6">
			<Button variant="secondary" onclick={() => verifyUserModal?.close()}>
				Cancel
			</Button>
			<Button type="submit">Verify User</Button>
		</div>
	</form>
</Modal>

<Modal bind:this={unverifyUserModal} button class="xl:w-xl!">
	<h1 class="text-xl font-semibold">Unverify {user?.username}</h1>
	<h2 class="text-card-foreground/60 sm:max-w-10/12">
		Removing the email verification from this user will not allow them to login
		without verifying their email themselves. This action will <b>not</b>
		automatically send an email to the user.
	</h2>

	<form onsubmit={unverifyUser} class="mt-6">
		<div class="flex items-center gap-2 justify-end mt-6">
			<Button variant="secondary" onclick={() => unverifyUserModal?.close()}>
				Cancel
			</Button>
			<Button type="submit">Remove Verification</Button>
		</div>
	</form>
</Modal>

<Modal bind:this={resendEmailModal} button class="xl:w-xl!">
	<h1 class="text-xl font-semibold">Send email verification</h1>
	<h2 class="text-card-foreground/60 sm:max-w-10/12">
		Manually resend the email verification email to this user. This will allow
		them to verify their email address by clicking on the link in the email.
	</h2>

	<form onsubmit={resendEmail} class="mt-6">
		<div class="flex items-center gap-2 justify-end mt-6">
			<Button variant="secondary" onclick={() => resendEmailModal?.close()}>
				Cancel
			</Button>
			<Button type="submit" disabled={loading}>
				{#if loading}
					<Spinner />
				{/if}

				Resend Email
			</Button>
		</div>
	</form>
</Modal>

<Modal bind:this={changePasswordModal} button class="xl:w-xl!">
	<h1 class="text-xl font-semibold">Change Password</h1>
	<h2 class="text-card-foreground/60 sm:max-w-10/12">
		Manually change the password of this user, invalidating their previous
		sessions and requiring them to sign in with the new password
	</h2>

	<form onsubmit={changePassword} class="mt-6">
		<div class="flex flex-col gap-6 w-full mx-auto">
			<div class="grid gap-2">
				<Label for="password">New Password</Label>
				<Input
					id="password"
					type="password"
					placeholder=""
					bind:value={password}
					required
				/>
			</div>
		</div>

		<div class="flex items-center gap-2 justify-end mt-6">
			<Button variant="secondary" onclick={() => changePasswordModal?.close()}>
				Cancel
			</Button>
			<Button type="submit">Change Password</Button>
		</div>
	</form>
</Modal>

<Modal bind:this={changeRolesModal} button class="xl:w-xl!">
	<h1 class="text-xl font-semibold">Change Roles</h1>
	<h2 class="text-card-foreground/60 sm:max-w-10/12">
		Add or remove roles from this user
	</h2>

	<form onsubmit={updateRoles} class="mt-6">
		<RoleSelector bind:roles />

		<div class="flex items-center gap-2 justify-end mt-6">
			<Button variant="secondary" onclick={() => changeRolesModal?.close()}>
				Cancel
			</Button>
			<Button disabled={loading} type="submit">
				{#if loading}
					<Spinner />
				{/if}
				Apply Changes
			</Button>
		</div>
	</form>
</Modal>
