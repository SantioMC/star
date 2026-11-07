<script lang="ts">
    import { tooltip } from "@/tooltip";
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
    import { relativeTime, titleCase } from "@/utils";
    import { onMount } from "svelte";
    import { toast } from "svelte-sonner";

    const CUSTOM_PRESET = "__custom";

    const canCreate = hasPermission("admin:provider:create");

    let providers = $state<star.SocialProviderConfig[]>([]);
    let presets = $state<star.SocialProviderPreset[]>([]);
    let loaded = $state(false);
    let query = $state("");

    let createModal = $state<Modal>();
    let creating = $state(false);
    let selectedPreset = $state<string | null>(null);
    let customName = $state("");
    let discoveryUri = $state("");
    let clientId = $state("");
    let clientSecret = $state("");

    const isCustom = $derived(selectedPreset === CUSTOM_PRESET);

    const availablePresets = $derived(
        presets.filter(
            (preset) =>
                !providers.some((existing) => existing.name === preset.name),
        ),
    );

    const filteredProviders = $derived.by(() => {
        const search = query.trim().toLowerCase();
        const sorted = [...providers].sort((a, b) => a.order - b.order);
        if (!search) return sorted;

        return sorted.filter(
            (provider) =>
                provider.name.toLowerCase().includes(search) ||
                provider.client_id.toLowerCase().includes(search),
        );
    });

    onMount(async () => {
        if (!hasPermission("admin:provider:read")) return goto("/");

        const presetRequest = star.getSocialProviderPresets();

        providers = await star.getSocialProviders();
        loaded = true;

        presets = await presetRequest;
    });

    const openCreate = () => {
        selectedPreset = null;
        customName = "";
        discoveryUri = "";
        clientId = "";
        clientSecret = "";
        createModal?.open();
    };

    const createProvider = async (event: Event) => {
        event.preventDefault();
        if (!selectedPreset || creating) return;

        const preset = presets.find((it) => it.name === selectedPreset);
        const name = isCustom ? customName.trim() : selectedPreset;

        creating = true;
        const result = await star.createSocialProvider({
            name,
            icon: preset?.icon,
            client_id: clientId,
            client_secret: clientSecret,
            scopes: preset?.scopes,
            oidc_url: isCustom ? discoveryUri : preset?.oidc_url,
            authorization_url: preset?.authorization_url,
            token_url: preset?.token_url,
            userinfo_url: preset?.userinfo_url,
            trust_email: preset?.trust_email ?? false,
            order: providers.length,
        });
        creating = false;

        if ("error" in result) {
            return toast.error(result.message);
        }

        providers = [...providers, result];
        createModal?.close();
        toast.success(`Added the ${titleCase(name)} provider`);
    };

    const providerHref = (provider: star.SocialProviderConfig) =>
        resolve("/admin/providers/[provider]", { provider: provider.id });
</script>

<PageHeader
    title="Social Login Providers"
    description="Configure external identity providers users can sign in with"
