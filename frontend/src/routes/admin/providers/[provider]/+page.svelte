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
	import { Input } from '@/components/ui/input';
	import { Label } from '@/components/ui/label';
	import { Spinner } from '@/components/ui/spinner';
	import { hasPermission } from '@/index';
	import { formatTime, titleCase } from '@/utils';
	import { onMount } from 'svelte';
	import { toast } from 'svelte-sonner';

	let loading = $state<boolean>(false);
	let provider = $state<star.SocialProviderConfig | null>(null);

	let clientId = $state<string>('');
	let clientSecret = $state<string>('');
	let order = $state<number>(0);
	let icon = $state<string>('');

	let deleteModal = $state<Modal>();
	let credentialsModal = $state<Modal>();

	let canModify = $derived(hasPermission('admin:provider:write'));

	onMount(async () => {
		if (!browser) return;
		if (!hasPermission('admin:provider:read')) return goto('/');

		provider = await star.getSocialProvider(page.params.provider!);
		if (provider == null) return goto('/admin/providers');

		order = provider.order;
		icon = provider.icon;
	});

	const updateProvider = async () => {
		loading = true;
		const result = await star.updateSocialProvider(provider!.id, {
			order,
			icon: icon || undefined
		});
		loading = false;

		if ('error' in result) {
			toast.error(result.message);
		} else {
			provider = result;
			toast.success('Successfully updated provider');
		}
	};

	const updateCredentials = async () => {
		loading = true;

		const changes: star.UpdateSocialProviderRequest = {};
		if (clientId) changes.client_id = clientId;
		if (clientSecret) changes.client_secret = clientSecret;

		const result = await star.updateSocialProvider(provider!.id, changes);
		loading = false;

		if ('error' in result) {
			toast.error(result.message);
		} else {
			provider = result;
			clientId = '';
			clientSecret = '';
			credentialsModal?.close();
			toast.success('Successfully updated credentials');
		}
	};

	const deleteProvider = async () => {
		loading = true;
		const result = await star.deleteSocialProvider(provider!.id);
		loading = false;
		deleteModal?.close();

		if ('error' in result) {
			toast.error(result.message);
		} else {
			goto('/admin/providers');
		}
	};
</script>

