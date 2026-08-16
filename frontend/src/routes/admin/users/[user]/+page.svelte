<script lang="ts">
	import { browser } from '$app/environment';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import * as star from '$lib/star';
	import Icon from '@/components/icon.svelte';
	import Modal from '@/components/modal.svelte';
	import RoleSelector from '@/components/role-selector.svelte';
	import Sidebar from '@/components/sidebar.svelte';
	import { Badge } from '@/components/ui/badge';
	import { Button } from '@/components/ui/button';
	import * as Card from '@/components/ui/card';
	import * as DropdownMenu from '@/components/ui/dropdown-menu';
	import { Input } from '@/components/ui/input';
	import * as Item from '@/components/ui/item';
	import { Label } from '@/components/ui/label';
	import { Spinner } from '@/components/ui/spinner';
	import { hasPermission } from '@/index';
	import { titleCase } from '@/utils';
	import { onMount } from 'svelte';
	import { toast } from 'svelte-sonner';

	const dateFormatter = new Intl.DateTimeFormat(undefined, {
		month: 'short',
		day: '2-digit',
		year: 'numeric'
	});

	let loading = $state<boolean>(false);

	let deleteUserModal = $state<Modal>();
	let changeRolesModal = $state<Modal>();
	let signOutModal = $state<Modal>();
	let setPasswordModal = $state<Modal>();
	let setPasswordModalOpen = $state<boolean>(false);

	let newPassword = $state<string>('');
	let confirmPassword = $state<string>('');

	let username = $state<string>('');
	let firstName = $state<string>('');
	let lastName = $state<string>('');
	let roles = $state<star.Role[]>([]);

	let user = $state<star.AdminUser | null>(null);

	onMount(async () => {
		if (!browser) return;
		if (!hasPermission('admin:user:read')) return goto('/');

		user = await star.getUser(page.params.user!);
		if (user == null) return goto('/admin/users');

		username = user.username;
		firstName = user.first_name ?? '';
		lastName = user.last_name ?? '';
		roles = Array.from(user.roles);
	});

	$effect(() => {
		if (!setPasswordModalOpen) {
			newPassword = '';
			confirmPassword = '';
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

	const updateUser = async () => {
		const result = await star.modifyUser(user!.id, {
			username: username,
			first_name: firstName,
			last_name: lastName
		});

		if ('error' in result) {
			username = user!.username;
			firstName = user!.first_name ?? '';
			lastName = user!.last_name ?? '';
			toast.error(result.message);
		} else {
			user = result;
			toast.success('Successfully updated user');
		}
	};

	const invalidateAllSessions = async () => {
		signOutModal?.close();
		const result = await star.invalidateUserSessions(user!.id);

		if ('error' in result) {
			toast.error(result.message);
		} else {
			toast.success('Successfully invalidated all user sessions');
		}
	};

	const setPassword = async (e: Event) => {
		e.preventDefault();

		if (!newPassword || !confirmPassword) return;

		if (newPassword !== confirmPassword) {
			return toast.error('The passwords do not match!');
		}

		loading = true;
		const result = await star.adminSetPassword(user!.id, newPassword);
		loading = false;

		if ('error' in result) {
			return toast.error(result.message);
		}

		setPasswordModal?.close();
		toast.success("Successfully changed the user's password");
	};

	const updateRoles = async (e: Event | null = null) => {
		e?.preventDefault();

		loading = true;
		const result = await star.modifyUser(user!.id, {
			roles: roles.map((r) => r.id)
		});

		loading = false;
		if ('error' in result) {
			roles = user!.roles;
			toast.error(result.message);
		} else {
			user = result;
			roles = Array.from(result.roles);
			toast.success("Successfully updated the user's roles");
			changeRolesModal?.close();
		}
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
							{user.email ?? 'No email'}
						</span>
						<span>
							<Icon icon="calendar-event" />
							Joined {dateFormatter.format(new Date(user.created_at))}
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
					{#if hasPermission('admin:user:write') || hasPermission('admin:user:delete')}
						<DropdownMenu.Root>
							<DropdownMenu.Trigger class="ml-auto">
								<Button variant="ghost" size="icon">
									<Icon icon="dots" class="text-2xl" />
								</Button>
							</DropdownMenu.Trigger>
							<DropdownMenu.Content>
								<DropdownMenu.Group>
									{#if hasPermission('admin:user:write')}
										<DropdownMenu.Item onclick={() => setPasswordModal?.open()}>
											<Icon icon="key" />
											Set password
										</DropdownMenu.Item>
									{/if}
									{#if hasPermission('admin:user:delete')}
										<DropdownMenu.Item onclick={() => signOutModal?.open()}>
											<Icon icon="logout" />
											Sign out all sessions
										</DropdownMenu.Item>
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

			{#if !user.email_verified}
				<Card.Root class="w-full">
					<Card.Content
						class="flex flex-col overflow-hidden flex-wrap text-red-100"
					>
						<Card.Title class="text-red-200 mb-4">
							<Icon icon="mail" /> Email not verified
						</Card.Title>
						<span>
							This user has not yet verified their email address.
						</span>
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
							{#if hasPermission('admin:user:write')}
								<Button class="mt-1" onclick={() => changeRolesModal?.open()}>
									Add Role
								</Button>
							{/if}
						</div>
					{:else}
						<div class="flex gap-2 items-center flex-col">
							{#each user.roles as role (role.id)}
								<Item.Root variant="outline" class="w-full">
									<Item.Content class="flex flex-row items-center gap-2">
										<Icon class="text-xl" icon={'shield'} />
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
				{#if user.roles.length > 0 && hasPermission('admin:user:write')}
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
				{#if hasPermission('admin:user:write')}
					<Card.Footer>
						<Button class="ml-auto" onclick={updateUser}>Update</Button>
					</Card.Footer>
				{/if}
			</Card.Root>
		</div>
	{/if}
</Sidebar>

<Modal bind:this={deleteUserModal} button class="xl:w-xl!">
	<h1 class="text-xl font-semibold">Are you sure?</h1>
	<h2 class="text-card-foreground/60 sm:max-w-10/12">
		This action can not be reversed! All of the user's sessions, emails and
		accounts will be permanently deleted.
	</h2>

	<div class="flex items-center gap-2 justify-end mt-4">
		<Button variant="secondary" onclick={() => deleteUserModal?.close()}>
			Cancel
		</Button>
		<Button variant="destructive" onclick={deleteUser}>Delete User</Button>
	</div>
</Modal>

<Modal bind:this={signOutModal} button class="xl:w-xl!">
	<h1 class="text-xl font-semibold">Sign out all sessions</h1>
	<h2 class="text-card-foreground/60 sm:max-w-10/12">
		This will invalidate every session for {user?.username}, issued access
		tokens remain valid until they expire.
	</h2>

	<div class="flex items-center gap-2 justify-end mt-4">
		<Button variant="secondary" onclick={() => signOutModal?.close()}>
			Cancel
		</Button>
		<Button variant="destructive" onclick={invalidateAllSessions}>
			Sign out sessions
		</Button>
	</div>
</Modal>

<Modal
	bind:this={setPasswordModal}
	bind:showModal={setPasswordModalOpen}
	button
	class="xl:w-xl!"
>
	<h1 class="text-xl font-semibold">Set password</h1>
	<h2 class="text-card-foreground/60 sm:max-w-10/12">
		Set a new password for {user?.username}. All of their sessions will be
		signed out and they will need to log in again with the new password.
	</h2>

	<form onsubmit={setPassword} class="mt-6 flex flex-col gap-5">
		<div class="grid gap-2">
			<Label for="newPassword">New Password</Label>
			<Input
				bind:value={newPassword}
				id="newPassword"
				type="password"
				autocomplete="new-password"
				required
			/>
		</div>

		<div class="grid gap-2">
			<Label for="confirmPassword">Confirm Password</Label>
			<Input
				bind:value={confirmPassword}
				id="confirmPassword"
				type="password"
				autocomplete="new-password"
				required
			/>
		</div>

		<div class="flex items-center gap-2 justify-end mt-1">
			<Button variant="secondary" onclick={() => setPasswordModal?.close()}>
				Cancel
			</Button>
			<Button
				type="submit"
				disabled={loading || !newPassword || !confirmPassword}
			>
				{#if loading}
					<Spinner />
				{/if}
				Set Password
			</Button>
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
