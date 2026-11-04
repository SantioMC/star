<script lang="ts">
    import { goto } from "$app/navigation";
    import { resolve } from "$app/paths";
    import { page } from "$app/state";
    import * as star from "$lib/star";
    import Badge from "@/components/badge.svelte";
    import ConfirmModal from "@/components/confirm-modal.svelte";
    import Icon from "@/components/icon.svelte";
    import Menu, { type MenuItem } from "@/components/menu.svelte";
    import Tabs, { type TabItem } from "@/components/tabs.svelte";
    import { hasPermission } from "@/index";
    import { Resource, setResource } from "@/resource.svelte";
    import { formatTime } from "@/utils";
    import { onMount } from "svelte";
    import { toast } from "svelte-sonner";

    let { children } = $props();

    const canReadAudit = hasPermission("admin:audit:read");

    const resource = new Resource<star.AdminApplication>();
    setResource(resource);

    const app = $derived(resource.current);

    let deleteAppModal = $state<ConfirmModal>();

    const tabs = $derived.by(() => {
        if (!app) return [];

        const params = { app: app.id };
        const items: TabItem[] = [
            {
                label: "Overview",
                href: resolve("/admin/apps/[app]", params),
                icon: "layout-list",
            },
            {
                label: "Endpoints",
                href: resolve("/admin/apps/[app]/endpoints", params),
                icon: "link",
            },
        ];

        if (canReadAudit) {
            items.push({
                label: "Audit Log",
                href: resolve("/admin/apps/[app]/audit", params),
                icon: "history",
            });
        }

        return items;
    });

    const menuItems = $derived.by(() => {
        const items: MenuItem[] = [];

        if (app?.internal === false && hasPermission("application:delete")) {
            items.push({
                label: "Delete application",
                icon: "trash",
                danger: true,
                onclick: () => deleteAppModal?.open(),
            });
        }

        return items;
    });

    onMount(async () => {
        if (!hasPermission("application:read")) return goto("/");

        const data = await star.getApplication(page.params.app!);
        if (!data) return goto("/admin/apps");

        resource.current = data;
    });

    const deleteApp = async () => {
        if (!app) return;

        const result = await star.deleteApplication(app.id);
        if ("error" in result) {
            toast.error(result.message);
            return false;
        }

        toast.success(`Deleted ${app.name}`);
        goto("/admin/apps");
    };
</script>

{#if app}
    <div class="flex flex-row items-center gap-4">
        <div
            class="flex size-14 shrink-0 items-center justify-center rounded-full bg-background-200"
        >
            <Icon icon="package" class="text-2xl text-muted" />
        </div>

        <div class="min-w-0 flex-1">
            <h1 class="leading-8 text-2xl font-semibold truncate">
                {app.name}
            </h1>
            <div
                class="flex flex-wrap items-center gap-x-3 gap-y-1 text-sm text-muted mt-0.5"
            >
                <span class="font-mono text-xs">{app.id}</span>
                <Badge tone={app.kind === "confidential" ? "info" : "neutral"}>
                    {app.kind}
                </Badge>
                {#if app.internal}
                    <Badge tone="primary">built-in</Badge>
                {/if}
                {#if app.development_mode}
                    <Badge tone="warning">development</Badge>
                {/if}
                <span>Created {formatTime(new Date(app.created_at))}</span>
            </div>
        </div>

        <Menu items={menuItems} label="Application actions" />
    </div>

    <Tabs items={tabs} label="Application sections" />

    {@render children()}

    <ConfirmModal
        bind:this={deleteAppModal}
        title="Delete application"
        confirmLabel="Delete application"
        confirmText={app.name}
        onconfirm={deleteApp}
    >
        Deleting <span class="text-foreground">{app.name}</span> invalidates all of
        its sessions and breaks every login that uses it. This action can not be reversed.
    </ConfirmModal>
{/if}
