<script lang="ts">
	import Icon from '@/components/icon.svelte';
	import * as Item from '@/components/ui/item';
	import type { Snippet } from 'svelte';

	const {
		name,
		description,
		badges,
		actions
	}: {
		name: string;
		description: string;
		badges?: Snippet;
		actions?: Snippet;
	} = $props();
</script>

<Item.Root variant="outline">
	<Item.Content class="flex flex-row items-center gap-4">
		<Icon class="text-2xl" icon="lock-access" />
		<div class="min-w-0">
			<p class="text-base flex items-center gap-2 flex-wrap">
				<code class="truncate">{name}</code>

				{@render badges?.()}
			</p>

			<Item.Description class="text-xs">
				{description}
			</Item.Description>
		</div>

		{#if actions}
			<div class="ml-auto flex items-center gap-1.5">
				{@render actions()}
			</div>
		{/if}
	</Item.Content>
</Item.Root>
