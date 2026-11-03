<script lang="ts">
    import { branding } from "@/branding";
    import { goto } from "$app/navigation";
    import { page } from "$app/state";
    import * as star from "$lib/star";
    import Badge from "@/components/badge.svelte";
    import Button from "@/components/button.svelte";
    import Card from "@/components/card.svelte";
    import Checkbox from "@/components/checkbox.svelte";
    import ConfirmModal from "@/components/confirm-modal.svelte";
    import Notice from "@/components/notice.svelte";
    import PermissionChip from "@/components/permission-chip.svelte";
    import Input from "@/components/input.svelte";
    import Menu, { type MenuItem } from "@/components/menu.svelte";
    import Modal from "@/components/modal.svelte";
    import SearchInput from "@/components/search-input.svelte";
    import { Label } from "@/components/ui/label";
    import { hasPermission } from "@/index";
    import { groupByNamespace, groupPermissions } from "@/permissions";
    import { formatTime, titleCase } from "@/utils";
    import { onMount } from "svelte";
    import { toast } from "svelte-sonner";

    let role = $state<star.Role | null>(null);
    let available = $state<star.AvailablePermission[]>([]);

    let name = $state("");
    let description = $state("");
    let saving = $state(false);

    let pendingPermission = $state<string | null>(null);

    let addPermissionModal = $state<Modal>();
    let permissionQuery = $state("");
    let selectedPermissions = $state<string[]>([]);
    let addingPermissions = $state(false);

    let deleteRoleModal = $state<ConfirmModal>();

    const canEditDetails = $derived(
        hasPermission("admin:role:write") && role != null && !role.internal,
    );

    const canEditPermissions = $derived(
        hasPermission("admin:role:write") &&
            role != null &&
            (!role.internal || role.default),
    );

    const canDelete = $derived(
        hasPermission("admin:role:delete") && role != null && !role.internal,
    );

    const menuItems = $derived.by(() => {
        const items: MenuItem[] = [];

        if (canDelete) {
            items.push({
                label: "Delete role",
                icon: "trash",
                danger: true,
                onclick: () => deleteRoleModal?.open(),
            });
        }

        return items;
    });

    const descriptions = $derived(
        new Map(available.map((it) => [it.name, it.description])),
    );

    const unassigned = $derived(
        available.filter((it) => !role?.permissions.includes(it.name)),
    );

    const filteredUnassigned = $derived.by(() => {
        const search = permissionQuery.trim().toLowerCase();
        if (!search) return unassigned;

        return unassigned.filter(
            (it) =>
                it.name.toLowerCase().includes(search) ||
                it.description.toLowerCase().includes(search),
        );
    });

    const groupedUnassigned = $derived(
        groupByNamespace(filteredUnassigned, (it) => it.name),
    );

    const groupedAssigned = $derived(groupPermissions(role?.permissions ?? []));

    const customPermission = $derived.by(() => {
        const value = permissionQuery.trim();
        if (!value || value.includes(" ")) return null;
        if (role?.permissions.includes(value)) return null;
        if (available.some((it) => it.name === value)) return null;

        return value;
    });

    const additionCount = $derived(
        selectedPermissions.length + (customPermission ? 1 : 0),
    );

    const resetForm = (data: star.Role) => {
        name = data.name;
        description = data.description;
    };

    onMount(async () => {
        if (!hasPermission("admin:role:read")) return goto("/");

        const data = await star.getRole(page.params.role!);
        if (!data) return goto("/admin/roles");

        role = data;
        resetForm(data);
        available = await star.getAvailablePermissions();
    });

    const updateRole = async (event: Event) => {
        event.preventDefault();
        if (!role || saving) return;

        saving = true;
        const result = await star.updateRole(role.id, { name, description });
        saving = false;

        if ("error" in result) {
            resetForm(role);
            return toast.error(result.message);
        }

        role = result;
        toast.success("Successfully updated role");
    };

    const savePermissions = async (permissions: string[]) => {
        if (!role) return false;

        const result = await star.updateRole(role.id, { permissions });
        if ("error" in result) {
            toast.error(result.message);
            return false;
        }

        role = result;
        return true;
    };

    const removePermission = async (permission: string) => {
        if (!role || pendingPermission) return;

        pendingPermission = permission;
        const remaining = role.permissions.filter((it) => it !== permission);
        const saved = await savePermissions(remaining);
        pendingPermission = null;

        if (saved) toast.success(`Removed ${permission}`);
    };

    const openAddPermissionModal = () => {
        permissionQuery = "";
        selectedPermissions = [];
        addPermissionModal?.open();
    };

    const toggleSelected = (permission: string, checked: boolean) => {
        if (checked) selectedPermissions = [...selectedPermissions, permission];
        else
            selectedPermissions = selectedPermissions.filter(
                (it) => it !== permission,
            );
    };

    const addPermissions = async (event: Event) => {
        event.preventDefault();
        if (!role || addingPermissions) return;

        const additions = [...selectedPermissions];
        if (customPermission) {
            additions.push(customPermission);
        }

        if (additions.length === 0) return;

        addingPermissions = true;
        const saved = await savePermissions([
            ...role.permissions,
            ...additions,
        ]);
        addingPermissions = false;

        if (!saved) return;

        addPermissionModal?.close();
        toast.success(
            `Added ${additions.length} permission${additions.length === 1 ? "" : "s"}`,
        );
    };

    const deleteRole = async () => {
        if (!role) return;

        const result = await star.deleteRole(role.id);
        if ("error" in result) {
            toast.error(result.message);
            return false;
        }

        toast.success(`Deleted role ${role.name}`);
        goto("/admin/roles");
    };