<Sidebar class="md:p-8 max-w-6xl mx-auto">
	{#if provider}
		<div class="flex flex-row items-center">
			<div>
				<h1 class="text-2xl font-semibold">Manage Provider</h1>
				<h2 class="text-lg text-primary/60">
					Viewing {titleCase(provider.name)} social login provider
				</h2>
			</div>

			<Button
				variant="ghost"
				class="ml-auto"
				onclick={() => goto('/admin/providers')}
			>
				<Icon icon="chevron-left" />Go back
			</Button>
		</div>

		<div class="mt-16 flex flex-col gap-8">
			<div class="flex flex-col xl:flex-row items-center gap-4">
				<Icon class="text-3xl" icon={provider.icon || 'plug-connected'} />
				<div class="flex flex-col items-center xl:items-start justify-center gap-2">
					<h3 class="text-lg leading-4 flex items-center gap-2">
						{titleCase(provider.name)}
						{#if provider.oidc_url}
							<Badge variant="secondary">OIDC Discovery</Badge>
						{/if}
					</h3>
					<div class="flex flex-col xl:flex-row items-center xl:gap-3 text-sm text-primary/60 flex-wrap">
						<span>
							<Icon icon="id" />
							{provider.id}
						</span>
						<span>
							<Icon icon="calendar-event" />
							Created {formatTime(new Date(provider.created_at))}
						</span>
						<span>
							<Icon icon="sort-ascending" />
							Order: {provider.order}
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
								<DropdownMenu.Item onclick={() => credentialsModal?.open()}>
									<Icon icon="key" />
									Update Credentials
								</DropdownMenu.Item>
							</DropdownMenu.Group>
							{#if hasPermission('admin:provider:delete')}
								<DropdownMenu.Separator />
								<DropdownMenu.Group>
									<DropdownMenu.Item
										variant="destructive"
										onclick={() => deleteModal?.open()}
									>
										<Icon icon="trash" />
										Delete
									</DropdownMenu.Item>
								</DropdownMenu.Group>
							{/if}
						</DropdownMenu.Content>
					</DropdownMenu.Root>
				{/if}
			</div>

			<Card.Root class="min-w-1/3 w-full md:w-auto">
				<Card.CardHeader>
					<Card.Title class="flex items-center gap-1">
						<Icon icon="plug-connected" class="text-2xl" /> Provider Information
					</Card.Title>
				</Card.CardHeader>
				<Card.Content class="flex flex-row justify-between">
					<div class="flex flex-col gap-4 justify-center items-baseline flex-1">
						<Categorizied title='Name'>
							<p class="text-md">{titleCase(provider.name)}</p>
						</Categorizied>
						<Categorizied title='Client ID'>
							<code class="text-md mt-1 flex items-center gap-1">
								{provider.client_id}
								<Copy content={provider.client_id} />
							</code>
						</Categorizied>
						<Categorizied title='Client Secret'>
							<Hidden content={provider.client_secret} />
						</Categorizied>
					</div>
					<div class="flex flex-col gap-4 justify-center items-baseline flex-1">
						<Categorizied title='Created At'>
							<p class="text-md">
								<Icon icon="clock" />
								{formatTime(new Date(provider.created_at))}
							</p>
						</Categorizied>
						{#if provider.scopes && provider.scopes.length > 0}
							<Categorizied title='Scopes'>
								<div class="flex gap-1 flex-wrap">
									{#each provider.scopes as scope}
										<Badge variant="secondary">{scope}</Badge>
									{/each}
								</div>
							</Categorizied>
						{/if}
					</div>
				</Card.Content>
			</Card.Root>

			<Card.Root class="min-w-1/3 w-full md:w-auto">
				<Card.CardHeader>
					<Card.Title class="flex items-center gap-1">
						<Icon icon="link" class="text-2xl" /> Endpoints
					</Card.Title>
					<Card.Description>The OAuth/OIDC endpoints used for this provider</Card.Description>
				</Card.CardHeader>
				<Card.Content class="flex flex-col gap-4">
					{#if provider.oidc_url}
						<Categorizied title='Discovery URL'>
							<code class="text-md mt-1 flex items-center gap-1">
								{provider.oidc_url}
								<Copy content={provider.oidc_url} />
							</code>
						</Categorizied>
					{/if}
					{#if provider.authorization_url}
						<Categorizied title='Authorization URL'>
							<code class="text-md mt-1 flex items-center gap-1">
								{provider.authorization_url}
								<Copy content={provider.authorization_url} />
							</code>
						</Categorizied>
					{/if}
					{#if provider.token_url}
						<Categorizied title='Token URL'>
							<code class="text-md mt-1 flex items-center gap-1">
								{provider.token_url}
								<Copy content={provider.token_url} />
							</code>
						</Categorizied>
					{/if}
					{#if provider.userinfo_url}
						<Categorizied title='User Info URL'>
							<code class="text-md mt-1 flex items-center gap-1">
								{provider.userinfo_url}
								<Copy content={provider.userinfo_url} />
							</code>
						</Categorizied>
					{/if}
				</Card.Content>
			</Card.Root>

			<Card.Root class="min-w-1/3 w-full md:w-auto">
				<Card.CardHeader>
					<Card.Title class="flex items-center gap-1">
						<Icon icon="settings" class="text-2xl" /> Display Settings
					</Card.Title>
					<Card.Description>Configure how this provider appears on the login page</Card.Description>
				</Card.CardHeader>
				<Card.Content class="flex flex-col gap-6">
					<div class="grid gap-2">
						<Label for="icon">Icon</Label>
						<div class="flex items-center gap-2">
							<Icon icon={icon || 'plug-connected'} class="text-2xl" />
							<Input id="icon" type="text" placeholder="brand-discord" bind:value={icon} disabled={!canModify} />
						</div>
						<p class="text-sm text-primary/50">Uses Tabler icon names</p>
					</div>
					<div class="grid gap-2">
						<Label for="order">Display Order</Label>
						<Input id="order" type="number" bind:value={order} disabled={!canModify} />
						<p class="text-sm text-primary/50">Lower numbers appear first on the login page</p>
					</div>
				</Card.Content>
				{#if canModify}
					<Card.Footer>
						<Button class="ml-auto" onclick={updateProvider} disabled={loading}>
							{#if loading}
								<Spinner />
							{/if}
							Update
						</Button>
					</Card.Footer>
				{/if}
			</Card.Root>
		</div>
	{/if}
</Sidebar>

<Modal bind:this={deleteModal} button class="xl:w-xl!">
	<h1 class="text-xl font-semibold">Are you sure?</h1>
	<h2 class="text-card-foreground/60 sm:max-w-10/12">
		This will remove the {titleCase(provider?.name ?? '')} social login provider. Users who signed up exclusively through this provider will no longer be able to log in.
	</h2>

	<div class="flex items-center gap-2 justify-end mt-4">
		<Button variant="secondary" onclick={() => deleteModal?.close()}>
			Cancel
		</Button>
		<Button variant="destructive" onclick={deleteProvider} disabled={loading}>
			{#if loading}
				<Spinner />
			{/if}
			Delete Provider
		</Button>
	</div>
</Modal>

<Modal bind:this={credentialsModal} button class="xl:w-xl!">
	<h1 class="text-xl font-semibold">Update Credentials</h1>
	<h2 class="text-card-foreground/60 sm:max-w-10/12">
		Update the OAuth client credentials for this provider. Leave a field empty to keep the current value.
	</h2>

	<form onsubmit={updateCredentials} class="mt-6">
		<div class="flex flex-col gap-6 w-full mx-auto">
			<div class="grid gap-2">
				<Label for="newClientId">Client ID</Label>
				<Input id="newClientId" type="text" placeholder="Leave empty to keep current" bind:value={clientId} />
			</div>
			<div class="grid gap-2">
				<Label for="newClientSecret">Client Secret</Label>
				<Input id="newClientSecret" type="password" placeholder="Leave empty to keep current" bind:value={clientSecret} />
			</div>
		</div>

		<div class="flex items-center gap-2 justify-end mt-6">
			<Button variant="secondary" onclick={() => credentialsModal?.close()}>
				Cancel
			</Button>
			<Button disabled={loading} type="submit">
				{#if loading}
					<Spinner />
				{/if}
				Update Credentials
			</Button>
		</div>
	</form>
</Modal>
