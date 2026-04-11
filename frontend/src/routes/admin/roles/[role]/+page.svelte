<script lang="ts">
	import { browser } from '$app/environment';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import * as star from '$lib/star';
	import Icon from '@/components/icon.svelte';
	import Modal from '@/components/modal.svelte';
	import Sidebar from '@/components/sidebar.svelte';
	import { Badge } from '@/components/ui/badge';
	import { Button } from '@/components/ui/button';
	import * as Card from '@/components/ui/card';
	import * as DropdownMenu from '@/components/ui/dropdown-menu';
	import { Input } from '@/components/ui/input';
	import { Label } from '@/components/ui/label';
	import { Spinner } from '@/components/ui/spinner';
	import { hasPermission } from '@/index';
	import { formatTime } from '@/utils';
	import { onMount } from 'svelte';
	import { toast } from 'svelte-sonner';

	let loading = $state<boolean>(false);

	let role = $state<star.Role | null>(null);
	let name = $state<string>('');
	let description = $state<string>('');
	let permissionInput = $state<string>('');

	let deleteRoleModal = $state<Modal>();
	let addPermissionModal = $state<Modal>();

	let canModify = $derived(
		hasPermission('star:admin:update:roles') && role != null && !role.super && !role.global
	);

	onMount(async () => {
		if (!browser) return;
		if (!hasPermission('star:admin:read:roles')) return goto('/');

		role = await star.getRole(page.params.role!);
		if (role == null) return goto('/admin/roles');

		name = role.name;
		description = role.description;
	});

	const updateRole = async () => {
		loading = true;
		const result = await star.updateRole(role!.id, {
			name,
			description
		});
		loading = false;

		if ('error' in result) {
			toast.error(result.message);
		} else {
			role = result;
			toast.success('Successfully updated role');
		}
	};

	const addPermission = async () => {
		const trimmed = permissionInput.trim();
		if (!trimmed) return;

		if (role!.permissions.includes(trimmed)) {
			toast.error('This permission already exists on this role');
			return;
		}

		loading = true;
		const updated = [...role!.permissions, trimmed];
		const result = await star.updateRole(role!.id, { scopes: updated });
		loading = false;

		if ('error' in result) {
			toast.error(result.message);
		} else {
			role = result;
			permissionInput = '';
			addPermissionModal?.close();
			toast.success('Successfully added permission');
		}
	};

	const removePermission = async (permission: string) => {
		loading = true;
		const updated = role!.permissions.filter(p => p !== permission);
		const result = await star.updateRole(role!.id, { scopes: updated });
		loading = false;

		if ('error' in result) {
			toast.error(result.message);
		} else {
			role = result;
			toast.success('Removed permission');
		}
	};

	const deleteRole = async () => {
		loading = true;
		const result = await star.deleteRole(role!.id);
		loading = false;
		deleteRoleModal?.close();

		if ('error' in result) {
			toast.error(result.message);
		} else {
			goto('/admin/roles');
		}
	};
</script>

