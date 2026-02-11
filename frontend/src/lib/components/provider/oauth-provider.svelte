<script lang="ts">
	import * as star from '$lib/star';
	import Categorizied from '@/components/categorizied.svelte';
	import Copy from '@/components/copy.svelte';
	import Hidden from '@/components/hidden.svelte';
	import Icon from '@/components/icon.svelte';
	import Modal from '@/components/modal.svelte';
	import RedirectUri from '@/components/provider/redirect-uri.svelte';
	import { Badge } from '@/components/ui/badge';
	import { Button, buttonVariants } from '@/components/ui/button';
	import * as Card from '@/components/ui/card';
	import * as DropdownMenu from '@/components/ui/dropdown-menu';
	import { Input } from '@/components/ui/input';
	import * as Sheet from '@/components/ui/sheet';
	import { Spinner } from '@/components/ui/spinner';
	import { dateFormatter, titleCase } from '@/utils';
	import { toast } from 'svelte-sonner';

	const {
		provider,
		modifiable,
		ondelete
	}: {
		provider: star.OAuthProvider;
		modifiable?: boolean;
		ondelete?: () => void;
	} = $props();

	let deleteModal = $state<Modal>();
	let addRedirectUriModal = $state<Modal>();
	let fingerprintModal = $state<Modal>();

	let redirectUri = $state<string>('');

	let loading = $state<boolean>(false);
	let showInformation = $state<boolean>(false);
	let secret = $derived(provider.client_secret);

	// svelte-ignore state_referenced_locally
	let redirects = $state<string[]>(provider.redirect_uris);

	const deleteProvider = async () => {
		loading = true;
		const result = await star.deleteProvider(
			provider.application_id,
			provider.id
		);

		deleteModal?.close();
		loading = false;

		if ('error' in result) {
			toast.error(result.message);
		} else {
			ondelete?.();
		}
	};

	const rotateProvider = async () => {
		const result = await star.rotateProvider(
			provider.application_id,
			provider.id
		);

		if ('error' in result) {
			toast.error(result.message);
			return;
		}

		secret = result.secret;
		toast.success('Successfully rotated provider secret');
	};

	const toggleFingerprinting = async () => {
		const result = await star.modifyProvider(
			provider.application_id,
			provider.id,
			{
				fingerprinting: !provider!.fingerprinting
			}
		);

		fingerprintModal?.close();
		if ('error' in result) {
			toast.error(result.message);
			return;
		}

		provider!.fingerprinting = !provider!.fingerprinting;
		toast.success('Successfully toggled fingerprinting');
	};

	const addRedirectUri = async (e: Event) => {
		e.preventDefault();
		loading = true;

		const result = await star.addRedirectUri(
			provider.application_id,
			provider.id,
			redirectUri
		);

		loading = false;
		if ('error' in result) {
			toast.error(result.message);
			return;
		}

		redirects.push(redirectUri);
		redirectUri = '';
		addRedirectUriModal?.close();

		toast.success('Successfully added redirect uri');
	};

	const removeRedirectUri = async (index: number) => {
		loading = true;

		const result = await star.removeRedirectUri(
			provider.application_id,
			provider.id,
			index
		);

		loading = false;
		if ('error' in result) {
			toast.error(result.message);
			return;
		}

		redirects.splice(index, 1);
		addRedirectUriModal?.close();

		toast.success('Successfully removed redirect uri');
	};
</script>

