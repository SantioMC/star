<script lang="ts">
	import Icon from "@/components/icon.svelte";
	import * as star from "@/star";
	import { titleCase } from "@/utils";
	import type { ClassValue } from "svelte/elements";

	const props: {
		type: star.DefaultSocialProvider
		account?: star.Account
		class?: ClassValue
	} = $props()

	const icons: { [key in star.DefaultSocialProvider]: string } = {
		'google': 'brand-google',
		'discord': 'brand-discord',
		'github': 'brand-github',
		'twitch': 'brand-twitch',
		'slack': 'brand-slack',
	}
</script>

<div class='flex items-center h-16 basis-[calc(50%-0.25rem)] gap-1 border {props.account ? 'border-green-500/20' : 'border-border'} p-4 w-full rounded {props.account ? 'hover:border-green-700/20' : 'hover:border-border/30'} transition-colors cursor-pointer {props.class ?? ''}'>
	<Icon icon={icons[props.type]} class='text-lg' />

	{#if props.account}
		<div class="flex flex-col mx-1">
			<span class='leading-5'>{titleCase(props.type)}</span>
			<span class='leading-5 text-sm text-primary/60 flex items-center gap-1'>
				{#if props.account.platform_avatar}
					<img aria-hidden='true' src={props.account.platform_avatar} alt='profile picture' class='w-4 h-4 rounded-full' />
				{/if}
				
				{props.account.username}
			</span>
		</div>
		<span class='ml-auto text-green-200/60'>Connected</span>
	{:else}
		{titleCase(props.type)}
		<span class='ml-auto text-primary/60'>Connect</span>
	{/if}
</div>