<script lang="ts">
	import { browser } from "$app/environment";
	import { goto } from "$app/navigation";
	import * as star from '$lib/star';
	import Icon from "@/components/icon.svelte";
	import Sidebar from "@/components/sidebar.svelte";
	import { Badge } from "@/components/ui/badge";
	import { Button } from "@/components/ui/button";
	import * as InputGroup from '@/components/ui/input-group';
	import { Spinner } from "@/components/ui/spinner";
	import * as Table from '@/components/ui/table';
	import { hasPermission } from "@/index";
	import { debounced, relativeTime } from "@/utils";
	import { onMount } from "svelte";

	let users = $state<star.AdminUser[]>([])
	let loading = $state<boolean>(true)
	let query = $state<string>('')

	const onSearch = async () => {
		loading = true;
		users = await star.getAllUsers(1, query)
		loading = false;
	}

	onMount(async () => {
		if (!browser) return
		if (!hasPermission('admin:user:read')) return location.href = '/'
		users = await star.getAllUsers()
		loading = false
	})
</script>

<Sidebar class='lg:p-8 max-w-6xl mx-auto'>
	<h1 class='text-2xl font-semibold'>User Management</h1>
	<h2 class='text-lg text-primary/60'>View and manage users</h2>

	<div class='mt-24'>
		<div class="flex flex-row items-center gap-2">
			<InputGroup.Root class='w-full xl:max-w-1/3 my-4'>
				<InputGroup.Input onkeyup={debounced(onSearch)} bind:value={query} placeholder="Search..." />
				<InputGroup.Addon>
					{#if loading}
					<Spinner />
					{:else}
					<Icon icon='search' />
					{/if}
				</InputGroup.Addon>
			</InputGroup.Root>
		</div>

		<Table.Root>
			<Table.Caption>Found {users.length} user{users.length == 1 ? '' : 's'}</Table.Caption>
			<Table.Header>
				<Table.Row>
				<Table.Head class='w-1/4'>User</Table.Head>
				<Table.Head>Email</Table.Head>
				<Table.Head>Roles</Table.Head>
				<Table.Head>Created</Table.Head>
				<Table.Head class='text-end'>Actions</Table.Head>
				</Table.Row>
			</Table.Header>
			<Table.Body>
				{#each users as user (user.id)}
				<Table.Row>
					<Table.Cell class='flex flex-row gap-3 items-center'>
						<img src={user.profile_picture ?? ""} alt="{user.username}'s profile picture" class='w-10 h-10 rounded-full' />
						<div class="flex flex-col items-start">
							<span>{user.username}</span>
							<span class='text-primary/70 text-ellipsis overflow-hidden w-56'>{user.email ?? 'No email'}</span>
						</div>
					</Table.Cell>
					<Table.Cell>
						{#if user.email_verified}
							<Badge variant='outline' class='border-emerald-600/60 text-emerald-500'>Verified</Badge>
						{:else}
							<Badge variant='outline' class='border-gray-400/60 text-gray-400'>Unverified</Badge>
						{/if}
					</Table.Cell>
					<Table.Cell>
						{#if user.roles.length == 0}
							<span class='text-primary/50'>None</span>
						{:else}
							<div class='flex flex-row flex-wrap gap-1'>
								{#each user.roles as role (role.id)}
									<Badge variant='outline'>{role.name}</Badge>
								{/each}
							</div>
						{/if}
					</Table.Cell>
					<Table.Cell>{relativeTime(user.created_at)} ago</Table.Cell>
					<Table.Cell class='w-0'><Button variant='secondary' class='ml-auto' size='sm' onclick={() => goto(`/admin/users/${user.id}`)}>Manage</Button></Table.Cell>
				</Table.Row>
				{/each}
			</Table.Body>
		</Table.Root>
	</div>
</Sidebar>