<Sidebar class="md:p-8 max-w-6xl mx-auto">
	{#if role}
		<div class="flex flex-row items-center">
			<div>
				<h1 class="text-2xl font-semibold">Role Management</h1>
				<h2 class="text-lg text-primary/60">
					Viewing role {role.name}
				</h2>
			</div>

			<Button
				variant="ghost"
				class="ml-auto"
				onclick={() => goto('/admin/roles')}
			>
				<Icon icon="chevron-left" />Go back
			</Button>
		</div>

		<div class="mt-16 flex flex-col gap-8">
			<div class="flex flex-col xl:flex-row items-center gap-4">
				<Icon class="text-3xl" icon="package" />
				<div class="flex flex-col items-center xl:items-start justify-center gap-2">
					<h3 class="text-lg leading-4 flex items-center gap-2">
						{role.name}
						{#if role.global}
							<Badge variant="secondary">Default role</Badge>
						{/if}
						{#if role.super}
							<Badge>Super role</Badge>
						{/if}
					</h3>
					<div class="flex flex-col xl:flex-row items-center xl:gap-3 text-sm text-primary/60 flex-wrap">
						<span>
							<Icon icon="id" />
							{role.id}
						</span>
						<span>
							<Icon icon="calendar-event" />
							Created {formatTime(role.created_at)}
						</span>
						<span>
							<Icon icon="lock" />
							{role.permissions.length} permission{role.permissions.length === 1 ? '' : 's'}
						</span>
					</div>
				</div>

				{#if canModify}
					<DropdownMenu.Root>
						<DropdownMenu.Trigger class="ml-auto">
							<Button variant="ghost" size="icon">
								<Icon icon="dots" class="text-2xl" />
							</Button>
						</DropdownMenu.Trigger>
						<DropdownMenu.Content>
							<DropdownMenu.Group>
								{#if hasPermission('star:admin:delete:roles')}
									<DropdownMenu.Item
										variant="destructive"
										onclick={() => deleteRoleModal?.open()}
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

			{#if role.super}
				<Card.Root class="w-full">
					<Card.Content class="flex flex-col overflow-hidden flex-wrap text-yellow-100">
						<Card.Title class="text-yellow-200 mb-4">
							<Icon icon="shield" /> Super Role
						</Card.Title>
						<span>
							This role grants full access to all permissions and can not be modified.
						</span>
					</Card.Content>
				</Card.Root>
			{/if}

			{#if role.global}
				<Card.Root class="w-full">
					<Card.Content class="flex flex-col overflow-hidden flex-wrap text-blue-100">
						<Card.Title class="text-blue-200 mb-4">
							<Icon icon="users" /> Default Role
						</Card.Title>
						<span>
							This role is automatically assigned to all users and can not be modified.
						</span>
					</Card.Content>
				</Card.Root>
			{/if}

			<Card.Root class="min-w-1/3 w-full md:w-auto">
				<Card.CardHeader>
					<Card.Title class="flex items-center gap-1">
						<Icon icon="package" class="text-2xl" /> Role Information
					</Card.Title>
				</Card.CardHeader>
				<Card.Content class="flex flex-col gap-6">
					<form class="flex flex-col gap-6">
						<div class="grid gap-2">
							<Label for="name">Name</Label>
							<Input id="name" type="text" required bind:value={name} disabled={!canModify} />
						</div>

						<div class="grid gap-2">
							<Label for="description">Description</Label>
							<Input id="description" type="text" bind:value={description} disabled={!canModify} />
						</div>
					</form>
				</Card.Content>
				{#if canModify}
					<Card.Footer>
						<Button class="ml-auto" onclick={updateRole} disabled={loading}>
							{#if loading}
								<Spinner />
							{/if}
							Update
						</Button>
					</Card.Footer>
				{/if}
			</Card.Root>

			<Card.Root class="min-w-1/3 w-full md:w-auto">
				<Card.CardHeader>
					<Card.Title class="flex items-center gap-1">
						<Icon icon="lock" class="text-2xl" /> Permissions
					</Card.Title>
					<Card.Description class="md:max-w-2/3">
						Permissions granted to users with this role. These are matched against the <code>scopes</code> claim in access tokens.
					</Card.Description>
				</Card.CardHeader>
				<Card.Content>
					{#if role.permissions.length === 0}
						<div class="flex flex-col gap-4 items-center w-full my-12">
							<span class="text-xl font-semibold leading-2">No Permissions</span>
							<span class="text-primary/60 leading-5 m-0 -translate-y-1.5 text-center max-w-3/4">
								This role currently has no assigned permissions.
							</span>
							{#if canModify}
								<Button onclick={() => addPermissionModal?.open()}>Add Permission</Button>
							{/if}
						</div>
					{:else}
						<div class="flex flex-col gap-2">
							{#each role.permissions as permission (permission)}
								<div class="flex items-center justify-between rounded-md border border-primary/20 px-4 py-2">
									<code class="text-sm">{permission}</code>
									{#if canModify}
										<Button
											variant="ghost"
											size="icon"
											class="h-8 w-8 text-destructive hover:text-destructive"
											onclick={() => removePermission(permission)}
										>
											<Icon icon="x" />
										</Button>
									{/if}
								</div>
							{/each}
						</div>
					{/if}
				</Card.Content>
				{#if canModify && role.permissions.length > 0}
					<Card.Footer>
						<Button class="ml-auto" onclick={() => addPermissionModal?.open()}>
							Add Permission
						</Button>
					</Card.Footer>
				{/if}
			</Card.Root>
		</div>
	{/if}
</Sidebar>

<Modal bind:this={deleteRoleModal} button class="xl:w-xl!">
	<h1 class="text-xl font-semibold">Are you sure?</h1>
	<h2 class="text-card-foreground/60 sm:max-w-10/12">
		This action can not be reversed! Deleting this role will remove it from all users who currently have it assigned.
	</h2>

	<div class="flex items-center gap-2 justify-end mt-4">
		<Button variant="secondary" onclick={() => deleteRoleModal?.close()}>
			Cancel
		</Button>
		<Button variant="destructive" onclick={deleteRole} disabled={loading}>
			{#if loading}
				<Spinner />
			{/if}
			Delete Role
		</Button>
	</div>
</Modal>

<Modal bind:this={addPermissionModal} button class="xl:w-xl!">
	<h1 class="text-xl font-semibold">Add Permission</h1>
	<h2 class="text-card-foreground/60 sm:max-w-10/12">
		Add a permission scope to this role. Users with this role will be granted this permission.
	</h2>

	<form onsubmit={addPermission} class="mt-6">
		<div class="flex flex-col gap-6 w-full mx-auto">
			<div class="grid gap-2">
				<Label for="permission">Permission</Label>
				<Input
					id="permission"
					type="text"
					placeholder="star:user:read"
					bind:value={permissionInput}
					required
				/>
			</div>
		</div>

		<div class="flex items-center gap-2 justify-end mt-6">
			<Button variant="secondary" onclick={() => addPermissionModal?.close()}>
				Cancel
			</Button>
			<Button disabled={loading} type="submit">
				{#if loading}
					<Spinner />
				{/if}
				Add Permission
			</Button>
		</div>
	</form>
</Modal>
