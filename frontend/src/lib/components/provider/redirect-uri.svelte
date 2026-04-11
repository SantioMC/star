<script lang="ts">
	import Copy from '@/components/copy.svelte';
	import Icon from '@/components/icon.svelte';
	import { Button } from '@/components/ui/button';
	import { Spinner } from '@/components/ui/spinner';
	import Modal from './../modal.svelte';

	let {
		redirectUri,
		onremove
	}: {
		redirectUri: string;
		onremove: () => void;
	} = $props();

	let loading = $state<boolean>(false);
	let removeModal = $state<Modal>();

	const removeRedirectUri = async (e: Event) => {
		e.preventDefault();

		removeModal?.close();
		onremove?.();
	};
</script>

<code class="w-auto! flex! items-center gap-2 px-2">
	<Icon icon="world" />
	{redirectUri}
	<Copy content={redirectUri} class="ml-auto" />
	<Icon
		icon="trash"
		class="hover:text-red-300 transition-colors"
		onclick={() => removeModal?.open()}
	/>
</code>

<Modal bind:this={removeModal} button class="xl:w-xl!">
	<h1 class="text-xl font-semibold">Remove Redirect URI</h1>
	<h2 class="text-card-foreground/60 sm:max-w-10/12">
		Remove a redirect URI for this provider, this action can not be reversed.
	</h2>

	<form
		class="flex items-center gap-2 justify-end mt-4"
		onsubmit={removeRedirectUri}
	>
		<Button variant="secondary" onclick={() => removeModal?.close()}>
			Cancel
		</Button>
		<Button type="submit" disabled={loading}>
			{#if loading}
				<Spinner />
			{/if}

			Remove Redirect URI
		</Button>
	</form>
</Modal>
