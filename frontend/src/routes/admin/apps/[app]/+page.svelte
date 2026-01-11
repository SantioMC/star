<script lang="ts">
	import { browser } from '$app/environment';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import oauth2 from '$lib/assets/oauth2.png';
	import * as star from '$lib/star';
	import CardChoice from '@/components/card-choice.svelte';
	import Categorizied from '@/components/categorizied.svelte';
	import Copy from '@/components/copy.svelte';
	import Icon from '@/components/icon.svelte';
	import Modal from '@/components/modal.svelte';
	import OauthProvider from '@/components/oauth_provider.svelte';
	import Sidebar from '@/components/sidebar.svelte';
	import { Badge } from '@/components/ui/badge';
	import { Button } from '@/components/ui/button';
	import * as Card from '@/components/ui/card';
	import * as DropdownMenu from '@/components/ui/dropdown-menu';
	import { Input } from '@/components/ui/input';
	import { Label } from '@/components/ui/label';
	import { Spinner } from '@/components/ui/spinner';
	import { hasPermission } from '@/index';
	import { dateFormatter, formatTime } from '@/utils';
	import { onMount } from 'svelte';
	import { toast } from 'svelte-sonner';

	let app = $state<star.AdminApplication | null>(null);
	let permissions = $state<star.Permission[]>([])

	let loading = $state<boolean>(false);
	let hasProviders = $state<boolean>(false);
	let callbackUrl = $state<string>('');

	let createProviderModal = $state<Modal>()
	let deleteAppModal = $state<Modal>()
	let editCallbackURLModal = $state<Modal>()

	let currentCallbackUrl = $derived(
		app?.callback?.startsWith('/') ? app?.url + app?.callback : app?.callback
	);

	let canModify = $derived(
		app?.internal == false && hasPermission('star:admin:update:applications')
	);

	onMount(async () => {
		if (!browser) return;
		if (!hasPermission('star:admin:read:applications')) return goto('/');
		
		app = await star.getApplication(page.params.app!);
		if (app == null) return goto('/admin/apps');

		permissions = await star.getPermissions(app.id);
		hasProviders = !!app.oauth_provider
	});

	const deleteApp = async () => {
		loading = true;
		const result = await star.deleteApplication(app!.id);
		deleteAppModal?.close();
		loading = false;

		if ('error' in result) {
			toast.error(result.message)
		} else {
			goto('/admin/apps')
		}
	}

	$effect(() => {
		callbackUrl = currentCallbackUrl ?? ''
	})

	const updateCallbackURL = async () => {
		loading = true
		const result = await star.updateApplication(app!.id, { callback_url: callbackUrl });
		loading = false;

		if ('error' in result) {
			toast.error(result.message)
		} else {
			app!.callback = callbackUrl
			editCallbackURLModal?.close();
		}
	}

	const createProvider = async (kind: string) => {
		loading = true
		const result = await star.createProvider(app!.id, kind);
		loading = false;

		if ('error' in result) {
			toast.error(result.message)
		} else {
			location.reload()
		}
	}
</script>

