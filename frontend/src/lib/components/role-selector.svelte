<script lang="ts">
	import * as star from '$lib/star';
	import { Badge } from '@/components/ui/badge';
	import { Checkbox } from '@/components/ui/checkbox';
	import Label from '@/components/ui/label/label.svelte';
	import { titleCase } from '@/utils';
	import { onMount } from 'svelte';

	let {
		roles = $bindable([])
	}: {
		roles?: star.Role[]
	} = $props()

	let allRoles = $state<star.Role[]>([])

	onMount(async () => {
		allRoles = await star.getAllRoles()
	})

	const handleClick = (value: boolean, role: star.Role) => {
		if (value) roles.push(role)
		else roles = roles.filter(r => r.id != role.id)
	}

	const isChecked = (role: star.Role) => {
		return roles.some(r => r.id == role.id)
	}
</script>

<div class="flex flex-col gap-1">
	{#each allRoles as role (role.id)}
		<Label for={role.id} class='p-3 border rounded-lg'>
			<Checkbox id={role.id} checked={isChecked(role)} onCheckedChange={(val) => handleClick(val, role)} />
			<div class='flex flex-col items-start gap-1'>
				<div class='flex items-center gap-1'>
					<span>{titleCase(role.name)}</span>
					{#if role.internal}
						<Badge class='ml-1' variant='secondary'>Built-in role</Badge>
					{/if}
				</div>
				<span class='text-primary/60 leading-5'>{role.description}</span>
			</div>
		</Label>
	{/each}
</div>
