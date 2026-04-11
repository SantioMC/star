<script lang="ts">
	import type { Snippet } from "svelte";
	import type { ClassValue } from "svelte/elements";

	const props: {
		title: string
		children?: Snippet,
		class?: ClassValue,
	} = $props()
</script>

<div class='flex flex-col gap-1 items-baseline {props.class ?? ''}'>
	<p
		class="text-sm text-primary/70 select-none font-semibold uppercase"
	>
		{props.title}
	</p>
	{@render props.children?.()}
</div>