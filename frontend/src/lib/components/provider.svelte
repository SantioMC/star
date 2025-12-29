<script lang="ts">
	import * as star from '$lib/star';
	import Categorizied from '@/components/categorizied.svelte';
	import Copy from '@/components/copy.svelte';
	import Hidden from '@/components/hidden.svelte';
	import Icon from '@/components/icon.svelte';
	import Modal from '@/components/modal.svelte';
	import { Badge } from '@/components/ui/badge';
	import { Button } from '@/components/ui/button';
	import * as Card from '@/components/ui/card';
	import * as DropdownMenu from '@/components/ui/dropdown-menu';
	import { Spinner } from '@/components/ui/spinner';
	import { dateFormatter, titleCase } from '@/utils';
	import { toast } from 'svelte-sonner';
	import { } from './../star.ts';

	const { 
		provider,
		modifiable,
		ondelete
	}: {
		provider: star.AdminProvider
		modifiable?: boolean
		ondelete?: () => void
	} = $props()

	let deleteModal = $state<Modal>()

	let loading = $state<boolean>(false);
	let secret = $derived(provider.secret)

	const deleteProvider = async () => {
		loading = true;
		const result = await star.deleteProvider(provider.application_id, provider.id);
		deleteModal?.close();
		loading = false;

		if ('error' in result) {
			toast.error(result.message)
		} else {
			ondelete?.()
		}
	}

	const rotateProvider = async () => {
		const result = await star.rotateProvider(provider.application_id, provider.id)
		if ('error' in result) {
			toast.error(result.message)
		} else {
			secret = result.secret
			toast.success('Successfully rotated provider secret')
		}
	}
</script>

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
			{#if modifiable}
				<div class="ml-auto">
					<DropdownMenu.Root>
						<DropdownMenu.Trigger class='ml-auto'>
							<Button variant='ghost' size='icon'>
								<Icon icon='dots' class='text-2xl' />
							</Button>
						</DropdownMenu.Trigger>
						<DropdownMenu.Content>
							<DropdownMenu.Group>
								<DropdownMenu.Item onclick={() => rotateProvider()}>
									<Icon icon='rotate' />
									Rotate Secret
								</DropdownMenu.Item>
							</DropdownMenu.Group>
							<DropdownMenu.Separator />
							<DropdownMenu.Group>
								<DropdownMenu.Item variant='destructive' onclick={() => deleteModal?.open()}>
									<Icon icon='trash' />
									Delete
								</DropdownMenu.Item>
							</DropdownMenu.Group>
						</DropdownMenu.Content>
					</DropdownMenu.Root>
				</div>
			{/if}
		</Card.Title>
	</Card.CardHeader>
	<Card.Content class="flex flex-row justify-between items-stretch">
		<div
			class="flex flex-col gap-4 justify-center items-baseline flex-1"
		>
			<Categorizied title='Client Secret'>
				<Hidden content={secret} />
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
					{#if modifiable}
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
						{dateFormatter.format(new Date(provider.created_at * 1000))}
					</p>
				</Categorizied>
				<Categorizied title='Updated At'>
					<p class="text-sm">
						<Icon icon="clock" />
						{dateFormatter.format(new Date(provider.updated_at * 1000))}
					</p>
				</Categorizied>
			</div>
		</div>
	</Card.Content>
</Card.Root>

<Modal bind:this={deleteModal} button class='xl:w-xl!'>
	<h1 class='text-xl font-semibold'>Are you sure?</h1>
	<h2 class='text-card-foreground/60 sm:max-w-10/12'>This action can not be reversed! You will not be able to restore this provider if it is deleted.</h2>

	<div class='flex items-center gap-2 justify-end mt-4'>
		<Button variant='secondary' onclick={() => deleteModal?.close()}>Cancel</Button>
		<Button variant='destructive' onclick={deleteProvider} disabled={loading}>
			{#if loading}
				<Spinner />
			{/if}

			Delete Provider
		</Button>
	</div>
</Modal>