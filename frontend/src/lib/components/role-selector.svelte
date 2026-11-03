<script lang="ts">
    import * as star from "$lib/star";
    import { titleCase } from "@/utils";
    import { onMount } from "svelte";
    import Badge from "./badge.svelte";
    import Checkbox from "./checkbox.svelte";

    let {
        roles = $bindable([]),
    }: {
        roles?: star.Role[];
    } = $props();

    let allRoles = $state<star.Role[]>([]);

    onMount(async () => {
        const available = await star.getAllRoles();
        allRoles = available.filter((role) => !role.default);
    });

    const toggle = (role: star.Role, checked: boolean) => {
        if (checked) roles = [...roles, role];
        else roles = roles.filter((it) => it.id != role.id);
    };

    const isChecked = (role: star.Role) => roles.some((it) => it.id == role.id);
</script>

<div class="flex flex-col gap-1.5 max-h-80 overflow-y-auto overscroll-contain">
    {#each allRoles as role (role.id)}
        <label
            class="flex cursor-pointer items-start gap-3 rounded-lg border border-background-200 p-3 transition-colors hover:bg-hover has-checked:border-primary/40 has-checked:bg-hover"
        >
            <Checkbox
                class="mt-0.5"
                checked={isChecked(role)}
                onchange={(event) => toggle(role, event.currentTarget.checked)}
            />
            <div class="flex min-w-0 flex-col gap-0.5">
                <div class="flex items-center gap-2 text-sm">
                    <span>{titleCase(role.name)}</span>
                    {#if role.internal}
                        <Badge tone="primary">built-in</Badge>
                    {/if}
                </div>
                {#if role.description}
                    <span class="text-sm text-muted">{role.description}</span>
                {/if}
            </div>
        </label>
    {:else}
        <p class="py-6 text-center text-sm text-muted">No roles available.</p>
    {/each}
</div>
