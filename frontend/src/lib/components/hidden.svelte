<script lang="ts">
	import Copy from "@/components/copy.svelte";
	import Icon from "@/components/icon.svelte";
	import type { ClassValue } from "svelte/elements";

	let hidden = $state(true)

	const props: {
		content: string
		class?: ClassValue
	} = $props()
</script>

<div class="flex flex-row items-center gap-1 {props.class ?? ''}">
	<code class="text-md mt-1 flex items-center gap-1">
		<span>{hidden ? '•'.repeat(props.content.length) : props.content} </span>
		<Icon icon={hidden ? 'eye-off' : 'eye'} class='cursor-pointer' onclick={() => hidden = !hidden} />
		<Copy content={props.content} />
	</code>
</div>