<Sidebar class="md:p-8 max-w-6xl mx-auto">
	{#if app}
		<div class="flex flex-col gap-4">
			<div class="flex flex-row items-center">
				<div>
					<h1 class="text-2xl font-semibold">Manage Application</h1>
					<h2 class="text-lg text-primary/60">
						Viewing application {app.name}
					</h2>
				</div>

				<Button
					variant="ghost"
					class="ml-auto"
					onclick={() => goto('/admin/apps')}
				>
					<Icon icon="chevron-left" />Go back
				</Button>
			</div>

			<div class="mt-16 flex flex-col gap-8">
				<div class="flex flex-col xl:flex-row items-center gap-4">
					<Icon class="text-3xl" icon="package" />
					<div
						class="flex flex-col items-center xl:items-start justify-center gap-2"
					>
						<h3 class="text-lg leading-4 flex items-center gap-2">
							{app.name}
							{#if app.internal}
								<Badge>Internal</Badge>
							{/if}
						</h3>
						<div
							class="flex flex-col xl:flex-row items-center xl:gap-3 text-sm text-primary/60 flex-wrap"
						>
							<span>
								<Icon icon="id" />
								{app.id}
							</span>
							<span>
								<Icon icon="tag" />
								{app.slug}
							</span>
							<span>
								<Icon icon="calendar-event" />
								Created {dateFormatter.format(new Date(app.created_at * 1000))}
							</span>
						</div>
					</div>

					{#if canModify}
						<DropdownMenu.Root>
							<DropdownMenu.Trigger class='ml-auto'>
								<Button variant='ghost' size='icon'>
									<Icon icon='dots' class='text-2xl' />
								</Button>
							</DropdownMenu.Trigger>
							<DropdownMenu.Content>
								<DropdownMenu.Group>
									<DropdownMenu.Item variant='destructive' onclick={() => deleteAppModal?.open()}>
										<Icon icon='trash' />
										Delete
									</DropdownMenu.Item>
								</DropdownMenu.Group>
							</DropdownMenu.Content>
						</DropdownMenu.Root>
					{/if}
				</div>

				{#if app.internal}
					<Card.Root class="w-full">
						<Card.Content
							class="flex flex-col overflow-hidden flex-wrap text-red-100"
						>
							<Card.Title class="text-red-200 mb-4">
								<Icon icon="lock" /> Internal Application
							</Card.Title>
							<span>
								This application is managed internally by star and can not be
								modified manually
							</span>
						</Card.Content>
					</Card.Root>
				{/if}
			</div>

			<Card.Root class="min-w-1/3 w-full md:w-auto">
				<Card.CardHeader>
					<Card.Title class="flex items-center gap-1">
						<Icon icon="package" class="text-2xl" /> Application Information
					</Card.Title>
				</Card.CardHeader>
				<Card.Content class="flex flex-row justify-between">
					<div class="flex flex-col gap-4 justify-center items-baseline flex-1">
						<Categorizied title='Application Name'>
							<p class="text-md">{app.name}</p>
						</Categorizied>
						<Categorizied title='Slug'>
							<code class="text-md mt-1">
								{app.slug}
								<Copy content={app.slug} />
							</code>
						</Categorizied>
						<Categorizied title='Application URL'>
							<a href={app.url} target="_blank" class="text-md text-blue-300">
								{app.url}
								<Icon icon="external-link" />
							</a>
						</Categorizied>
					</div>
					<div class="flex flex-col gap-4 justify-center items-baseline flex-1">
						<Categorizied title='Callback URL'>
							<div class="flex flex-item items-center gap-1">
								<code class="text-md mt-1 block">
									{#if currentCallbackUrl}
										{currentCallbackUrl}
										<Copy content={currentCallbackUrl} />
									{:else}
										Unspecified
									{/if}
								</code>
								{#if canModify}
									<Button size='icon' variant='ghost' onclick={() => editCallbackURLModal?.open()}>
										<Icon icon='pencil' />
									</Button>
								{/if}
							</div>
						</Categorizied>
						<Categorizied title='Created At'>
							<p class="text-md">
								<Icon icon="clock" />
								{formatTime(app.created_at)}
							</p>
						</Categorizied>
						<Categorizied title='Updated At'>
							<p class="text-md">
								<Icon icon="clock" />
								{formatTime(app.updated_at)}
							</p>
						</Categorizied>
					</div>
				</Card.Content>
			</Card.Root>
			
			<Card.Root class="min-w-1/3 w-full md:w-auto">
				<Card.CardHeader>
					<Card.Title class="flex items-center gap-1">
						<Icon icon="package" class="text-2xl" /> Application Permissions
					</Card.Title>
					<Card.Description class='md:max-w-2/3'>Applications compatible with star can check whether a user is authorized perform actions based on their permissions. The <code>scopes</code> claim in access tokens defines the scopes the token is authorized to perform.</Card.Description>
				</Card.CardHeader>
				<Card.Content class="flex flex-row justify-between">
					{#if permissions.length == 0}
						<div class="flex flex-col gap-4 items-center w-full my-12">
							<span class="text-xl font-semibold leading-2">No Permissions</span>
							<span class="text-primary/60 leading-5 m-0 -translate-y-1.5 text-center max-w-3/4">
								Leverage star's permission system to keep all access managed by star. Permissions are a great way to have consistent RBAC between all your applications.
							</span>
							{#if canModify}
								<Button>Create Permission</Button>
							{/if}
						</div>
					{:else}
						<div class="w-full grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-2">
							{#each permissions as permission (permission.id)}
								<div role='button' class='w-full h-16 py-1 px-3 border border-primary/30 rounded-md flex gap-2 hover:*:opacity-100'>
									<div class='flex flex-col items-baseline'>
										<span class="text-md overflow-hidden text-clip">{permission.scope}</span>
										<span class="text-xs text-primary/60 overflow-hidden text-clip">{permission.description}</span>
									</div>

									{#if canModify}
										<Button variant='destructive' size='icon' class='ml-auto my-auto opacity-0 transition-all'>
											<Icon icon='trash' class='text-lg' />
										</Button>
									{/if}
								</div>
							{/each}

							{#if canModify && permissions.length > 0}
								<Button class='w-full border border-primary/30 h-full hover:bg-primary/80! hover:text-secondary!' variant='ghost'>
									<Icon icon='plus' />
									<span class="text-md">Create</span>
								</Button>
							{/if}
						</div>
					{/if}
				</Card.Content>
			</Card.Root>

			<h3 class="text-xl flex items-center gap-1">
				<Icon icon="shield" />
				Providers
			</h3>

			{#if !hasProviders}
				<div class="flex flex-col gap-4 items-center w-full my-32">
					<span class="text-xl font-semibold leading-2">No Providers</span>
					<span class="text-primary/60 leading-2">
						This application currently has no assigned providers
					</span>
					{#if canModify}
						<Button class="my-2" onclick={() => createProviderModal?.open()}>Create Provider</Button>
					{/if}
				</div>
			{/if}

			{#if app.oauth_provider}
				<OauthProvider modifiable={canModify} provider={app.oauth_provider} ondelete={() => app!.oauth_provider = null} />
			{/if}
		</div>
	{/if}
</Sidebar>

<Modal bind:this={createProviderModal} button class='xl:w-xl!'>
	<h1 class='text-xl font-semibold'>Create Provider</h1>
	<h2 class='text-card-foreground/60 sm:max-w-10/12'>Create a new provider for this application, allowing the app to authenticate against star using an authorization protocol</h2>

	<div class="flex flex-row gap-4 flex-wrap my-4">
		<CardChoice image={oauth2} class='w-[calc(33%-0.75rem)] h-48' recommended name='oauth2' onclick={() => createProvider('oauth')} />
	</div>
</Modal>

<Modal bind:this={deleteAppModal} button class='xl:w-xl!'>
	<h1 class='text-xl font-semibold'>Are you sure?</h1>
	<h2 class='text-card-foreground/60 sm:max-w-10/12'>This action can not be reversed! You will not be able to restore this application if it is deleted.</h2>

	<div class='flex items-center gap-2 justify-end mt-4'>
		<Button variant='secondary' onclick={() => deleteAppModal?.close()}>Cancel</Button>
		<Button variant='destructive' onclick={deleteApp} disabled={loading}>
			{#if loading}
				<Spinner />
			{/if}

			Delete Application
		</Button>
	</div>
</Modal>

<Modal bind:this={editCallbackURLModal} button class='xl:w-xl!'>
	<h1 class='text-xl font-semibold'>Change Callback URL</h1>
	<h2 class='text-card-foreground/60 sm:max-w-10/12'>Modify where star will redirect the user logging into the application, star will set the <code>access_token</code> and <code>refresh_token</code> query parameters once redirected successfully</h2>

	<form onsubmit={updateCallbackURL} class='mt-6'>
		<div class="flex flex-col gap-6 w-full mx-auto">
			<div class="grid gap-2">
				<Label for="callback_url">Callback URL</Label>
				<Input id="callback_url" type="text" placeholder='' bind:value={callbackUrl} required />
			</div>
		</div>

		<div class='flex items-center gap-2 justify-end mt-6'>
			<Button variant='secondary' onclick={() => editCallbackURLModal?.close()}>Cancel</Button>
			<Button onclick={updateCallbackURL}>Change Callback</Button>
		</div>
	</form>
</Modal>