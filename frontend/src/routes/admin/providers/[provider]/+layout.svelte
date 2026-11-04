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
    import { formatTime, titleCase } from "@/utils";
    import { onMount } from "svelte";
    import { toast } from "svelte-sonner";

    let { children } = $props();

    const canReadAudit = hasPermission("admin:audit:read");

    const resource = new Resource<star.SocialProviderConfig>();
    setResource(resource);

    const provider = $derived(resource.current);

    let deleteModal = $state<ConfirmModal>();

    const tabs = $derived.by(() => {
        if (!provider) return [];

        const params = { provider: provider.id };
        const items: TabItem[] = [
            {
                label: "Overview",
                href: resolve("/admin/providers/[provider]", params),
                icon: "layout-list",
            },
            {
                label: "Endpoints",
                href: resolve("/admin/providers/[provider]/endpoints", params),
                icon: "link",
            },
        ];

        if (canReadAudit) {
            items.push({
                label: "Audit Log",
                href: resolve("/admin/providers/[provider]/audit", params),
                icon: "history",
            });
        }

        return items;
    });

    const menuItems = $derived.by(() => {
        const items: MenuItem[] = [];

        if (hasPermission("admin:provider:delete")) {
            items.push({
                label: "Delete provider",
                icon: "trash",
                danger: true,
                onclick: () => deleteModal?.open(),
            });
        }

        return items;
    });

    onMount(async () => {
        if (!hasPermission("admin:provider:read")) return goto("/");

        const data = await star.getSocialProvider(page.params.provider!);
        if (!data) return goto("/admin/providers");

        resource.current = data;
    });

    const deleteProvider = async () => {
        if (!provider) return;

        const result = await star.deleteSocialProvider(provider.id);
        if ("error" in result) {
            toast.error(result.message);
            return false;
        }

        toast.success(`Deleted the ${titleCase(provider.name)} provider`);
        goto("/admin/providers");
    };
</script>

{#if provider}
    <div class="flex flex-row items-center gap-4">
        <div
            class="flex size-14 shrink-0 items-center justify-center rounded-lg bg-background-200/80"
        >
            <Icon icon={provider.icon || "plug-connected"} class="text-3xl" />
        </div>

        <div class="min-w-0 flex-1">
            <h1 class="leading-8 text-2xl font-semibold truncate">
                {titleCase(provider.name)} Login Provider
            </h1>
            <div
                class="flex flex-wrap items-center gap-x-3 gap-y-1 text-sm text-muted mt-0.5"
            >
                <span class="font-mono text-xs">{provider.id}</span>
                {#if provider.oidc_url}
                    <Badge tone="info">oidc discovery</Badge>
                {/if}
                <span>Created {formatTime(new Date(provider.created_at))}</span>
            </div>
        </div>

        <Menu items={menuItems} label="Provider actions" />
    </div>

    <Tabs items={tabs} label="Provider sections" />

    {@render children()}

    <ConfirmModal
        bind:this={deleteModal}
        title="Delete provider"
        confirmLabel="Delete provider"
        confirmText={provider.name}
        onconfirm={deleteProvider}
    >
        Social login through
        <span class="text-foreground">{titleCase(provider.name)}</span> will
        stop working. Users who only sign in with this provider will lose access
        to their accounts.
    </ConfirmModal>
{/if}
