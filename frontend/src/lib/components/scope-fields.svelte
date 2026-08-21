<script lang="ts">
	import { Checkbox } from '@/components/ui/checkbox';
	import { Input } from '@/components/ui/input';
	import { Label } from '@/components/ui/label';

	let {
		id,
		namespace,
		name = $bindable(),
		description = $bindable(),
		requestable = $bindable()
	}: {
		id: string;
		namespace: string;
		name: string;
		description: string;
		requestable: boolean;
	} = $props();
</script>

<div class="flex flex-col gap-6 w-full mx-auto">
	<div class="grid gap-2">
		<Label for="{id}_scope_name">Scope</Label>
		<Input
			id="{id}_scope_name"
			type="text"
			placeholder="{namespace}:example"
			bind:value={name}
			required
		/>
	</div>

	<div class="grid gap-2">
		<Label for="{id}_scope_description">Description</Label>
		<Input
			id="{id}_scope_description"
			type="text"
			placeholder="Allows doing something on your behalf"
			bind:value={description}
			required
		/>
	</div>

	<Label for="{id}_scope_requestable" class="p-3 border rounded-lg">
		<Checkbox id="{id}_scope_requestable" bind:checked={requestable} />
		<div class="flex flex-col items-start gap-1">
			<span>Requestable by other applications</span>
			<span class="text-primary/60 leading-5">
				Lets any user also ask for this scope, when authenticating with other
				applications.
			</span>
		</div>
	</Label>
</div>