</script>

{#if role}
    <div class="flex flex-row items-center gap-4">
        <div class="min-w-0 flex-1">
            <h1 class="leading-8 text-2xl font-semibold truncate">
                {titleCase(role.name)}
            </h1>
            <div
                class="flex flex-wrap items-center gap-x-3 gap-y-1 text-sm text-muted mt-0.5"
            >
                <span class="font-mono text-xs">{role.id}</span>
                {#if role.internal}
                    <Badge tone="primary">built-in</Badge>
                {/if}
                <span>Created {formatTime(new Date(role.created_at))}</span>
            </div>
        </div>

        <Menu items={menuItems} label="Role actions" />
    </div>

    {#if role.default}
        <Notice tone="info" icon="users" title="Default role.">
            Every user inherits these permissions automatically, without being
            assigned this role. Only its permissions can be changed.
        </Notice>
    {:else if role.internal}
        <Notice icon="lock" title="Built-in role.">
            This role is managed by {$branding.name} and can not be modified or deleted.
        </Notice>
    {/if}

    <Card title="Role Information" icon="package">
        <form onsubmit={updateRole} class="mt-6">
            <div class="flex flex-col gap-5">
                <div class="grid gap-2">
                    <Label for="name">Name</Label>
                    <Input
                        bind:value={name}
                        id="name"
                        type="text"
                        disabled={!canEditDetails}
                        required
                    />
                </div>

                <div class="grid gap-2">
                    <Label for="description">Description</Label>
                    <Input
                        bind:value={description}
                        id="description"
                        type="text"
                        disabled={!canEditDetails}
                    />
                </div>
            </div>

            {#if canEditDetails}
                <div class="mt-4 flex items-center gap-2 justify-end">
                    <Button type="submit" loading={saving}>Update</Button>
                </div>
            {/if}
        </form>
    </Card>

    <Card
        title="Permissions"
        icon="lock"
        description={role.default
            ? "Granted to every user, and matched against the permissions claim in access tokens."
            : "Granted to every user with this role, and matched against the permissions claim in access tokens."}
        class="flex flex-col gap-4"
    >
        {#snippet actions()}
            {#if canEditPermissions}
                <Button variant="secondary" onclick={openAddPermissionModal}>
                    Add permission
                </Button>
            {/if}
        {/snippet}

        <div class="flex flex-col gap-5">
            {#each groupedAssigned as group (group.namespace)}
                <section class="flex flex-col gap-1">
                    <h3
                        class="flex items-center justify-between border-b border-background-200 pb-1.5 font-sans text-xs font-medium text-muted select-none"
                    >
                        {titleCase(group.namespace)}
                        <span class="font-mono font-normal">{group.count}</span>
                    </h3>

                    {#each group.resources as resource (resource.resource)}
                        <div
                            class="flex flex-col gap-1.5 py-1.5 sm:flex-row sm:items-center sm:gap-4"
                        >
                            <span
                                class="font-mono text-sm text-foreground/80 sm:w-40 sm:shrink-0"
                            >
                                {resource.resource}
                            </span>
                            <div class="flex flex-wrap gap-1.5">
                                {#each resource.actions as permission (permission.name)}
                                    <PermissionChip
                                        name={permission.name}
                                        label={permission.action}
                                        description={descriptions.get(
                                            permission.name,
                                        )}
                                        removable={canEditPermissions}
                                        pending={pendingPermission ===
                                            permission.name}
                                        disabled={pendingPermission != null}
                                        onremove={() =>
                                            removePermission(permission.name)}
                                    />
                                {/each}
                            </div>
                        </div>
                    {/each}
                </section>
            {:else}
                <p class="py-6 text-center text-sm text-muted">
                    This role has no permissions.
                </p>
            {/each}
        </div>
    </Card>

    <Modal bind:this={addPermissionModal} button class="sm:w-2xl">
        <form class="flex flex-col gap-6" onsubmit={addPermissions}>
            <div class="flex flex-col gap-1 pr-8">
                <h1 class="text-xl font-semibold">Add permissions</h1>
                <h2 class="text-sm text-muted">
                    Users with the {role.name} role will be granted the selected
                    permissions.
                </h2>
            </div>

            <div class="flex flex-col gap-3">
                <SearchInput
                    bind:value={permissionQuery}
                    placeholder="Search or enter a custom permission"
                    class="sm:max-w-none"
                />

                <div
                    class="max-h-[min(28rem,55vh)] overflow-y-auto overscroll-contain rounded-lg border border-background-200"
                >
                    {#each groupedUnassigned as [namespace, permissions] (namespace)}
                        <div
                            class="sticky top-0 z-10 flex items-center justify-between border-b border-background-200 bg-background-100 px-4 py-2 text-xs text-muted select-none"
                        >
                            <span class="font-medium">{titleCase(namespace)}</span>
                            <span class="font-mono">{permissions.length}</span>
                        </div>

                        {#each permissions as permission (permission.id)}
                            <label
                                class="flex cursor-pointer items-center gap-3 border-b border-background-200 px-4 py-3 transition-colors last:border-b-0 hover:bg-hover has-checked:bg-hover"
                            >
                                <Checkbox
                                    checked={selectedPermissions.includes(
                                        permission.name,
                                    )}
                                    onchange={(event) =>
                                        toggleSelected(
                                            permission.name,
                                            event.currentTarget.checked,
                                        )}
                                />
                                <div class="min-w-0 flex-1">
                                    <span class="font-mono text-sm">
                                        {permission.name}
                                    </span>
                                    {#if permission.description}
                                        <p class="text-sm text-muted">
                                            {permission.description}
                                        </p>
                                    {/if}
                                </div>
                            </label>
                        {/each}
                    {:else}
                        <p class="px-4 py-8 text-center text-sm text-muted">
                            {#if customPermission}
                                No known permission matches, submit to add
                                <span class="font-mono text-foreground"
                                    >{customPermission}</span
                                > as a custom permission.
                            {:else}
                                No permissions left to add.
                            {/if}
                        </p>
                    {/each}
                </div>
            </div>

            <div class="flex items-center gap-2 justify-end">
                <span class="mr-auto text-sm text-muted select-none">
                    {selectedPermissions.length} selected
                </span>
                <Button
                    type="button"
                    variant="secondary"
                    onclick={() => addPermissionModal?.close()}
                >
                    Cancel
                </Button>
                <Button
                    type="submit"
                    loading={addingPermissions}
                    disabled={additionCount === 0}
                >
                    {additionCount > 1 ? `Add ${additionCount}` : "Add"}
                </Button>
            </div>
        </form>
    </Modal>

    <ConfirmModal
        bind:this={deleteRoleModal}
        title="Delete role"
        confirmLabel="Delete role"
        confirmText={role.name}
        onconfirm={deleteRole}
    >
        <span class="text-foreground">{role.name}</span> will be removed from every
        user who currently has it assigned. This action can not be reversed.
    </ConfirmModal>
{/if}
