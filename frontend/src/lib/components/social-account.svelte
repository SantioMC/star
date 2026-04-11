<script lang="ts">
	import Icon from "@/components/icon.svelte";
	import * as star from "@/star";
	import { titleCase } from "@/utils";
	import type { ClassValue } from "svelte/elements";

	const props: {
		provider: star.SocialProvider
		account?: star.Account|null
		class?: ClassValue
		modifiable?: boolean
	} = $props()
	
	let loginUrl = $derived(star.getLoginUrl(props.provider.name))

	const onclick = () => {
		if (!props.modifiable) return;
		location.href = loginUrl;
	}
</script>

<button class='flex items-center h-16 basis-[calc(50%-0.25rem)] gap-1 border {props.account ? 'border-green-500/20' : 'border-border'} p-4 w-full rounded {props.account ? 'not-disabled:hover:border-green-700/20' : 'not-disabled:hover:border-border/30'} transition-colors not-disabled:cursor-pointer {props.class ?? ''}' {onclick} disabled={!props.modifiable}>
	{#if props.provider.icon}
		<Icon icon={props.provider.icon} class='text-lg' />
	{/if}

	{#if props.account}
		<div class="flex flex-col mx-1">
			<span class='leading-5'>{titleCase(props.provider.name)}</span>
			<span class='leading-5 text-sm text-primary/60 flex items-center gap-1'>
				{#if props.account.platform_avatar}
					<img aria-hidden='true' src={props.account.platform_avatar} alt='profile picture' class='w-4 h-4 rounded-full' />
				{/if}
				
				{props.account.username}
			</span>
		</div>
		<span class='ml-auto text-green-200/60'>Connected</span>
	{:else}
		{titleCase(props.provider.name)}
		<span class='ml-auto text-primary/60'>{props.modifiable ? `Connect` : 'Not Connected'}</span>
	{/if}
</button>