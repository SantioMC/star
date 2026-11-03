<script lang="ts">
    import { branding } from "@/branding";
    import { goto } from "$app/navigation";
    import { resolve } from "$app/paths";
    import * as star from "$lib/star";
    import Badge from "@/components/badge.svelte";
    import Button from "@/components/button.svelte";
    import Card from "@/components/card.svelte";
    import Icon from "@/components/icon.svelte";
    import Input from "@/components/input.svelte";
    import Modal from "@/components/modal.svelte";
    import PageHeader from "@/components/page-header.svelte";
    import SearchInput from "@/components/search-input.svelte";
    import { Label } from "@/components/ui/label";
    import { hasPermission } from "@/index";
    import { simpleDate, titleCase } from "@/utils";
    import { onMount } from "svelte";
    import { toast } from "svelte-sonner";

    const canCreate = hasPermission("admin:role:create");

    let roles = $state<star.Role[]>([]);
    let loaded = $state(false);
    let query = $state("");

    let createRoleModal = $state<Modal>();
    let name = $state("");
    let description = $state("");
    let creating = $state(false);

    const filteredRoles = $derived.by(() => {
        const search = query.trim().toLowerCase();
        if (!search) return roles;

        return roles.filter(
            (role) =>
                role.name.toLowerCase().includes(search) ||
                role.description.toLowerCase().includes(search),
        );
    });

    onMount(async () => {
        if (!hasPermission("admin:role:read")) return goto("/");

        roles = await star.getAllRoles();
        loaded = true;
    });

    const openCreateModal = () => {
        name = "";
        description = "";
        createRoleModal?.open();
    };

    const createRole = async (event: Event) => {
        event.preventDefault();
        if (creating) return;

        creating = true;
        const result = await star.createRole({
            name: name.trim(),
            description: description.trim() || undefined,
        });
        creating = false;

        if ("error" in result) {
            return toast.error(result.message);
        }

        createRoleModal?.close();
        toast.success(`Created role ${result.name}`);
        goto(resolve("/admin/roles/[role]", { role: result.id }));
    };

    const permissionCount = (role: star.Role): string => {
        const count = role.permissions.length;
        return `${count} permission${count === 1 ? "" : "s"}`;
    };
</script>

<PageHeader
    title="Roles & Permissions"
    description="{$branding.name} uses role-based access control, grant permissions to users through roles"
>
    {#snippet actions()}
        {#if canCreate}
            <Button onclick={openCreateModal}>
                <Icon icon="plus" />
                Create role
            </Button>
        {/if}
    {/snippet}
</PageHeader>

<SearchInput bind:value={query} placeholder="Search roles" />

<Card class="p-2 sm:p-3">
    <table class="w-full text-sm">
        <thead>
            <tr class="hover:bg-transparent">
                <th class="px-3 text-xs font-medium text-muted select-none">
                    Role
                </th>
                <th
                    class="px-3 text-xs font-medium text-muted select-none max-sm:hidden"
                >
                    Permissions
                </th>
                <th
                    class="px-3 text-xs font-medium text-muted select-none text-right max-sm:hidden"
                >
                    Created
                </th>
            </tr>
        </thead>
        <tbody>
            {#each filteredRoles as role (role.id)}
                <tr
                    class="cursor-pointer hover:bg-hover"
                    onclick={() =>
                        goto(resolve("/admin/roles/[role]", { role: role.id }))}
                >
                    <td class="px-3 py-2.5">
                        <div class="flex items-center gap-2">
                            <a
                                href={resolve("/admin/roles/[role]", {
                                    role: role.id,
                                })}
                                class="font-medium"
                                onclick={(event) => event.stopPropagation()}
                            >
                                {titleCase(role.name)}
                            </a>
                            {#if role.internal}
                                <Badge tone="primary">built-in</Badge>
                            {/if}
                        </div>
                        {#if role.description}
                            <p class="text-xs text-muted mt-0.5">
                                {role.description}
                            </p>
                        {/if}
                    </td>
                    <td class="px-3 text-muted whitespace-nowrap max-sm:hidden">
                        {permissionCount(role)}
                    </td>
                    <td
                        class="px-3 text-right text-muted whitespace-nowrap max-sm:hidden"
                    >
                        {simpleDate(new Date(role.created_at))}
                    </td>
                </tr>
            {:else}
                <tr class="hover:bg-transparent">
                    <td colspan="3" class="py-10 text-center text-muted">
                        {#if !loaded}
                            Loading roles...
                        {:else if query}
                            No roles match "{query}".
                        {:else}
                            No roles found.
                        {/if}
                    </td>
                </tr>
            {/each}
        </tbody>
    </table>
</Card>

<Modal bind:this={createRoleModal} button>
    <form class="flex flex-col gap-6" onsubmit={createRole}>
        <div class="flex flex-col gap-1 pr-8">
            <h1 class="text-xl font-semibold">Create a role</h1>
            <h2 class="text-sm text-muted">
                Roles group permissions together so they can be granted to
                users.
            </h2>
        </div>

        <div class="flex flex-col gap-5">
            <div class="grid gap-2">
                <Label for="roleName">Name</Label>
                <Input
                    bind:value={name}
                    id="roleName"
                    type="text"
                    placeholder="moderator"
                    autocomplete="off"
                    required
                />
            </div>

            <div class="grid gap-2">
                <Label for="roleDescription">Description</Label>
                <Input
                    bind:value={description}
                    id="roleDescription"
                    type="text"
                    placeholder="Moderators with limited admin access"
                    autocomplete="off"
                />
            </div>
        </div>

        <div class="flex items-center gap-2 justify-end">
            <Button
                type="button"
                variant="secondary"
                onclick={() => createRoleModal?.close()}
            >
                Cancel
            </Button>
            <Button type="submit" loading={creating} disabled={!name.trim()}>
                Create role
            </Button>
        </div>
    </form>
</Modal>
