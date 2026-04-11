<script lang="ts">
	import Icon from "$lib/components/icon.svelte";
	import type { ClassValue } from "svelte/elements";

	const props: { class?: ClassValue } = $props()
</script>

<span class="select-none font-sans {props.class}"><Icon class='text-yellow-400' filled icon='star'/> Star</span>