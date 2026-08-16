<script lang="ts">
	import { browser } from "$app/environment";
	import { goto } from "$app/navigation";
	import * as star from '$lib/star';
	import Icon from "@/components/icon.svelte";
	import Modal from "@/components/modal.svelte";
	import Sidebar from "@/components/sidebar.svelte";
	import { Badge } from "@/components/ui/badge";
	import { Button } from "@/components/ui/button";
	import { Input } from "@/components/ui/input";
	import { Label } from "@/components/ui/label";
	import { Spinner } from "@/components/ui/spinner";
	import * as Table from '@/components/ui/table';
	import { hasPermission } from "@/index";
	import { relativeTime } from "@/utils";
	import { onMount } from "svelte";
	import { toast } from "svelte-sonner";

	let apps = $state<star.AdminApplication[]>([])
	let loading = $state<boolean>(false)

	let name = $state<string>('')
	let kind = $state<'public' | 'confidential'>('public')
	let url = $state<string>('')

	let createAppModal = $state<Modal>()

	onMount(async () => {
		if (!browser) return
		if (!hasPermission('application:read')) return location.href = '/'
		apps = await star.getAllApps()
	})

	const createApp = async (e: Event) => {
		e.preventDefault()
		loading = true

		const result = await star.createApplication({
			name, kind, url
		})

		if ('error' in result) {
			loading = false;
			toast.error(result.message)
		} else {
			apps = await star.getAllApps()
			toast.success('Successfully created a new application')
			createAppModal?.close()
			loading = false;
		}
	}
</script>

<Sidebar class='md:p-8 max-w-6xl mx-auto'>
	<h1 class='text-2xl font-semibold'>Applications</h1>
	<h2 class='text-lg text-primary/60'>View and manage different applications</h2>

	<div class='mt-24'>
		<div class="flex flex-row items-center gap-2 my-2">
			{#if hasPermission('application:create')}
				<Button class='ml-auto' onclick={() => createAppModal?.open()}>Create Application</Button>
			{/if}
		</div>

		<Table.Root>
			<Table.Caption>Found {apps.length} application{apps.length == 1 ? '' : 's'}</Table.Caption>
			<Table.Header>
				<Table.Row>
				<Table.Head>Application</Table.Head>
				<Table.Head>Type</Table.Head>
				<Table.Head>URL</Table.Head>
				<Table.Head>Created</Table.Head>
				<Table.Head class='text-end'>Actions</Table.Head>
				</Table.Row>
			</Table.Header>
			<Table.Body>
				{#each apps as app (app.id)}
				<Table.Row>
					<Table.Cell class='flex flex-row gap-2 items-center'>
						<Icon class='text-3xl' icon='package' />
						<div class="flex flex-col items-start">
							<div class='flex items-center gap-2'>
								<span>{app.name}</span>
								{#if app.internal}
									<Badge variant='secondary'>Built-in</Badge>
								{/if}
							</div>
							<span class='text-primary/70'>{app.id}</span>
						</div>
					</Table.Cell>
					<Table.Cell>
						<Badge variant='outline'>{app.kind}</Badge>
					</Table.Cell>
					<Table.Cell>
						<a href={app.application_url} class='text-blue-300' target="_blank">
							{app.application_url}
							<Icon icon='external-link' />
						</a>
					</Table.Cell>
					<Table.Cell>{relativeTime(app.created_at)} ago</Table.Cell>
					<Table.Cell class='w-0'><Button variant='secondary' class='ml-auto' size='sm' onclick={() => goto(`/admin/apps/${app.id}`)}>Manage</Button></Table.Cell>
				</Table.Row>
				{/each}
			</Table.Body>
		</Table.Root>
	</div>
</Sidebar>

<Modal bind:this={createAppModal} button>
	<h1 class='text-xl font-semibold'>Create Application</h1>
	<h2 class='text-card-foreground/60 sm:max-w-10/12'>Create a new application, allowing it to authenticate users that are registered with star</h2>

	<form onsubmit={createApp} class='mt-6'>
		<div class="flex flex-col gap-6 w-full mx-auto">
			<div class="grid gap-2">
				<Label for="name">Name</Label>
				<Input id="name" type="text" placeholder='' bind:value={name} required />
			</div>
			<div class="grid gap-2">
				<Label for="kind">Client Type</Label>
				<select id="kind" bind:value={kind} class='border rounded-md bg-transparent px-3 py-2 text-sm'>
					<option value='public'>Public (browser or native app, uses PKCE)</option>
					<option value='confidential'>Confidential (server-side, uses a client secret)</option>
				</select>
			</div>
			<div class="grid gap-2">
				<Label for="url">URL</Label>
				<Input id="url" type="url" placeholder='' bind:value={url} required />
			</div>
		</div>

		<div class='flex items-center gap-2 justify-end mt-6'>
			<Button variant='secondary' onclick={() => createAppModal?.close()}>Cancel</Button>
			<Button disabled={loading} type='submit'>
				{#if loading}
					<Spinner />
				{/if}

				Create Application
			</Button>
		</div>
	</form>
</Modal>
