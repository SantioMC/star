<script lang="ts">
	import { browser } from "$app/environment";
	import { goto } from "$app/navigation";
	import { page } from "$app/state";
	import * as star from '$lib/star';
	import Icon from "@/components/icon.svelte";
	import Modal from "@/components/modal.svelte";
	import Session from "@/components/session.svelte";
	import Sidebar from "@/components/sidebar.svelte";
	import { Badge } from "@/components/ui/badge";
	import { Button } from "@/components/ui/button";
	import * as Card from '@/components/ui/card';
	import { Input } from "@/components/ui/input";
	import { Label } from "@/components/ui/label";
	import { hasPermission } from "@/index";
	import { relativeTime, titleCase } from "@/utils";
	import { onMount } from "svelte";
	import { toast } from "svelte-sonner";

	const dateFormatter = new Intl.DateTimeFormat(undefined, {
		month: "short",
		day: "2-digit",
		year: "numeric",
	})

	let deleteUserModal = $state<Modal>()
	let suspendUserModal = $state<Modal>()
	let unsuspendUserModal = $state<Modal>()
	let changePasswordModal = $state<Modal>()

	let reason = $state<string>('')
	let password = $state<string>('')
	let username = $state<string>('')
	let firstName = $state<string>('')
	let lastName = $state<string>('')

	let user = $state<star.AdminUser|null>(null)
	let suspender = $state<star.AdminUser|null>(null)
	
	let activeSessions = $derived<star.Session[]>(
		user?.sessions?.filter(s => !s.expired)?.sort((a, b) => b.created_at - a.created_at) ?? []
	)

	onMount(async () => {
		if (!browser) return
		if (!hasPermission('star:admin:read:users')) return goto('/')
		user = await star.getUser(page.params.user!)
		if (user == null) return goto('/admin/users')

		username = user.username
		firstName = user.first_name
		lastName = user.last_name

		if (user.suspension?.issued_by) {
			suspender = await star.getUser(user.suspension.issued_by)
		}
	})

	const deleteUser = async () => {
		deleteUserModal?.close()
		const result = await star.deleteUser(user!.id)
		if ('success' in result) {
			goto('/admin/users')
		} else {
			toast.error(result.message)
		}
	}

	const suspendUser = async () => {
		suspendUserModal?.close()
		const result = await star.suspendUser(user!.id, reason)
		reason = '';

		if ('success' in result) {
			location.reload()
		} else {
			toast.error(result.message)
		}
	}

	const unsuspendUser = async () => {
		unsuspendUserModal?.close()
		const result = await star.unsuspendUser(user!.id)
		if ('success' in result) {
			location.reload()
		} else {
			toast.error(result.message)
		}
	}

	const changePassword = async () => {
		changePasswordModal?.close()

		const result = await star.forceChangePassword(user!.id, password)
		password = '';

		if ('success' in result) {
			toast.success('Successfully changed password')
		} else {
			toast.error(result.message)
		}
	}

	const updateUser = async () => {
		const result = await star.updateUser(user!.id, {
			family_name: lastName,
			given_name: firstName,
			username: username
		})

		if ('error' in result) {
			firstName = user!.first_name
			lastName = user!.last_name
			username = user!.username
		} else {
			user!.first_name = firstName
			user!.last_name = lastName
			user!.username = username
			toast.success('Successfully updated user')
		}
	}

	const invalidateSession = (session: star.Session) => {
		user!.sessions = user!.sessions.filter(s => s.id != session.id)
		toast.success('You have invalidated the session')
	}
</script>

