<script lang="ts">
    import { branding } from "@/branding";
    import { goto } from "$app/navigation";
    import { resolve } from "$app/paths";
    import {
        CLIENT_KIND_DESCRIPTIONS,
        CLIENT_KIND_OPTIONS,
        type ClientKind,
    } from "@/applications";
    import Badge from "@/components/badge.svelte";
    import Button from "@/components/button.svelte";
    import Card from "@/components/card.svelte";
    import ClientSecretModal from "@/components/client-secret-modal.svelte";
    import Icon from "@/components/icon.svelte";
    import Input from "@/components/input.svelte";
    import Modal from "@/components/modal.svelte";
    import PageHeader from "@/components/page-header.svelte";
    import SearchInput from "@/components/search-input.svelte";
    import Select from "@/components/select.svelte";
    import { Label } from "@/components/ui/label";
    import { hasPermission } from "@/index";
    import * as star from "@/star";
    import { relativeTime } from "@/utils";
    import { onMount } from "svelte";
    import { toast } from "svelte-sonner";

    const canCreate = hasPermission("application:create");

    let apps = $state<star.AdminApplication[]>([]);
    let loaded = $state(false);
    let query = $state("");

    let createModal = $state<Modal>();
    let secretModal = $state<ClientSecretModal>();
    let creating = $state(false);
    let name = $state("");
    let kind = $state<ClientKind>("public");
    let url = $state("");

    const filteredApps = $derived.by(() => {
        const needle = query.trim().toLowerCase();
        if (!needle) return apps;

        return apps.filter((app) =>
            [app.name, app.client_id, app.application_url].some((value) =>
                value.toLowerCase().includes(needle),
            ),
        );
    });

    const load = async () => {
        apps = await star.getAllApps();
        loaded = true;
    };

    onMount(() => {
        if (!hasPermission("application:read")) return goto("/");
        load();
    });

    const openCreate = () => {
        name = "";
        kind = "public";
        url = "";
        createModal?.open();
    };

    const createApp = async (event: Event) => {
        event.preventDefault();
        if (creating) return;

        creating = true;
        const result = await star.createApplication({
            name: name.trim(),
            kind,
            url: url.trim(),
        });
        creating = false;

        if ("error" in result) {
            return toast.error(result.message);
        }

        await load();
        createModal?.close();
        if (result.client_secret) secretModal?.show(result.client_secret);

        toast.success("Successfully created a new application");
    };

    const appHref = (app: star.AdminApplication) =>
        resolve("/admin/apps/[app]", { app: app.id });
</script>

<PageHeader
    title="Applications"
    description="Clients allowed to authenticate users registered with {$branding.name}"
>
    {#snippet actions()}
        {#if canCreate}
            <Button onclick={openCreate}>
                <Icon icon="plus" />
                New application
            </Button>
        {/if}
    {/snippet}
</PageHeader>

<SearchInput
    bind:value={query}
    placeholder="Search by name, client ID or URL"
/>

<Card class="p-2 sm:p-3">
    <table class="w-full text-sm">
        <thead>
            <tr class="hover:bg-transparent">
                <th class="px-3 text-xs font-medium text-muted select-none">
                    Application
                </th>
                <th
                    class="px-3 text-xs font-medium text-muted select-none max-sm:hidden"
                >
                    Type
                </th>
                <th
                    class="px-3 text-xs font-medium text-muted select-none text-right max-sm:hidden"
                >
                    Created
                </th>
            </tr>
        </thead>
        <tbody>
            {#each filteredApps as app (app.id)}
                <tr
                    class="cursor-pointer hover:bg-hover"
                    onclick={() => goto(appHref(app))}
                >
                    <td class="px-3 py-2.5">
                        <div class="flex items-center gap-3">
                            <Icon
                                icon="package"
                                class="shrink-0 text-lg text-muted"
                            />
                            <div class="min-w-0">
                                <div class="flex items-center gap-2">
                                    <a
                                        href={appHref(app)}
                                        class="truncate font-medium"
                                        onclick={(event) =>
                                            event.stopPropagation()}
                                    >
                                        {app.name}
                                    </a>
                                    {#if app.internal}
                                        <Badge tone="primary">built-in</Badge>
                                    {/if}
                                    {#if app.development_mode}
                                        <Badge tone="warning">development</Badge>
                                    {/if}
                                </div>
                                <p class="truncate font-mono text-xs text-muted">
                                    {app.application_url}
                                </p>
                            </div>
                        </div>
                    </td>
                    <td class="px-3 max-sm:hidden">
                        <Badge tone={app.kind === "confidential" ? "info" : "neutral"}>
                            {app.kind}
                        </Badge>
                    </td>
                    <td
                        class="px-3 text-right text-muted whitespace-nowrap max-sm:hidden"
                    >
                        {relativeTime(app.created_at)} ago
                    </td>
                </tr>
            {:else}
                <tr class="hover:bg-transparent">
                    <td colspan="3" class="py-10 text-center text-muted">
                        {#if !loaded}
                            Loading applications...
                        {:else if query}
                            No applications match "{query}".
                        {:else}
                            No applications found.
                        {/if}
                    </td>
                </tr>
            {/each}
        </tbody>
    </table>
</Card>

{#if canCreate}
    <Modal bind:this={createModal} button>
        <form class="flex flex-col gap-6" onsubmit={createApp}>
            <div class="flex flex-col gap-1 pr-8">
                <h1 class="text-xl font-semibold">Create application</h1>
                <h2 class="text-sm text-muted">
                    Register a new client that can authenticate users with {$branding.name}.
                </h2>
            </div>

            <div class="flex flex-col gap-5">
                <div class="grid gap-2">
                    <Label for="name">Name</Label>
                    <Input
                        bind:value={name}
                        id="name"
                        type="text"
                        placeholder="Acme Dashboard"
                        required
                    />
                </div>

                <div class="grid gap-2">
                    <Label for="kind">Client type</Label>
                    <Select
                        bind:value={kind}
                        id="kind"
                        options={CLIENT_KIND_OPTIONS}
                    />
                    <p class="text-xs text-muted">
                        {CLIENT_KIND_DESCRIPTIONS[kind]}
                    </p>
                </div>

                <div class="grid gap-2">
                    <Label for="url">Application URL</Label>
                    <Input
                        bind:value={url}
                        id="url"
                        type="url"
                        class="font-mono"
                        placeholder="https://example.com"
                        required
                    />
                </div>
            </div>

            <div class="flex items-center gap-2 justify-end">
                <Button
                    type="button"
                    variant="secondary"
                    onclick={() => createModal?.close()}
                >
                    Cancel
                </Button>
                <Button type="submit" loading={creating}>
                    Create application
                </Button>
            </div>
        </form>
    </Modal>

    <ClientSecretModal bind:this={secretModal} />
{/if}
