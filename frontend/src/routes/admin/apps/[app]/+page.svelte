<script lang="ts">
	import { browser } from '$app/environment';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import * as star from '$lib/star';
	import Categorizied from '@/components/categorizied.svelte';
	import Copy from '@/components/copy.svelte';
	import Hidden from '@/components/hidden.svelte';
	import Icon from '@/components/icon.svelte';
	import Modal from '@/components/modal.svelte';
	import Sidebar from '@/components/sidebar.svelte';
	import { Badge } from '@/components/ui/badge';
	import { Button } from '@/components/ui/button';
	import * as Card from '@/components/ui/card';
	import * as DropdownMenu from '@/components/ui/dropdown-menu';
	import { Spinner } from '@/components/ui/spinner';
	import { hasPermission } from '@/index';
	import { dateFormatter, formatTime, titleCase } from '@/utils';
	import { onMount } from 'svelte';
	import { toast } from 'svelte-sonner';

	let app = $state<star.AdminApplication | null>(null);
	let loading = $state<boolean>(false);

	let deleteAppModal = $state<Modal>()

	let callbackUrl = $derived(
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
						<div>
							<p
								class="text-sm text-primary/70 select-none font-semibold uppercase"
							>
								Application Name
							</p>
							<p class="text-md">{app.name}</p>
						</div>
						<div>
							<p
								class="text-sm text-primary/70 select-none font-semibold uppercase"
							>
								Slug
							</p>
							<code class="text-md mt-1">
								{app.slug}
								<Copy content={app.slug} />
							</code>
						</div>
						<div>
							<p
								class="text-sm text-primary/70 select-none font-semibold uppercase"
							>
								Application URL
							</p>
							<a href={app.url} target="_blank" class="text-md text-blue-300">
								{app.url}
								<Icon icon="external-link" />
							</a>
						</div>
					</div>
					<div class="flex flex-col gap-4 justify-center items-baseline flex-1">
						<div>
							<p
								class="text-sm text-primary/70 select-none font-semibold uppercase"
							>
								Callback URL
							</p>
							<code class="text-md mt-1">
								{callbackUrl}
								<Copy content={callbackUrl} />
							</code>
						</div>
						<div>
							<p
								class="text-sm text-primary/70 select-none font-semibold uppercase"
							>
								Created At
							</p>
							<p class="text-md">
								<Icon icon="clock" />
								{formatTime(app.created_at)}
							</p>
						</div>
						<div>
							<p
								class="text-sm text-primary/70 select-none font-semibold uppercase"
							>
								Updated At
							</p>
							<p class="text-md">
								<Icon icon="clock" />
								{formatTime(app.updated_at)}
							</p>
						</div>
					</div>
				</Card.Content>
			</Card.Root>

			<h3 class="text-xl flex items-center gap-1">
				<Icon icon="shield" />
				Providers
			</h3>

			{#if app.providers.length == 0}
				<div class="flex flex-col gap-4 items-center w-full my-32">
					<span class="text-xl font-semibold leading-2">No Providers</span>
					<span class="text-primary/60 leading-2">
						This application currently has no assigned providers
					</span>
					{#if canModify}
						<Button class="my-2">Create Provider</Button>
					{/if}
				</div>
			{/if}

			{#each app.providers as provider (provider.id)}
				<Card.Root class="min-w-1/3 w-full md:w-auto">
					<Card.CardHeader>
						<Card.Title class="flex flex-row gap-2 items-center">
							<Icon icon="key" class="text-4xl" />
							<div class="flex flex-col gap-0.5">
								<div class="flex flex-row items-center">
									{titleCase(provider.type)}
									<Badge
										variant="outline"
										class="border-emerald-600/60 text-emerald-500 mx-2"
									>
										Active
									</Badge>
								</div>
								<span class="text-primary/70 text-xs">{provider.id}</span>
							</div>
						</Card.Title>
					</Card.CardHeader>
					<Card.Content class="flex flex-row justify-between items-stretch">
						<div
							class="flex flex-col gap-4 justify-center items-baseline flex-1"
						>
							<Categorizied title='Client Secret'>
								<Hidden content={provider.secret} />
							</Categorizied>
							<Categorizied title='Application Type'>
								<Badge
									variant="outline"
									class="border-blue-400/60 text-blue-300"
								>
									{titleCase(provider.application_type)}
								</Badge>
							</Categorizied>
							<Categorizied title='Auth Method'>
								<Badge
									variant="outline"
									class="border-purple-400/60 text-purple-300"
								>
									{provider.auth_method.toLowerCase()}
								</Badge>
							</Categorizied>
							<Categorizied title='Access Token Type'>
								<Badge
									variant="outline"
									class="border-orange-400/60 text-orange-300"
								>
									{titleCase(provider.access_token_type)}
								</Badge>
							</Categorizied>
							<Categorizied title='Response Types'>
								<div class="flex flex-row items-center gap-1 flex-wrap">
									{#each provider.response_types as responseType}
										<Badge variant="outline">{responseType}</Badge>
									{/each}
								</div>
							</Categorizied>
							<Categorizied title='Grant Types'>
								<div class="flex flex-row items-center gap-1 flex-wrap">
									{#each provider.grant_types as grantType}
										<Badge variant="outline">{grantType}</Badge>
									{/each}
								</div>
							</Categorizied>
						</div>
						<div
							class="flex flex-col gap-4 items-baseline flex-1 h-auto"
						>
							<Categorizied title='Redirect URIs' class='w-full'>
								<div class="flex flex-col w-full gap-1">
									{#each provider.redirect_uris as redirectUri}
										<code class='w-auto!'>
											<Icon icon='world' />
											{redirectUri}
											<Copy content={redirectUri} class='ml-auto' />
											<Icon icon='trash' class="hover:text-red-300 transition-colors" />
										</code>
									{/each}
									{#if canModify}
										<Button variant='secondary'>
											<Icon icon="plus" />
											Add Redirect URI
										</Button>
									{:else if provider.redirect_uris.length == 0}
										<div class="flex flex-col gap-4 items-center w-full my-24 select-none">
											<span class="text-xl font-semibold text-primary/60 text-center">No Redirects Configured</span>
											<span class="text-primary/40 text-center">
												This application has no redirects configured
											</span>
										</div>
									{/if}
								</div>
							</Categorizied>
							<hr class='mt-auto!' />
							<div class="flex items-center justify-between w-full">
								<Categorizied title='Created At'>
									<p class="text-sm">
										<Icon icon="clock" />
										{dateFormatter.format(new Date(app.created_at * 1000))}
									</p>
								</Categorizied>
								<Categorizied title='Updated At'>
									<p class="text-sm">
										<Icon icon="clock" />
										{dateFormatter.format(new Date(app.updated_at * 1000))}
									</p>
								</Categorizied>
							</div>
						</div>
					</Card.Content>
				</Card.Root>
			{/each}
		</div>
	{/if}
</Sidebar>

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