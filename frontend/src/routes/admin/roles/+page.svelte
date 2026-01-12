<script lang="ts">
	import { browser } from "$app/environment";
	import { goto } from "$app/navigation";
	import * as star from '$lib/star';
	import Icon from "@/components/icon.svelte";
	import Sidebar from "@/components/sidebar.svelte";
	import { Badge } from "@/components/ui/badge";
	import { Button } from "@/components/ui/button";
	import * as Table from '@/components/ui/table';
	import { hasPermission } from "@/index";
	import { relativeTime, titleCase } from "@/utils";
	import { onMount } from "svelte";

	let roles = $state<star.Role[]>([])	

	onMount(async () => {
		if (!browser) return
		if (!hasPermission('star:admin:read:roles')) return location.href = '/'
		roles = await star.getAllRoles()
	})
</script>

<Sidebar class='md:p-8 max-w-6xl mx-auto'>
	<h1 class='text-2xl font-semibold'>Roles</h1>
	<h2 class='text-lg text-primary/60'>View and manage roles, star uses role-based access control.</h2>
	
	<div class='mt-24'>
		<div class="flex flex-row items-center gap-2 my-2">
			<Button class='ml-auto'>Create Role</Button>
		</div>

		<Table.Root>
			<Table.Caption>Found {roles.length} role{roles.length == 1 ? '' : 's'}</Table.Caption>
			<Table.Header>
				<Table.Row>
				<Table.Head>Role</Table.Head>
				<Table.Head>Permissions</Table.Head>
				<Table.Head>Created</Table.Head>
				<Table.Head class='text-end'>Actions</Table.Head>
				</Table.Row>
			</Table.Header>
			<Table.Body>
				{#each roles as role (role.id)}
				<Table.Row>
					<Table.Cell class='flex flex-row gap-2 items-center'>
						<Icon class='text-3xl' icon='package' />
						<div class="flex flex-col items-start">
							<div class='flex items-center gap-2'>
								<span>{titleCase(role.name)}</span>
								{#if role.global}
									<Badge variant='secondary'>Default role</Badge>
								{/if}
							</div>
							<span class='text-primary/70'>{role.description}</span>
						</div>
					</Table.Cell>
					<Table.Cell>
						<span>{role.permissions.length} permission{role.permissions.length == 1 ? '' : 's'}</span>
					</Table.Cell>
					<Table.Cell>{relativeTime(role.created_at)} ago</Table.Cell>
					<Table.Cell class='w-0'><Button variant='secondary' class='ml-auto' size='sm' onclick={() => goto(`/admin/roles/${role.id}`)}>Manage</Button></Table.Cell>
				</Table.Row>
				{/each}
			</Table.Body>
		</Table.Root>
	</div>
</Sidebar>

<!-- <Modal bind:this={createAppModal}>
	<h1 class='text-xl font-semibold'>Create Application</h1>
	<h2 class='text-card-foreground/60 sm:max-w-10/12'>Create a new application, allowing it to authenticate users that are registered with star</h2>

	<form onsubmit={createApp} class='mt-6'>
		<div class="flex flex-col gap-6 w-full mx-auto">
			<div class="grid gap-2">
				<Label for="name">Name</Label>
				<Input id="name" type="text" placeholder='' bind:value={name} required />
			</div>
			<div class="grid gap-2">
				<Label for="slug">Slug</Label>
				<Input id="slug" type="text" placeholder='' bind:value={slug} required />
			</div>
			<div class="grid gap-2">
				<Label for="url">URL</Label>
				<Input id="url" type="url" placeholder='' bind:value={url} required />
			</div>
		</div>

		<div class='flex items-center gap-2 justify-end mt-6'>
			<Button variant='secondary' onclick={() => createAppModal?.close()}>Cancel</Button>
			<Button disabled={loading} onclick={createApp}>
				{#if loading}
					<Spinner />
				{/if}

				Create Application
			</Button>
		</div>
	</form>
</Modal> -->