<Card.Root class="min-w-1/3 w-full md:w-auto">
	<Card.CardHeader>
		<Card.Title class="flex flex-row gap-2 items-center">
			<Icon icon="key" class="text-4xl" />
			<div class="flex flex-col gap-0.5">
				<div class="flex flex-row items-center">
					OAuth2
					<Badge
						variant="outline"
						class="border-emerald-600/60 text-emerald-500 mx-2"
					>
						Active
					</Badge>
				</div>
				<span class="text-primary/70 text-xs">{provider.id}</span>
			</div>
			<div class="ml-auto">
				<DropdownMenu.Root>
					<DropdownMenu.Trigger class="ml-auto">
						<Button variant="ghost" size="icon">
							<Icon icon="dots" class="text-2xl" />
						</Button>
					</DropdownMenu.Trigger>
					<DropdownMenu.Content>
						<DropdownMenu.Group>
							<DropdownMenu.Item onclick={() => (showInformation = true)}>
								<Icon icon="info-circle" />
								View Information
							</DropdownMenu.Item>
							{#if modifiable}
								<DropdownMenu.Item onclick={() => fingerprintModal?.open()}>
									<Icon icon="fingerprint" />
									{provider.fingerprinting ? 'Disable' : 'Enable'}
									Fingerprinting
								</DropdownMenu.Item>
								<DropdownMenu.Item onclick={() => rotateProvider()}>
									<Icon icon="rotate" />
									Rotate Secret
								</DropdownMenu.Item>
							{/if}
						</DropdownMenu.Group>
						{#if modifiable}
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
			</div>
		</Card.Title>
	</Card.CardHeader>
	<Card.Content class="flex flex-row justify-between items-stretch">
		<div class="flex flex-col gap-4 justify-center items-baseline flex-1">
			<Categorizied title="Client ID">
				<code class="w-auto! gap-2 flex! items-center">
					<Icon icon="tag" />{provider.client_id}
					<Copy content={provider.client_id} class="ml-auto" />
				</code>
			</Categorizied>
			{#if provider.application_type == 'confidential'}
				<Categorizied title="Client Secret">
					<Hidden content={secret} />
				</Categorizied>
			{:else}
				<Categorizied title="Client Secret">
					<span class="text-primary/60 max-w-3/4 select-none">
						This is not a confidential application, so it has no client secret.
					</span>
				</Categorizied>
			{/if}
			<Categorizied title="Application Type">
				<Badge variant="outline" class="border-blue-400/60 text-blue-300">
					{titleCase(provider.application_type)}
				</Badge>
			</Categorizied>
			<Categorizied title="Session Handling">
				<Badge variant="outline" class="border-orange-400/60 text-orange-300">
					{provider.rotate_refresh_token ? 'Rotating Sessions' : 'Static'}
				</Badge>
			</Categorizied>
			<Categorizied title="Fingerprinting">
				{#if provider.fingerprinting}
					<Badge variant="outline" class="border-green-400/60 text-green-300">
						Enabled
					</Badge>
				{:else}
					<Badge variant="outline" class="border-red-400/60 text-red-300">
						Disabled
					</Badge>
				{/if}
			</Categorizied>
			<Categorizied title="Response Types">
				<div class="flex flex-row items-center gap-1 flex-wrap">
					{#each provider.response_types as responseType}
						<Badge variant="outline">{responseType}</Badge>
					{/each}
				</div>
			</Categorizied>
		</div>
		<div class="flex flex-col gap-4 items-baseline flex-1 h-auto">
			<Categorizied title="Redirect URIs" class="w-full">
				<div class="flex flex-col w-full gap-1">
					{#each redirects as redirectUri, index}
						<RedirectUri
							{redirectUri}
							onremove={() => removeRedirectUri(index)}
						/>
					{/each}

					{#if modifiable}
						<Button
							variant="secondary"
							onclick={() => addRedirectUriModal?.open()}
						>
							<Icon icon="plus" />
							Add Redirect URI
						</Button>
					{:else if redirects.length == 0}
						<div
							class="flex flex-col gap-4 items-center w-full my-24 select-none"
						>
							<span class="text-xl font-semibold text-primary/60 text-center">
								No Redirects Configured
							</span>
							<span class="text-primary/40 text-center">
								This application has no redirects configured
							</span>
						</div>
					{/if}
				</div>
			</Categorizied>
			<hr class="mt-auto!" />
			<div class="flex items-center justify-between w-full">
				<Categorizied title="Created At">
					<p class="text-sm">
						<Icon icon="clock" />
						{dateFormatter.format(new Date(provider.created_at * 1000))}
					</p>
				</Categorizied>
				<Categorizied title="Updated At">
					<p class="text-sm">
						<Icon icon="clock" />
						{dateFormatter.format(new Date(provider.updated_at * 1000))}
					</p>
				</Categorizied>
			</div>
		</div>
	</Card.Content>
</Card.Root>

<Modal bind:this={deleteModal} button class="xl:w-xl!">
	<h1 class="text-xl font-semibold">Are you sure?</h1>
	<h2 class="text-card-foreground/60 sm:max-w-10/12">
		This action can not be reversed! You will not be able to restore this
		provider if it is deleted.
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

<Modal bind:this={addRedirectUriModal} button class="xl:w-xl!">
	<h1 class="text-xl font-semibold">Add Redirect URI</h1>
	<h2 class="text-card-foreground/60 sm:max-w-10/12">
		Add a new redirect URI for this provider, this will allow the application at
		the specified URI to receive the authorization code after the user has
		authorized the application.
	</h2>

	<form
		class="flex flex-col items-center gap-6 justify-end mt-4"
		onsubmit={addRedirectUri}
	>
		<Input
			bind:value={redirectUri}
			placeholder="https://example.com/callback"
			required
		/>

		<div class="flex gap-2 mr-2 w-full">
			<Button
				class="ml-auto"
				variant="secondary"
				onclick={() => addRedirectUriModal?.close()}
			>
				Cancel
			</Button>
			<Button type="submit" disabled={loading}>
				{#if loading}
					<Spinner />
				{/if}

				Add Redirect URI
			</Button>
		</div>
	</form>
</Modal>

<Modal bind:this={fingerprintModal} button class="xl:w-xl!">
	<h1 class="text-xl font-semibold">
		{provider.fingerprinting ? 'Disable' : 'Enable'} Fingerprinting
	</h1>
	<h2 class="text-card-foreground/60 sm:max-w-10/12">
		Fingerprinting is a star-specific feature that allows you to pass a
		fingerprint of the user's client to star to identify them by their hardware
		configuration. This is not used for allowing users to login, but rather an
		additional layer of security to make session hijacking attacks more
		difficult. To use this feature, pass <code>fp</code>
		as a parameter to the
		<code>/oauth/token</code>
		endpoint.
	</h2>

	<form
		class="flex flex-col items-center gap-6 justify-end mt-4"
		onsubmit={toggleFingerprinting}
	>
		<div class="flex gap-2 mr-2 w-full">
			<Button
				class="ml-auto"
				variant="secondary"
				onclick={() => fingerprintModal?.close()}
			>
				Cancel
			</Button>
			<Button type="submit" disabled={loading}>
				{#if loading}
					<Spinner />
				{/if}

				{provider.fingerprinting ? 'Disable' : 'Enable'} Fingerprinting
			</Button>
		</div>
	</form>
</Modal>

<Sheet.Root bind:open={showInformation}>
	<Sheet.Content side="right" class="max-w-xl!">
		<Sheet.Header>
			<Sheet.Title>Provider Information</Sheet.Title>
			<Sheet.Description>
				View necessary information for setting up this OAuth2 provider to allow
				users to sign in with star.
			</Sheet.Description>
		</Sheet.Header>

		<div class="px-2 flex flex-col gap-2">
			<Categorizied title="Client ID">
				<code class="gap-2 flex! items-center">
					{provider.client_id}
					<Copy content={provider.client_id} class="ml-auto" />
				</code>
			</Categorizied>
			{#if provider.application_type == 'confidential'}
				<Categorizied title="Client Secret">
					<Hidden content={secret} class="w-full!" />
				</Categorizied>
			{:else}
				<Categorizied title="Client Secret">
					<span class="text-primary/60 max-w-3/4 select-none">
						This is not a confidential application, so it has no client secret.
					</span>
				</Categorizied>
			{/if}
			<Categorizied title="Discovery URL">
				<code class="gap-2 flex! items-center">
					{star.getDiscoveryURL()}
					<Copy content={star.getDiscoveryURL()} class="ml-auto" />
				</code>
			</Categorizied>
			<Categorizied title="Authorization URL">
				<code class="gap-2 flex! items-center">
					{star.getAuthUrl()}
					<Copy content={star.getAuthUrl()} class="ml-auto" />
				</code>
			</Categorizied>
			<Categorizied title="Token URL">
				<code class="gap-2 flex! items-center">
					{star.getTokenUrl()}
					<Copy content={star.getTokenUrl()} class="ml-auto" />
				</code>
			</Categorizied>
			<Categorizied title="Userinfo URL">
				<code class="gap-2 flex! items-center">
					{star.getUserinfoUrl()}
					<Copy content={star.getUserinfoUrl()} class="ml-auto" />
				</code>
			</Categorizied>
		</div>

		<Sheet.Footer>
			<Sheet.Close class={buttonVariants({ variant: 'default' })}>
				Close
			</Sheet.Close>
		</Sheet.Footer>
	</Sheet.Content>
</Sheet.Root>
