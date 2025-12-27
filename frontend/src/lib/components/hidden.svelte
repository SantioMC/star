<script lang="ts">
	import Copy from "@/components/copy.svelte";
	import Icon from "@/components/icon.svelte";

	let hidden = $state(true)

	const props: {
		content: string
	} = $props()
</script>

<div class="flex flex-row items-center gap-1">
	<code class="text-md mt-1">{hidden ? '•'.repeat(props.content.length) : props.content} 
		<Icon icon={hidden ? 'eye-off' : 'eye'} class='cursor-pointer' onclick={() => hidden = !hidden} />
		<Copy content={props.content} />
	</code>
</div>