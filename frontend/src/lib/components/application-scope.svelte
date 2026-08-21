<script lang="ts">
	import * as star from '$lib/star';
	import Icon from '@/components/icon.svelte';
	import ScopeFields from '@/components/scope-fields.svelte';
	import ScopeItem from '@/components/scope-item.svelte';
	import { Badge } from '@/components/ui/badge';
	import { Button } from '@/components/ui/button';
	import { Spinner } from '@/components/ui/spinner';
	import { relativeTime } from '@/utils';
	import { toast } from 'svelte-sonner';
	import Modal from './modal.svelte';

	const {
		application,
		scope,
		namespace,
		canModify = true,
		onchange
	}: {
		application: string;
		scope: star.ApplicationScope;
		namespace: string;
		canModify?: boolean;
		onchange?: () => void;
	} = $props();

	let loading = $state<boolean>(false);
	let editModal = $state<Modal>();
	let deleteModal = $state<Modal>();

	let name = $state<string>('');
	let description = $state<string>('');
	let requestable = $state<boolean>(false);

	const openEdit = () => {
		name = scope.name;
		description = scope.description;
		requestable = scope.requestable;
		editModal?.open();
	};

	const edit = async (e: Event) => {
		e.preventDefault();

		loading = true;
		const result = await star.updateApplicationScope(application, scope.id, {
			name: name.trim(),
			description: description.trim(),
			requestable
		});
		loading = false;

		if ('error' in result) {
			toast.error(result.message);
			return;
		}

		editModal?.close();
		onchange?.();
		toast.success('Successfully updated scope');
	};

	const remove = async (e: Event) => {
		e.preventDefault();

		loading = true;
		const result = await star.deleteApplicationScope(application, scope.id);
		loading = false;

		if ('error' in result) {
			toast.error(result.message);
			return;
		}

		deleteModal?.close();
		onchange?.();
		toast.success('Removed scope');
	};
</script>

<ScopeItem
	name={scope.name}
	description="{scope.description} · Added {relativeTime(scope.created_at)} ago"
>
	{#snippet badges()}
		{#if scope.requestable}
			<Badge variant="outline" class="border-amber-500/60 text-amber-400">
				Requestable
			</Badge>
		{/if}
	{/snippet}

	{#snippet actions()}
		{#if canModify}
			<Button variant="ghost" size="icon" title="Edit scope" onclick={openEdit}>
				<Icon icon="pencil" />
			</Button>

			<Button
				variant="ghost"
				size="icon"
				title="Remove scope"
				onclick={() => deleteModal?.open()}
			>
				<Icon icon="trash" class="text-red-400" />
			</Button>
		{/if}
	{/snippet}
</ScopeItem>

<Modal bind:this={editModal} button class="xl:w-xl!">
	<h1 class="text-xl font-semibold">Edit Scope</h1>
	<h2 class="text-card-foreground/60 sm:max-w-10/12">
		Scopes must stay inside this application's
		<code>{namespace}:</code>
		namespace. The description is shown to users when they authorize the application.
	</h2>

	<form onsubmit={edit} class="mt-6">
		<ScopeFields
			id="edit"
			{namespace}
			bind:name
			bind:description
			bind:requestable
		/>

		<div class="flex items-center gap-2 justify-end mt-6">
			<Button type="button" variant="secondary" onclick={() => editModal?.close()}>
				Cancel
			</Button>
			<Button type="submit" disabled={loading}>
				{#if loading}
					<Spinner />
				{/if}
				Save Scope
			</Button>
		</div>
	</form>
</Modal>

<Modal bind:this={deleteModal} button class="xl:w-xl!">
	<h1 class="text-xl font-semibold">Remove Scope</h1>
	<h2 class="text-card-foreground/60 sm:max-w-10/12">
		Remove <code>{scope.name}</code> from this application. Tokens that were issued with this scope
		lose it the next time they are refreshed.
	</h2>

	<form class="flex items-center gap-2 justify-end mt-4" onsubmit={remove}>
		<Button type="button" variant="secondary" onclick={() => deleteModal?.close()}>
			Cancel
		</Button>
		<Button type="submit" variant="destructive" disabled={loading}>
			{#if loading}
				<Spinner />
			{/if}

			Remove Scope
		</Button>
	</form>
</Modal>
