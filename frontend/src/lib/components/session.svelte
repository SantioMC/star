<script lang="ts">
	import * as star from '$lib/star';
	import Icon from '@/components/icon.svelte';
	import { Button } from '@/components/ui/button';
	import * as Item from '@/components/ui/item';
	import { Spinner } from '@/components/ui/spinner';
	import { getIcon, relativeTime } from '@/utils';
	import { toast } from 'svelte-sonner';

	let loading = $state<boolean>(false);

	const {
		session,
		oninvalidate
	}: {
		session: star.Session,
		oninvalidate?: () => void
	} = $props()

	const invalidate = async () => {
		loading = true;
		const result = await star.invalidateSession(session.id);

		if ('error' in result) {
			toast.error(result.message);
		} else {
			loading = false;
			oninvalidate?.()
		}
	}
</script>

<Item.Root variant='outline'>
	<Item.Content class='flex flex-row items-center gap-4'>
		<Icon class='text-3xl' icon={getIcon(session)} />
		<div>
			<p class="text-lg">
				{#if !session.device && !session.operating_system}
					Unknown
				{:else}
					{session.device} {session.operating_system}
				{/if}
			</p>

			<Item.Description class='text-xs'>IP Address: {session.ip_address} • Created {relativeTime(session.created_at)} ago</Item.Description>
		</div>
		<div class='ml-auto'>
			<Button variant='destructive' disabled={loading} onclick={invalidate}>
				{#if loading}
					<Spinner />
				{/if}
				
				Log out
			</Button>
		</div>
	</Item.Content>
</Item.Root>