<script lang="ts">
	import Icon from '@/components/icon.svelte';
	import { Badge } from '@/components/ui/badge';

	import * as Card from "@/components/ui/card";
	import { titleCase } from '@/utils';
	import type { ClassValue } from 'svelte/elements';

	const props: {
		name: string
		recommended?: boolean
		class?: ClassValue
		image?: string
		onclick?: () => void
	} = $props();
</script>

<Card.Root role='button' onclick={props.onclick} class='cursor-pointer hover:bg-secondary transition-colors {props.class ?? ''}'>
	<Card.Content class='flex flex-col items-center justify-center px-6 gap-2 select-none'>
		<span class='w-full text-center'>{titleCase(props.name)}</span>
		
		{#if props.image && (props.image.includes('://') || props.image.startsWith('/'))}
			<img src={props.image} alt={props.name} class='w-16 h-16' />
		{:else if props.image}
			<Icon icon={props.image} class='w-16 h-16 text-6xl mx-auto' />
		{/if}

		{#if props.recommended}
			<Badge variant='outline' class='border-emerald-600/60 text-emerald-500 mt-2'>Recommended</Badge>
		{/if}
	</Card.Content>
</Card.Root>