>
    {#snippet actions()}
        {#if canCreate}
            <Button onclick={openCreate}>
                <Icon icon="plus" />
                Add provider
            </Button>
        {/if}
    {/snippet}
</PageHeader>

<SearchInput bind:value={query} placeholder="Search by name or client ID" />

<Card class="p-2 sm:p-3">
    <table class="w-full text-sm">
        <thead>
            <tr class="hover:bg-transparent">
                <th class="px-3 text-xs font-medium text-muted select-none">
                    Provider
                </th>
                <th
                    class="px-3 text-xs font-medium text-muted select-none max-sm:hidden"
                >
                    Client ID
                </th>
                <th
                    class="px-3 text-xs font-medium text-muted select-none text-right max-sm:hidden"
                >
                    Created
                </th>
            </tr>
        </thead>
        <tbody>
            {#each filteredProviders as provider (provider.id)}
                <tr
                    class="cursor-pointer hover:bg-hover"
                    onclick={() => goto(providerHref(provider))}
                >
                    <td class="px-3 py-2.5">
                        <div class="flex items-center gap-3">
                            <div
                                class="flex size-8 shrink-0 items-center justify-center rounded-md bg-background-200/80"
                            >
                                <Icon
                                    icon={provider.icon || "plug-connected"}
                                    class="text-lg"
                                />
                            </div>
                            <div class="min-w-0">
                                <div class="flex items-center gap-2">
                                    <a
                                        href={providerHref(provider)}
                                        class="truncate font-medium"
                                        onclick={(event) =>
                                            event.stopPropagation()}
                                    >
                                        {titleCase(provider.name)}
                                    </a>
                                    {#if provider.oidc_url}
                                        <Badge tone="info">oidc</Badge>
                                    {/if}
                                    {#if !provider.enabled}
                                        <Badge tone="warning">disabled</Badge>
                                    {/if}
                                </div>
                                <p class="truncate font-mono text-xs text-muted">
                                    {provider.id}
                                </p>
                            </div>
                        </div>
                    </td>
                    <td class="px-3 max-sm:hidden">
                        <span
                            class="block max-w-48 truncate font-mono text-xs text-foreground/80"
                            {@attach tooltip(provider.client_id)}
                        >
                            {provider.client_id}
                        </span>
                    </td>
                    <td
                        class="px-3 text-right text-muted whitespace-nowrap max-sm:hidden"
                    >
                        {relativeTime(provider.created_at)} ago
                    </td>
                </tr>
            {:else}
                <tr class="hover:bg-transparent">
                    <td colspan="3" class="py-10 text-center text-muted">
                        {#if !loaded}
                            Loading providers...
                        {:else if query}
                            No providers match "{query}".
                        {:else}
                            No social providers configured yet.
                        {/if}
                    </td>
                </tr>
            {/each}
        </tbody>
    </table>
</Card>

<Modal bind:this={createModal} button>
    <div class="flex flex-col gap-6">
        <div class="flex flex-col gap-1 pr-8">
            <h1 class="text-xl font-semibold">Add social provider</h1>
            <h2 class="text-sm text-muted">
                {#if !selectedPreset}
                    Choose a provider to enable for social login.
                {:else if isCustom}
                    Connect any OpenID Connect provider using its discovery
                    URL.
                {:else}
                    Enter the OAuth credentials issued by {titleCase(
                        selectedPreset,
                    )}.
                {/if}
            </h2>
        </div>

        {#if !selectedPreset}
            <div class="grid grid-cols-1 gap-2 sm:grid-cols-2">
                {#each availablePresets as preset (preset.name)}
                    <button
                        type="button"
                        class="flex items-center gap-3 rounded-lg border border-background-200 px-4 py-3 text-left transition-colors hover:bg-hover"
                        onclick={() => (selectedPreset = preset.name)}
                    >
                        <Icon icon={preset.icon} class="text-xl" />
                        <span>{titleCase(preset.name)}</span>
                    </button>
                {/each}
                <button
                    type="button"
                    class="flex items-center gap-3 rounded-lg border border-background-200 px-4 py-3 text-left transition-colors hover:bg-hover sm:col-span-2"
                    onclick={() => (selectedPreset = CUSTOM_PRESET)}
                >
                    <Icon icon="settings" class="text-xl" />
                    <div class="flex flex-col">
                        <span>Custom OIDC provider</span>
                        <span class="text-sm text-muted">
                            Connect via an OIDC discovery URL
                        </span>
                    </div>
                </button>
            </div>
        {:else}
            <form class="flex flex-col gap-6" onsubmit={createProvider}>
                <div class="flex flex-col gap-5">
                    {#if isCustom}
                        <div class="grid gap-2">
                            <Label for="customName">Provider name</Label>
                            <Input
                                bind:value={customName}
                                id="customName"
                                placeholder="my-provider"
                                required
                            />
                        </div>
                        <div class="grid gap-2">
                            <Label for="discoveryUri">
                                OIDC discovery URL
                            </Label>
                            <Input
                                bind:value={discoveryUri}
                                id="discoveryUri"
                                type="url"
                                class="font-mono"
                                placeholder="https://example.com/.well-known/openid-configuration"
                                required
                            />
                        </div>
                    {/if}
                    <div class="grid gap-2">
                        <Label for="clientId">Client ID</Label>
                        <Input
                            bind:value={clientId}
                            id="clientId"
                            class="font-mono"
                            autocomplete="off"
                            required
                        />
                    </div>
                    <div class="grid gap-2">
                        <Label for="clientSecret">Client secret</Label>
                        <Input
                            bind:value={clientSecret}
                            id="clientSecret"
                            type="password"
                            autocomplete="off"
                            required
                        />
                    </div>
                </div>

                <div class="flex items-center gap-2 justify-end">
                    <Button
                        type="button"
                        variant="secondary"
                        onclick={() => (selectedPreset = null)}
                    >
                        Back
                    </Button>
                    <Button type="submit" loading={creating}>
                        Add provider
                    </Button>
                </div>
            </form>
        {/if}
    </div>
</Modal>
