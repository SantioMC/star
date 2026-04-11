<script lang="ts">
	import { browser } from "$app/environment";
	import * as star from '$lib/star';
	import Icon from "@/components/icon.svelte";
	import Modal from "@/components/modal.svelte";
	import Sidebar from "@/components/sidebar.svelte";
	import { Badge } from "@/components/ui/badge";
	import { Button } from "@/components/ui/button";
	import { Spinner } from "@/components/ui/spinner";
	import * as Table from '@/components/ui/table';
	import { hasPermission } from "@/index";
	import { formatTime, relativeTime } from "@/utils";
	import { onMount } from "svelte";
	import { toast } from "svelte-sonner";

	let keys = $state<star.SigningKey[]>([])
	let loading = $state<boolean>(false)

	let rotateModal = $state<Modal>()
	let deleteModal = $state<Modal>()
	let deleteTarget = $state<star.SigningKey | null>(null)

	onMount(async () => {
		if (!browser) return
		if (!hasPermission('star:admin:read:keys')) return location.href = '/'
		keys = await star.getAllKeys()
	})

	const rotateKey = async () => {
		loading = true
		const result = await star.rotateSigningKey()
		loading = false
		rotateModal?.close()

		if ('error' in result) {
			toast.error(result.message)
		} else {
			keys = await star.getAllKeys()
			toast.success('Successfully rotated signing key')
		}
	}

	const confirmDelete = (key: star.SigningKey) => {
		deleteTarget = key
		deleteModal?.open()
	}

	const deleteKey = async () => {
		if (!deleteTarget) return

		loading = true
		const result = await star.deleteKey(deleteTarget.id)
		loading = false
		deleteModal?.close()

		if ('error' in result) {
			toast.error(result.message)
		} else {
			keys = keys.filter(k => k.id !== deleteTarget!.id)
			toast.success('Successfully deleted key')
			deleteTarget = null
		}
	}

	const isExpired = (key: star.SigningKey) => {
		return key.expires_at != null && key.expires_at < Math.floor(Date.now() / 1000)
	}
</script>

<Sidebar class='md:p-8 max-w-6xl mx-auto'>
	<h1 class='text-2xl font-semibold'>Signing Keys</h1>
	<h2 class='text-lg text-primary/60'>Manage the JWT signing keys used to sign access tokens. Star maintains two active keys to allow seamless rotation without requiring clients to immediately refetch the JWKS.</h2>

	<div class='mt-24'>
		<div class="flex flex-row items-center gap-2 my-2">
			{#if hasPermission('star:admin:create:keys')}
				<Button class='ml-auto' onclick={() => rotateModal?.open()}>Rotate Key</Button>
			{/if}
		</div>

		<Table.Root>
			<Table.Caption>Found {keys.length} key{keys.length === 1 ? '' : 's'}</Table.Caption>
			<Table.Header>
				<Table.Row>
				<Table.Head>Key</Table.Head>
				<Table.Head>Status</Table.Head>
				<Table.Head>Fingerprint</Table.Head>
				<Table.Head>Expires</Table.Head>
				<Table.Head>Created</Table.Head>
				{#if hasPermission('star:admin:delete:keys')}
					<Table.Head class='text-end'>Actions</Table.Head>
				{/if}
				</Table.Row>
			</Table.Header>
			<Table.Body>
				{#each keys as key (key.id)}
				<Table.Row>
					<Table.Cell class='flex flex-row gap-2 items-center'>
						<Icon class='text-3xl' icon='key' />
						<div class="flex flex-col items-start">
							<span>{key.id}</span>
							<span class='text-primary/70'>{key.name}</span>
						</div>
					</Table.Cell>
					<Table.Cell>
						{#if key.kind === 'signing' && !isExpired(key)}
							<Badge>Active</Badge>
						{:else if key.kind === 'next'}
							<Badge variant='secondary'>Next</Badge>
						{:else if key.kind === 'retired' && !isExpired(key)}
							<Badge variant='secondary'>Retired</Badge>
						{:else if isExpired(key)}
							<Badge variant='destructive'>Expired</Badge>
						{/if}
					</Table.Cell>
					<Table.Cell>
						<code class="text-sm">{key.fingerprint}</code>
					</Table.Cell>
					<Table.Cell>
						{#if key.expires_at}
							{#if isExpired(key)}
								<span class="text-destructive">Expired {relativeTime(key.expires_at)} ago</span>
							{:else}
								{formatTime(key.expires_at)} ({relativeTime(key.expires_at)})
							{/if}
						{:else}
							<span class="text-primary/50">No expiry</span>
						{/if}
					</Table.Cell>
					<Table.Cell>{formatTime(key.created_at)}</Table.Cell>
					{#if hasPermission('star:admin:delete:keys')}
						<Table.Cell class='w-0'>
							{#if key.kind !== 'signing' && key.kind !== 'next'}
								<Button variant='ghost' size='icon' class='text-destructive hover:text-destructive' onclick={() => confirmDelete(key)}>
									<Icon icon='trash' />
								</Button>
							{/if}
						</Table.Cell>
					{/if}
				</Table.Row>
				{/each}
			</Table.Body>
		</Table.Root>
	</div>
</Sidebar>

<Modal bind:this={rotateModal} button class="xl:w-xl!">
	<h1 class="text-xl font-semibold">Rotate Signing Key</h1>
	<h2 class="text-card-foreground/60 sm:max-w-10/12">
		This will promote the next key to the active signing key and generate a new next key. Existing tokens signed with the current key will remain valid until they expire. The JWKS endpoint will immediately reflect the change.
	</h2>

	<div class="flex items-center gap-2 justify-end mt-4">
		<Button variant="secondary" onclick={() => rotateModal?.close()}>
			Cancel
		</Button>
		<Button onclick={rotateKey} disabled={loading}>
			{#if loading}
				<Spinner />
			{/if}
			Rotate Key
		</Button>
	</div>
</Modal>

<Modal bind:this={deleteModal} button class="xl:w-xl!">
	<h1 class="text-xl font-semibold">Delete Key</h1>
	<h2 class="text-card-foreground/60 sm:max-w-10/12">
		Are you sure you want to delete this key? Any tokens signed with this key will no longer be verifiable via the JWKS endpoint.
	</h2>

	<div class="flex items-center gap-2 justify-end mt-4">
		<Button variant="secondary" onclick={() => deleteModal?.close()}>
			Cancel
		</Button>
		<Button variant="destructive" onclick={deleteKey} disabled={loading}>
			{#if loading}
				<Spinner />
			{/if}
			Delete Key
		</Button>
	</div>
</Modal>
