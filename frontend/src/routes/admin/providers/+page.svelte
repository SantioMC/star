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
	import { relativeTime, titleCase } from "@/utils";
	import { onMount } from "svelte";
	import { toast } from "svelte-sonner";

	let providers = $state<star.SocialProviderConfig[]>([])
	let presets = $state<star.SocialProviderPreset[]>([])
	let loading = $state<boolean>(false)

	let selectedPreset = $state<string | null>(null)
	let clientId = $state<string>('')
	let clientSecret = $state<string>('')
	let discoveryUri = $state<string>('')
	let customName = $state<string>('')

	let createProviderModal = $state<Modal>()

	let isCustom = $derived(selectedPreset === '__custom')
	let availablePresets = $derived(
		presets.filter(p => !providers.some(existing => existing.name === p.name))
	)

	onMount(async () => {
		if (!browser) return
		if (!hasPermission('star:admin:read:providers')) return location.href = '/'
		providers = await star.getSocialProviders()
		presets = await star.getSocialProviderPresets()
	})

	const openCreate = () => {
		selectedPreset = null
		clientId = ''
		clientSecret = ''
		discoveryUri = ''
		customName = ''
		createProviderModal?.open()
	}

	const selectPreset = (name: string) => {
		selectedPreset = name
	}

	const createProvider = async (e: Event) => {
		e.preventDefault()
		loading = true

		const name = isCustom ? customName : selectedPreset!

		const result = await star.createSocialProvider({
			name,
			client_id: clientId,
			client_secret: clientSecret,
			discovery_uri: isCustom && discoveryUri ? discoveryUri : undefined,
			order: providers.length
		})

		if ('error' in result) {
			loading = false
			toast.error(result.message)
		} else {
			providers = await star.getSocialProviders()
			toast.success(`Successfully added ${titleCase(name)} provider`)
			createProviderModal?.close()
			loading = false
		}
	}

	const back = () => {
		selectedPreset = null
	}
</script>

<Sidebar class='md:p-8 max-w-6xl mx-auto'>
	<h1 class='text-2xl font-semibold'>Social Providers</h1>
	<h2 class='text-lg text-primary/60'>Configure external identity providers for social login</h2>

	<div class='mt-24'>
		<div class="flex flex-row items-center gap-2 my-2">
			{#if hasPermission('star:admin:create:providers')}
				<Button class='ml-auto' onclick={openCreate}>Add Provider</Button>
			{/if}
		</div>

		<Table.Root>
			<Table.Caption>Found {providers.length} provider{providers.length === 1 ? '' : 's'}</Table.Caption>
			<Table.Header>
				<Table.Row>
				<Table.Head>Provider</Table.Head>
				<Table.Head>Client ID</Table.Head>
				<Table.Head>Order</Table.Head>
				<Table.Head>Created</Table.Head>
				<Table.Head class='text-end'>Actions</Table.Head>
				</Table.Row>
			</Table.Header>
			<Table.Body>
				{#each providers as provider (provider.id)}
				<Table.Row>
					<Table.Cell class='flex flex-row gap-2 items-center'>
						<Icon class='text-3xl' icon={provider.icon || 'plug-connected'} />
						<div class="flex flex-col items-start">
							<div class='flex items-center gap-2'>
								<span>{titleCase(provider.name)}</span>
								{#if provider.oidc_url}
									<Badge variant='secondary'>OIDC</Badge>
								{/if}
							</div>
							<span class='text-primary/70'>{provider.id}</span>
						</div>
					</Table.Cell>
					<Table.Cell>
						<code class="text-sm">{provider.client_id.substring(0, 16)}...</code>
					</Table.Cell>
					<Table.Cell>{provider.order}</Table.Cell>
					<Table.Cell>{relativeTime(provider.created_at)} ago</Table.Cell>
					<Table.Cell class='w-0'>
						<Button variant='secondary' class='ml-auto' size='sm' onclick={() => goto(`/admin/providers/${provider.id}`)}>Manage</Button>
					</Table.Cell>
				</Table.Row>
				{/each}
			</Table.Body>
		</Table.Root>
	</div>
</Sidebar>

<Modal bind:this={createProviderModal} button>
	<h1 class='text-xl font-semibold'>Add Social Provider</h1>
	<h2 class='text-card-foreground/60 sm:max-w-10/12'>
		{#if !selectedPreset}
			Select a provider to configure social login
		{:else if isCustom}
			Configure a custom OIDC provider
		{:else}
			Enter your {titleCase(selectedPreset)} OAuth credentials
		{/if}
	</h2>

	{#if !selectedPreset}
		<div class="flex flex-col gap-2 my-4">
			{#each availablePresets as preset (preset.name)}
				<button
					class="flex items-center gap-3 rounded-lg border border-primary/20 px-4 py-3 text-left transition-colors hover:bg-primary/10"
					onclick={() => selectPreset(preset.name)}
				>
					<Icon icon={preset.icon} class="text-2xl" />
					<span class="font-medium">{titleCase(preset.name)}</span>
				</button>
			{/each}
			<button
				class="flex items-center gap-3 rounded-lg border border-primary/20 px-4 py-3 text-left transition-colors hover:bg-primary/10"
				onclick={() => selectPreset('__custom')}
			>
				<Icon icon="settings" class="text-2xl" />
				<div class="flex flex-col">
					<span class="font-medium">Custom OIDC Provider</span>
					<span class="text-sm text-primary/60">Connect via OIDC discovery URL</span>
				</div>
			</button>
		</div>
	{:else}
		<form onsubmit={createProvider} class='mt-6'>
			<div class="flex flex-col gap-6 w-full mx-auto">
				{#if isCustom}
					<div class="grid gap-2">
						<Label for="customName">Provider Name</Label>
						<Input id="customName" type="text" placeholder='my-provider' bind:value={customName} required />
					</div>
					<div class="grid gap-2">
						<Label for="discoveryUri">OIDC Discovery URL</Label>
						<Input id="discoveryUri" type="url" placeholder='https://example.com/.well-known/openid-configuration' bind:value={discoveryUri} required />
					</div>
				{/if}
				<div class="grid gap-2">
					<Label for="clientId">Client ID</Label>
					<Input id="clientId" type="text" bind:value={clientId} required />
				</div>
				<div class="grid gap-2">
					<Label for="clientSecret">Client Secret</Label>
					<Input id="clientSecret" type="password" bind:value={clientSecret} required />
				</div>
			</div>

			<div class='flex items-center gap-2 justify-end mt-6'>
				<Button variant='secondary' onclick={back}>Back</Button>
				<Button disabled={loading} type='submit'>
					{#if loading}
						<Spinner />
					{/if}

					Add Provider
				</Button>
			</div>
		</form>
	{/if}
</Modal>
