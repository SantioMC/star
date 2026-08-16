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
	import RedirectUri from '@/components/provider/redirect-uri.svelte';
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

	let app = $state<star.AdminApplication | null>(null);

	let loading = $state<boolean>(false);
	let name = $state<string>('');
	let url = $state<string>('');
	let kind = $state<'public' | 'confidential'>('public');
	let developmentMode = $state<boolean>(false);
	let tokenDuration = $state<number>(3600);
	let sessionDuration = $state<number>(604800);
	let redirectUri = $state<string>('');

	let deleteAppModal = $state<Modal>();
	let rotateSecretModal = $state<Modal>();
	let addRedirectUriModal = $state<Modal>();

	let canModify = $derived(
		app?.internal == false && hasPermission('application:write')
	);

	onMount(async () => {
		if (!browser) return;
		if (!hasPermission('application:read')) return goto('/');

		app = await star.getApplication(page.params.app!);
		if (app == null) return goto('/admin/apps');

		name = app.name;
		url = app.application_url;
		kind = app.kind;
		developmentMode = app.development_mode;
		tokenDuration = app.token_duration;
		sessionDuration = app.session_duration;
	});

	const deleteApp = async () => {
		loading = true;
		const result = await star.deleteApplication(app!.id);
		deleteAppModal?.close();
		loading = false;

		if ('error' in result) {
			toast.error(result.message);
		} else {
			goto('/admin/apps');
		}
	};

	const refresh = async () => {
		app = await star.getApplication(app!.id);
	};

	const updateApp = async () => {
		loading = true;
		const result = await star.updateApplication(app!.id, {
			name,
			url,
			kind,
			development_mode: developmentMode,
			token_duration: tokenDuration,
			session_duration: sessionDuration
		});
		loading = false;

		if ('error' in result) {
			toast.error(result.message);
		} else {
			await refresh();
			toast.success('Successfully updated application');
		}
	};

	const rotateSecret = async () => {
		rotateSecretModal?.close();
		const result = await star.rotateApplicationSecret(app!.id);

		if ('error' in result) {
			toast.error(result.message);
		} else {
			await refresh();
			toast.success('Successfully rotated the client secret');
		}
	};

	const addRedirectUri = async (e: Event) => {
		e.preventDefault();

		const trimmed = redirectUri.trim();
		if (!trimmed) return;

		if (app!.redirect_uris.includes(trimmed)) {
			toast.error('This redirect URI is already registered');
			return;
		}

		loading = true;
		const result = await star.updateApplication(app!.id, {
			redirect_uris: [...app!.redirect_uris, trimmed]
		});
		loading = false;

		if ('error' in result) {
			toast.error(result.message);
		} else {
			await refresh();
			redirectUri = '';
			addRedirectUriModal?.close();
			toast.success('Successfully added redirect URI');
		}
	};

	const removeRedirectUri = async (uri: string) => {
		const result = await star.updateApplication(app!.id, {
			redirect_uris: app!.redirect_uris.filter((u) => u != uri)
		});

		if ('error' in result) {
			toast.error(result.message);
		} else {
			await refresh();
			toast.success('Removed redirect URI');
		}
	};
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
							<Badge variant="outline">{app.kind}</Badge>
							{#if app.development_mode}
								<Badge variant="outline" class="border-orange-400/60 text-orange-400">Development</Badge>
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
								<Icon icon="calendar-event" />
								Created {formatTime(new Date(app.created_at))}
							</span>
						</div>
					</div>

					{#if canModify && hasPermission('application:delete')}
						<DropdownMenu.Root>
							<DropdownMenu.Trigger class="ml-auto">
								<Button variant="ghost" size="icon">
									<Icon icon="dots" class="text-2xl" />
								</Button>
							</DropdownMenu.Trigger>
							<DropdownMenu.Content>
								<DropdownMenu.Group>
									<DropdownMenu.Item
										variant="destructive"
										onclick={() => deleteAppModal?.open()}
									>
										<Icon icon="trash" />
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

				<Card.Root class="min-w-1/3 w-full md:w-auto">
					<Card.CardHeader>
						<Card.Title class="flex items-center gap-1">
							<Icon icon="key" class="text-2xl" /> OAuth 2.0 Credentials
						</Card.Title>
						<Card.Description class="md:max-w-2/3">
							The credentials the application uses to identify itself against
							star's OAuth 2.0 endpoints.
						</Card.Description>
					</Card.CardHeader>
					<Card.Content class="flex flex-col gap-4">
						<Categorizied title="Client ID">
							<code class="text-md mt-1">
								{app.client_id}
								<Copy content={app.client_id} />
							</code>
						</Categorizied>

						{#if app.kind == 'confidential'}
							<Categorizied title="Client Secret">
								{#if app.client_secret}
									<Hidden content={app.client_secret} />
								{:else}
									<code class="text-md mt-1">Unset</code>
								{/if}
							</Categorizied>
						{:else}
							<Categorizied title="Client Secret">
								<p class="text-md text-primary/60">
									Public clients have no client secret and must use PKCE
								</p>
							</Categorizied>
						{/if}
					</Card.Content>
					{#if canModify && app.kind == 'confidential'}
						<Card.Footer>
							<Button
								variant="outline"
								class="ml-auto"
								onclick={() => rotateSecretModal?.open()}
							>
								<Icon icon="refresh" />
								Rotate Secret
							</Button>
						</Card.Footer>
					{/if}
				</Card.Root>

				<Card.Root class="min-w-1/3 w-full md:w-auto">
					<Card.CardHeader>
						<Card.Title class="flex items-center gap-1">
							<Icon icon="world" class="text-2xl" /> Redirect URIs
						</Card.Title>
						<Card.Description class="md:max-w-2/3">
							The exact redirect URIs the application is allowed to use during
							the authorization code flow. Redirects are matched exactly.
						</Card.Description>
					</Card.CardHeader>
					<Card.Content class="flex flex-col gap-2">
						{#if app.redirect_uris.length == 0}
							<div class="flex flex-col gap-4 items-center w-full my-12">
								<span class="text-xl font-semibold leading-2">
									No Redirect URIs
								</span>
								<span
									class="text-primary/60 leading-5 m-0 -translate-y-1.5 text-center max-w-3/4"
								>
									This application has no registered redirect URIs, the
									authorization flow will fail until one is added.
								</span>
								{#if canModify}
									<Button onclick={() => addRedirectUriModal?.open()}>
										Add Redirect URI
									</Button>
								{/if}
							</div>
						{:else}
							{#each app.redirect_uris as uri (uri)}
								{#if canModify}
									<RedirectUri
										redirectUri={uri}
										onremove={() => removeRedirectUri(uri)}
									/>
								{:else}
									<code class="w-auto! flex! items-center gap-2 px-2">
										<Icon icon="world" />
										{uri}
									</code>
								{/if}
							{/each}
						{/if}
					</Card.Content>
					{#if canModify && app.redirect_uris.length > 0}
						<Card.Footer>
							<Button
								class="ml-auto"
								onclick={() => addRedirectUriModal?.open()}
							>
								Add Redirect URI
							</Button>
						</Card.Footer>
					{/if}
				</Card.Root>

				<Card.Root class="min-w-1/3 w-full md:w-auto">
					<Card.CardHeader>
						<Card.Title class="flex items-center gap-1">
							<Icon icon="package" class="text-2xl" /> Application Settings
						</Card.Title>
					</Card.CardHeader>
					<Card.Content class="flex flex-col gap-6">
						<form class="flex flex-col gap-6">
							<div class="grid gap-2">
								<Label for="name">Name</Label>
								<Input
									id="name"
									type="text"
									required
									bind:value={name}
									disabled={!canModify}
								/>
							</div>

							<div class="grid gap-2">
								<Label for="url">Application URL</Label>
								<Input
									id="url"
									type="url"
									required
									bind:value={url}
									disabled={!canModify}
								/>
							</div>

							<div class="grid gap-2">
								<Label for="kind">Client Type</Label>
								<select
									id="kind"
									bind:value={kind}
									disabled={!canModify}
									class="border rounded-md bg-transparent px-3 py-2 text-sm"
								>
									<option value="public">
										Public (browser or native app, uses PKCE)
									</option>
									<option value="confidential">
										Confidential (server-side, uses a client secret)
									</option>
								</select>
							</div>

							<div class="flex gap-2 sm:gap-8">
								<div class="flex-1/2 gap-2 grid">
									<Label for="tokenDuration">Token Duration (seconds)</Label>
									<Input
										id="tokenDuration"
										type="number"
										min="60"
										bind:value={tokenDuration}
										disabled={!canModify}
									/>
								</div>

								<div class="flex-1/2 gap-2 grid">
									<Label for="sessionDuration">
										Session Duration (seconds)
									</Label>
									<Input
										id="sessionDuration"
										type="number"
										min="60"
										bind:value={sessionDuration}
										disabled={!canModify}
									/>
								</div>
							</div>

							<Label class="flex items-center gap-2">
								<input
									type="checkbox"
									bind:checked={developmentMode}
									disabled={!canModify}
								/>
								Development mode
							</Label>
						</form>
					</Card.Content>
					{#if canModify}
						<Card.Footer>
							<Button class="ml-auto" onclick={updateApp} disabled={loading}>
								{#if loading}
									<Spinner />
								{/if}
								Update
							</Button>
						</Card.Footer>
					{/if}
				</Card.Root>
			</div>
		</div>
	{/if}
</Sidebar>

<Modal bind:this={deleteAppModal} button class="xl:w-xl!">
	<h1 class="text-xl font-semibold">Are you sure?</h1>
	<h2 class="text-card-foreground/60 sm:max-w-10/12">
		This action can not be reversed! Deleting this application will invalidate
		all of its sessions and break any logins using it.
	</h2>

	<div class="flex items-center gap-2 justify-end mt-4">
		<Button variant="secondary" onclick={() => deleteAppModal?.close()}>
			Cancel
		</Button>
		<Button variant="destructive" onclick={deleteApp} disabled={loading}>
			{#if loading}
				<Spinner />
			{/if}

			Delete Application
		</Button>
	</div>
</Modal>

<Modal bind:this={rotateSecretModal} button class="xl:w-xl!">
	<h1 class="text-xl font-semibold">Rotate Client Secret</h1>
	<h2 class="text-card-foreground/60 sm:max-w-10/12">
		Rotating the client secret immediately invalidates the current secret, the
		application will fail to authenticate until it is updated with the new
		secret.
	</h2>

	<div class="flex items-center gap-2 justify-end mt-4">
		<Button variant="secondary" onclick={() => rotateSecretModal?.close()}>
			Cancel
		</Button>
		<Button variant="destructive" onclick={rotateSecret}>Rotate Secret</Button>
	</div>
</Modal>

<Modal bind:this={addRedirectUriModal} button class="xl:w-xl!">
	<h1 class="text-xl font-semibold">Add Redirect URI</h1>
	<h2 class="text-card-foreground/60 sm:max-w-10/12">
		Add an exact redirect URI the application may redirect users to after
		authorization.
	</h2>

	<form onsubmit={addRedirectUri} class="mt-6">
		<div class="flex flex-col gap-6 w-full mx-auto">
			<div class="grid gap-2">
				<Label for="redirect_uri">Redirect URI</Label>
				<Input
					id="redirect_uri"
					type="url"
					placeholder="https://example.com/callback"
					bind:value={redirectUri}
					required
				/>
			</div>
		</div>

		<div class="flex items-center gap-2 justify-end mt-6">
			<Button
				variant="secondary"
				onclick={() => addRedirectUriModal?.close()}
			>
				Cancel
			</Button>
			<Button disabled={loading} type="submit">
				{#if loading}
					<Spinner />
				{/if}
				Add Redirect URI
			</Button>
		</div>
	</form>
</Modal>
