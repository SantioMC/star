<script lang="ts">
	import * as star from '$lib/star';
	import Icon from '@/components/icon.svelte';
	import { Badge } from '@/components/ui/badge';
	import { Button } from '@/components/ui/button';
	import * as Item from '@/components/ui/item';
	import { Spinner } from '@/components/ui/spinner';
	import { relativeTime } from '@/utils';
	import { toast } from 'svelte-sonner';
	import Modal from './modal.svelte';

	const {
		email,
		canDelete = true,
		onprimary,
		ondelete
	}: {
		email: star.Email;
		canDelete?: boolean;
		onprimary?: () => void;
		ondelete?: () => void;
	} = $props();

	let primaryLoading = $state<boolean>(false);
	let deleteLoading = $state<boolean>(false);
	let deleteModal = $state<Modal>();

	const makePrimary = async () => {
		primaryLoading = true;
		const result = await star.setPrimaryEmail(email.id);
		primaryLoading = false;

		if ('error' in result) {
			toast.error(result.message);
		} else {
			onprimary?.();
		}
	};

	const remove = async (e: Event) => {
		e.preventDefault();
		deleteLoading = true;
		const result = await star.deleteEmail(email.id);

		if ('error' in result) {
			deleteLoading = false;
			toast.error(result.message);
		} else {
			deleteModal?.close();
			ondelete?.();
		}
	};
</script>

<Item.Root variant="outline">
	<Item.Content class="flex flex-row items-center gap-4">
		<Icon class="text-2xl" icon="mail" />
		<div class="min-w-0">
			<p class="text-base flex items-center gap-2 flex-wrap">
				<span class="truncate">{email.email}</span>

				{#if email.primary}
					<Badge variant="outline" class="border-blue-500/60 text-blue-400">Primary</Badge>
				{/if}

				{#if email.verified}
					<Badge variant="outline" class="border-emerald-600/60 text-emerald-500">Verified</Badge>
				{:else}
					<Badge variant="outline" class="border-amber-500/60 text-amber-400">Unverified</Badge>
				{/if}
			</p>

			<Item.Description class="text-xs">Added {relativeTime(email.created_at)} ago</Item.Description>
		</div>

		<div class="ml-auto flex items-center gap-1.5">
			{#if !email.primary && email.verified}
				<Button variant="secondary" size="sm" disabled={primaryLoading} onclick={makePrimary}>
					{#if primaryLoading}
						<Spinner />
					{/if}

					Make primary
				</Button>
			{/if}

			{#if canDelete}
				<Button
					variant="ghost"
					size="icon"
					title="Remove email"
					onclick={() => deleteModal?.open()}
				>
					<Icon icon="trash" class="text-red-400" />
				</Button>
			{/if}
		</div>
	</Item.Content>
</Item.Root>

<Modal bind:this={deleteModal} button class="xl:w-xl!">
	<h1 class="text-xl font-semibold">Remove Email Address</h1>
	<h2 class="text-card-foreground/60 sm:max-w-10/12">
		Remove <span class="text-primary">{email.email}</span> from your account. This action can not be
		reversed.
	</h2>

	<form class="flex items-center gap-2 justify-end mt-4" onsubmit={remove}>
		<Button type="button" variant="secondary" onclick={() => deleteModal?.close()}>Cancel</Button>
		<Button type="submit" variant="destructive" disabled={deleteLoading}>
			{#if deleteLoading}
				<Spinner />
			{/if}

			Remove Email
		</Button>
	</form>
</Modal>
