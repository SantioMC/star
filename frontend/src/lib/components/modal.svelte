<script lang="ts">
	import { Button } from '@/components/ui/button';
	import type { ClassValue } from 'svelte/elements';
	import Icon from './icon.svelte';

	let {
		showModal = $bindable(false),
		closable = true,
		button = false,
		class: className,
		children
	}: {
		showModal?: boolean;
		children?: any;
		closable?: boolean;
		button?: boolean;
		class?: ClassValue;
	} = $props();

	let dialog: HTMLDialogElement = $state()!;

	$effect(() => {
		if (showModal) dialog.showModal();
	});

	export const close = () => {
		dialog.close();
		showModal = false;
	};

	export const open = () => {
		dialog.showModal();
		showModal = true;
	};

	const handleClick = (e: MouseEvent) => {
		if (e.target != dialog || !closable) return;
		dialog.close();
	};
</script>

<!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_noninteractive_element_interactions -->
<dialog
	bind:this={dialog}
	onclose={() => (showModal = false)}
	onclick={handleClick}
	class="relative top-1/2 left-1/2 w-full -translate-1/2 rounded-lg border-none bg-card p-0 text-card-foreground backdrop:bg-black/50 sm:max-w-1/3 sm:min-w-96 md:overflow-hidden {className}"
>
	<div class="rounded-lg bg-card p-4 text-card-foreground">
		{#if button}
			<Button
				type="button"
				variant='ghost'
				class="absolute top-0 right-0 h-10 w-10 rounded p-[0.5px]"
				onclick={close}
			>
				<Icon icon="x" class='text-lg' />
			</Button>
		{/if}

		{@render children?.()}
	</div>
</dialog>

<style>
	@keyframes zoom {
		from {
			transform: scale(0.95);
		}
		to {
			transform: scale(1);
		}
	}

	@keyframes fade {
		from {
			opacity: 0;
		}
		to {
			opacity: 1;
		}
	}

	@media screen and (width >= 48rem) {
		dialog {
			width: var(--width) !important;
		}
	}

	dialog[open] {
		animation: zoom 0.3s cubic-bezier(0.34, 1.56, 0.64, 1);
	}

	dialog[open]::backdrop {
		animation: fade 0.2s ease-out;
	}
</style>
