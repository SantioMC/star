<script lang="ts">
	import * as star from '$lib/star';
	import Icon from '@/components/icon.svelte';
	import Modal from '@/components/modal.svelte';
	import { Button } from '@/components/ui/button';
	import { Spinner } from '@/components/ui/spinner';
	import { toast } from 'svelte-sonner';

	const props: {
		permission: star.Permission
		modifiable?: boolean
		ondelete?: () => void
	} = $props()

	let deleteModal = $state<Modal>()
	let loading = $state<boolean>(false)

	const deletePermission = async () => {
		loading = true
		const result = await star.deletePermission(props.permission.application_id, props.permission.id)

		loading = false
		if ('error' in result) {
			return toast.error(result.message)
		}

		deleteModal?.close()
		props.ondelete?.()
	}
</script>

<div role='button' class='w-full h-16 py-1 px-3 border border-primary/30 rounded-md flex gap-2 hover:*:opacity-100'>
	<div class='flex flex-col items-baseline'>
		<span class="text-md overflow-hidden text-clip">{props.permission.scope}</span>
		<span class="text-xs text-primary/60 overflow-hidden text-clip">{props.permission.description}</span>
	</div>

	{#if props.modifiable}
		<Button variant='destructive' size='icon' class='ml-auto my-auto opacity-0 transition-all' onclick={() => deleteModal?.open()}>
			<Icon icon='trash' class='text-lg' />
		</Button>
	{/if}
</div>

<Modal bind:this={deleteModal} button class='xl:w-xl!'>
	<h1 class='text-xl font-semibold'>Are you sure?</h1>
	<h2 class='text-card-foreground/60 sm:max-w-10/12'>This action can not be reversed! You will not be able to restore this permission if it is deleted.</h2>

	<div class='flex items-center gap-2 justify-end mt-4'>
		<Button variant='secondary' onclick={() => deleteModal?.close()}>Cancel</Button>
		<Button variant='destructive' onclick={deletePermission} disabled={loading}>
			{#if loading}
				<Spinner />
			{/if}

			Delete Permission
		</Button>
	</div>
</Modal>