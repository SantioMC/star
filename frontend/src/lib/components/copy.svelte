<script lang="ts">
	import Icon from "@/components/icon.svelte";
	import { toast } from "svelte-sonner";
	import type { ClassValue } from "svelte/elements";

	const props: {
		content?: string
		toast?: boolean
		class?: ClassValue
	} = $props()

	let clicked = $state(false)

	const click = async () => {
		if (clicked) return
		clicked = true

		await navigator.clipboard.writeText(props.content ?? "").catch(e => null)
		toast.success('Copied to clipboard')
		
		setTimeout(() => {
			clicked = false
		}, 1500)
	}
</script>

<Icon icon={clicked ? 'check' : 'copy'} class='cursor-pointer {props.class ?? ''}' onclick={click} />