<Sidebar class='md:p-8 max-w-6xl mx-auto'>
	{#if user}
		<div class='flex flex-row items-center'>
			<div>
				<h1 class='text-2xl font-semibold'>User Management</h1>
				<h2 class='text-lg text-primary/60'>Viewing profile for {user?.username}</h2>
			</div>

			<Button variant='ghost' class='ml-auto' onclick={() => goto('/admin/users')}><Icon icon='chevron-left' />Go back</Button>
		</div>

		<div class='mt-24 flex flex-col gap-8'>
			<div class="flex flex-col xl:flex-row items-center gap-4">
				<img src={user.profile_picture} alt="{user.username}'s profile picture" class='rounded-full w-16 h-16' />
				<div class="flex flex-col items-center xl:items-start justify-center gap-2">
					<h3 class='text-lg leading-4'>
						{#if user.first_name}
							{user.first_name} {user.last_name}
						{:else}
							{user.username}
						{/if}
					</h3>
					<div class="flex flex-col xl:flex-row items-center xl:gap-3 text-sm text-primary/60 flex-wrap">
						<span>
							<Icon icon='id' />
							{user.id}
						</span>
						<span>
							<Icon icon='mail' />
							{user.email}
						</span>
						<span>
							<Icon icon='calendar-event' />
							Joined {dateFormatter.format(new Date(user.created_at * 1000))}
						</span>
					</div>
					{#if user.roles.length > 0}
						<div class='flex flex-col xl:flex-row items-center gap-4'>
							{#each user.roles as role (role.id)}
								<Badge variant='secondary'>{role.name}</Badge>
							{/each}
						</div>
					{/if}
				</div>

				<div class='xl:ml-auto flex items-center gap-2'>
					{#if !user.suspension && hasPermission('star:admin:update:users')}
					<Button onclick={() => changePasswordModal?.open()} variant='ghost'>
						<Icon icon='password' />
						Change Password
					</Button>

					<Button onclick={() => suspendUserModal?.open()} class='bg-transparent text-red-400 hover:bg-red-400/30 hover:border-transparent hover:text-primary'>
						<Icon icon='hammer' />
						Suspend User
					</Button>
					{/if}

					{#if hasPermission('star:admin:delete:users')}
					<Button variant='destructive' onclick={() => deleteUserModal?.open()}>
						<Icon icon='trash' />
						Delete User
					</Button>
					{/if}
				</div>
			</div>

			{#if user.suspension}
			<Card.Root class='w-full'>
				<Card.Content class="flex flex-col overflow-hidden flex-wrap text-orange-100">
					<Card.Title class='text-orange-200 mb-4'><Icon icon='gavel' /> User is suspended</Card.Title>
					<span>This user is currently suspended for the reason "{user.suspension.reason}"</span>
					
					<div class="my-4 flex flex-col">
					{#if user.suspension.expires_at}
						<span>
							<Icon icon='clock' />
							This suspension expires on {dateFormatter.format(user.suspension.expires_at * 1000)} (in {relativeTime(user.suspension.expires_at)})
						</span>
					{/if}
					
					{#if suspender}
						<span>
							<Icon icon='hammer' />
							This suspension was issued by {suspender?.username}
						</span>
					{/if}
					</div>

					{#if hasPermission('star:admin:update:users') || suspender}
					<div class="ml-auto flex items-center gap-2">
						{#if hasPermission('star:admin:update:users')}
						<Button onclick={() => unsuspendUserModal?.open()} class='bg-transparent text-orange-400 hover:bg-orange-400/30 hover:border-transparent hover:text-primary'>
							<Icon icon='hammer-off' />
							Revoke Suspension
						</Button>
						{/if}

						{#if suspender}
							<Button variant='outline' onclick={() => location.href = `/admin/users/${suspender!.id}`}>View Issuer</Button>
						{/if}
					</div>
					{/if}
				</Card.Content>
			</Card.Root>
			{/if}

			<Card.Root class='min-w-1/3 w-full md:w-auto'>
				<Card.CardHeader>
					<Card.Title><Icon icon='gavel' /> Roles</Card.Title>
				</Card.CardHeader>
				<Card.Content class="flex flex-col gap-1">
					<p class="text-primary/60 text-sm">The user has the following roles:</p>
					{#if user.roles.length == 0}
						<div class="flex flex-col gap-4 items-center w-full my-16">
							<span class='text-xl font-semibold leading-2'>No Roles</span>
							<span class='text-primary/60 leading-2'>This user currently has no assigned roles</span>
						</div>
					{:else}
						<div class='flex gap-2 items-center'>
							{#each user.roles as role (role.id)}
								<Badge variant='secondary' class='hover:*:block hover:*:opacity-100 hover:bg-secondary/60 cursor-pointer transition-all'>
									<Icon icon='folder' class='p-1' />{titleCase(role.name)}
									<Icon icon='x' class='p-1 hidden opacity-0 transition-all' />
								</Badge>
							{/each}
							<Badge variant='secondary' class='w-fit hover:bg-secondary/60 transition-colors cursor-pointer'>
								<Icon icon='plus' class='p-1' /> Add Role
							</Badge>
						</div>
					{/if}
				</Card.Content>
			</Card.Root>

			<Card.Root class='min-w-1/3 w-full md:w-auto'>
				<Card.CardHeader>
					<Card.Title><Icon icon='user' /> Profile Information</Card.Title>
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
					<Button class='ml-auto' onclick={updateUser}>Update</Button>
				</Card.Footer>
			</Card.Root>

			<Card.Root class='min-w-1/3 w-full md:w-auto'>
				<Card.CardHeader>
					<Card.Title><Icon icon='device-desktop' /> Sessions</Card.Title>
					<Card.Description class='max-w-xl'>The different devices and sessions currently logged into the account, invalidating a session might take up to an hour to take effect.</Card.Description>
				</Card.CardHeader>
				<Card.Content class="flex flex-col gap-2">
					{#each activeSessions as session}
						<Session {session} oninvalidate={() => invalidateSession(session)} />
					{/each}

					{#if activeSessions.length == 0}
						<div class="flex flex-col gap-4 items-center w-full my-24">
							<span class='text-xl font-semibold leading-2'>No Sessions</span>
							<span class='text-primary/60 leading-2'>This user currently has no sessions</span>
						</div>
					{/if}
				</Card.Content>
				{#if activeSessions.length != 0}
					<Card.Footer>
						<Button class='ml-auto' variant='destructive'>Sign out all sessions</Button>
					</Card.Footer>
				{/if}
			</Card.Root>
		</div>
	{/if}
</Sidebar>

<Modal bind:this={deleteUserModal} button class='xl:w-xl!'>
	<h1 class='text-xl font-semibold'>Are you sure?</h1>
	<h2 class='text-card-foreground/60 sm:max-w-10/12'>This action can not be reversed! It is recommended to instead suspend a user over deleting.</h2>

	<div class='flex items-center gap-2 justify-end mt-4'>
		<Button variant='secondary' onclick={() => deleteUserModal?.close()}>Cancel</Button>
		<Button variant='destructive' onclick={deleteUser}>Delete User</Button>
	</div>
</Modal>

<Modal bind:this={suspendUserModal} button class='xl:w-xl!'>
	<h1 class='text-xl font-semibold'>Suspend {user?.username}</h1>
	<h2 class='text-card-foreground/60 sm:max-w-10/12'>Suspending a user will immediately invalidate all sessions and prevent the user from signing in again.</h2>

	<form onsubmit={suspendUser} class='mt-6'>
		<div class="flex flex-col gap-6 w-full mx-auto">
			<div class="grid gap-2">
				<Label for="reason">Reason</Label>
				<Input id="username" type="text" placeholder='Suspicious account behaviour' bind:value={reason} required />
			</div>
		</div>

		<div class='flex items-center gap-2 justify-end mt-6'>
			<Button variant='secondary' onclick={() => suspendUserModal?.close()}>Cancel</Button>
			<Button variant='destructive' onclick={suspendUser}>Suspend User</Button>
		</div>
	</form>
</Modal>

<Modal bind:this={unsuspendUserModal} button class='xl:w-xl!'>
	<h1 class='text-xl font-semibold'>Revoke Suspension</h1>
	<h2 class='text-card-foreground/60 sm:max-w-10/12'>Revoking {user?.username}'s suspension will allow them to continue using their account</h2>

	<form onsubmit={unsuspendUser} class='mt-6'>
		<div class='flex items-center gap-2 justify-end mt-6'>
			<Button variant='secondary' onclick={() => unsuspendUserModal?.close()}>Cancel</Button>
			<Button onclick={unsuspendUser}>Revoke Suspension</Button>
		</div>
	</form>
</Modal>

<Modal bind:this={changePasswordModal} button class='xl:w-xl!'>
	<h1 class='text-xl font-semibold'>Change Password</h1>
	<h2 class='text-card-foreground/60 sm:max-w-10/12'>Manually change the password of this user, invalidating their previous sessions and requiring them to sign in with the new password</h2>

	<form onsubmit={changePassword} class='mt-6'>
		<div class="flex flex-col gap-6 w-full mx-auto">
			<div class="grid gap-2">
				<Label for="password">New Password</Label>
				<Input id="password" type="password" placeholder='' bind:value={password} required />
			</div>
		</div>

		<div class='flex items-center gap-2 justify-end mt-6'>
			<Button variant='secondary' onclick={() => changePasswordModal?.close()}>Cancel</Button>
			<Button onclick={changePassword}>Change Password</Button>
		</div>
	</form>
</Modal>