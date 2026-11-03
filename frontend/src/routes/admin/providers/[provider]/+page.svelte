<script lang="ts">
    import { branding } from "@/branding";
    import { goto } from "$app/navigation";
    import { page } from "$app/state";
    import * as star from "$lib/star";
    import Badge from "@/components/badge.svelte";
    import Button from "@/components/button.svelte";
    import Card from "@/components/card.svelte";
    import ConfirmModal from "@/components/confirm-modal.svelte";
    import Checkbox from "@/components/checkbox.svelte";
    import Copy from "@/components/copy.svelte";
    import Icon from "@/components/icon.svelte";
    import Input from "@/components/input.svelte";
    import Menu, { type MenuItem } from "@/components/menu.svelte";
    import Modal from "@/components/modal.svelte";
    import { Label } from "@/components/ui/label";
    import { hasPermission } from "@/index";
    import { formatTime, titleCase } from "@/utils";
    import { onMount } from "svelte";
    import { toast } from "svelte-sonner";

    const canWrite = hasPermission("admin:provider:write");
    const canDelete = hasPermission("admin:provider:delete");

    let provider = $state<star.SocialProviderConfig | null>(null);

    let icon = $state("");
    let order = $state("");
    let trustEmail = $state(false);
    let saving = $state(false);

    let clientId = $state("");
    let clientSecret = $state("");
    let savingCredentials = $state(false);

    let credentialsModal = $state<Modal>();
    let deleteModal = $state<ConfirmModal>();

    const endpoints = $derived.by(() => {
        if (!provider) return [];

        return [
            { label: "Discovery URL", value: provider.oidc_url },
            { label: "Authorization URL", value: provider.authorization_url },
            { label: "Token URL", value: provider.token_url },
            { label: "Userinfo URL", value: provider.userinfo_url },
        ].filter(
            (it): it is { label: string; value: string } => !!it.value,
        );
    });

    const menuItems = $derived.by(() => {
        const items: MenuItem[] = [];

        if (canWrite) {
            items.push({
                label: "Update credentials",
                icon: "key",
                onclick: openCredentialsModal,
            });
        }

        if (canDelete) {
            items.push({
                label: "Delete provider",
                icon: "trash",
                danger: true,
                onclick: () => deleteModal?.open(),
            });
        }

        return items;
    });

    const resetForm = (data: star.SocialProviderConfig) => {
        icon = data.icon;
        order = String(data.order);
        trustEmail = data.trust_email;
    };

    onMount(async () => {
        if (!hasPermission("admin:provider:read")) return goto("/");

        const data = await star.getSocialProvider(page.params.provider!);
        if (!data) return goto("/admin/providers");

        provider = data;
        resetForm(data);
    });

    const updateDisplay = async (event: Event) => {
        event.preventDefault();
        if (!provider || saving) return;

        saving = true;
        const result = await star.updateSocialProvider(provider.id, {
            icon: icon.trim() || undefined,
            order: Number(order),
            trust_email: trustEmail,
        });
        saving = false;

        if ("error" in result) {
            resetForm(provider);
            return toast.error(result.message);
        }

        provider = result;
        toast.success("Successfully updated provider");
    };

    const openCredentialsModal = () => {
        clientId = "";
        clientSecret = "";
        credentialsModal?.open();
    };

    const updateCredentials = async (event: Event) => {
        event.preventDefault();
        if (!provider || savingCredentials) return;

        const changes: star.UpdateSocialProviderRequest = {};
        if (clientId.trim()) changes.client_id = clientId.trim();
        if (clientSecret) changes.client_secret = clientSecret;

        if (Object.keys(changes).length === 0) {
            credentialsModal?.close();
            return;
        }

        savingCredentials = true;
        const result = await star.updateSocialProvider(provider.id, changes);
        savingCredentials = false;

        if ("error" in result) {
            return toast.error(result.message);
        }

        provider = result;
        credentialsModal?.close();
        toast.success("Successfully updated credentials");
    };

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

    <Card
        title="Credentials"
        icon="key"
        description="The OAuth client issued to {$branding.name} by this provider."
        class="flex flex-col gap-4"
    >
        {#snippet actions()}
            {#if canWrite}
                <Button variant="secondary" onclick={openCredentialsModal}>
                    Update
                </Button>
            {/if}
        {/snippet}

        <dl class="grid gap-4 sm:grid-cols-2">
            <div class="min-w-0">
                <dt class="text-xs text-muted select-none">Client ID</dt>
                <dd
                    class="mt-1 flex items-center gap-2 font-mono text-sm break-all"
                >
                    {provider.client_id}
                    <Copy content={provider.client_id} class="text-muted" />
                </dd>
            </div>
            <div class="min-w-0">
                <dt class="text-xs text-muted select-none">Client Secret</dt>
                <dd class="mt-1">
                    {#if provider.has_client_secret}
                        <Badge tone="success">configured</Badge>
                    {:else}
                        <Badge tone="warning">missing</Badge>
                    {/if}
                </dd>
            </div>
            <div class="min-w-0 sm:col-span-2">
                <dt class="text-xs text-muted select-none">Callback URL</dt>
                <dd
                    class="mt-1 flex items-center gap-2 font-mono text-sm break-all"
                >
                    {star.getProviderCallbackUrl(provider.id)}
                    <Copy
                        content={star.getProviderCallbackUrl(provider.id)}
                        class="text-muted"
                    />
                </dd>
            </div>
            {#if provider.scopes?.length}
                <div class="min-w-0 sm:col-span-2">
                    <dt class="text-xs text-muted select-none">Scopes</dt>
                    <dd class="mt-1.5 flex flex-wrap gap-1">
                        {#each provider.scopes as scope (scope)}
                            <Badge>{scope}</Badge>
                        {/each}
                    </dd>
                </div>
            {/if}
        </dl>
    </Card>

    {#if endpoints.length > 0}
        <Card
            title="Endpoints"
            icon="link"
            description="The OAuth and OpenID Connect endpoints {$branding.name} calls for this provider."
            class="flex flex-col gap-4"
        >
            <dl class="flex flex-col gap-4">
                {#each endpoints as endpoint (endpoint.label)}
                    <div class="min-w-0">
                        <dt class="text-xs text-muted select-none">
                            {endpoint.label}
                        </dt>
                        <dd
                            class="mt-1 flex items-center gap-2 font-mono text-sm break-all"
                        >
                            {endpoint.value}
                            <Copy
                                content={endpoint.value}
                                class="text-muted"
                            />
                        </dd>
                    </div>
                {/each}
            </dl>
        </Card>
    {/if}

    <Card
        title="Provider Settings"
        icon="settings"
        description="How this provider appears on the login page and how its accounts are trusted."
    >
        <form onsubmit={updateDisplay} class="mt-6">
            <div class="flex flex-col gap-5">
                <div class="grid gap-2">
                    <Label for="icon">Icon</Label>
                    <div class="flex items-center gap-3">
                        <div
                            class="flex size-9 shrink-0 items-center justify-center rounded-md border border-background-200"
                        >
                            <Icon
                                icon={icon || "plug-connected"}
                                class="text-lg"
                            />
                        </div>
                        <Input
                            bind:value={icon}
                            id="icon"
                            class="font-mono"
                            placeholder="brand-discord"
                            disabled={!canWrite}
                        />
                    </div>
                    <p class="text-xs text-muted select-none">
                        Any Tabler icon name
                    </p>
                </div>

                <div class="grid gap-2">
                    <Label for="order">Display order</Label>
                    <Input
                        bind:value={order}
                        id="order"
                        type="number"
                        min="0"
                        class="font-mono"
                        disabled={!canWrite}
                    />
                    <p class="text-xs text-muted select-none">
                        Lower numbers appear first
                    </p>
                </div>

                <label
                    class="flex items-start gap-3 rounded-lg border border-background-200 p-3 transition-colors has-checked:border-red-400/40 has-checked:bg-red-400/5 has-enabled:cursor-pointer has-enabled:hover:bg-hover"
                >
                    <Checkbox
                        class="mt-0.5"
                        bind:checked={trustEmail}
                        disabled={!canWrite}
                    />
                    <div class="flex flex-col gap-0.5">
                        <span class="text-sm select-none">
                            Trust verified emails
                        </span>
                        <span class="text-sm text-muted select-none">
                            New accounts created through this provider keep the
                            email it reports as verified. Applications may link
                            accounts by verified email, so only enable this for
                            providers that confirm email ownership themselves.
                        </span>
                    </div>
                </label>
            </div>

            {#if canWrite}
                <div class="mt-4 flex items-center gap-2 justify-end">
                    <Button type="submit" loading={saving}>Update</Button>
                </div>
            {/if}
        </form>
    </Card>

    <Modal bind:this={credentialsModal} button>
        <form class="flex flex-col gap-6" onsubmit={updateCredentials}>
            <div class="flex flex-col gap-1 pr-8">
                <h1 class="text-xl font-semibold">Update credentials</h1>
                <h2 class="text-sm text-muted">
                    Leave a field empty to keep its current value.
                </h2>
            </div>

            <div class="flex flex-col gap-5">
                <div class="grid gap-2">
                    <Label for="newClientId">Client ID</Label>
                    <Input
                        bind:value={clientId}
                        id="newClientId"
                        class="font-mono"
                        autocomplete="off"
                        placeholder={provider.client_id}
                    />
                </div>
                <div class="grid gap-2">
                    <Label for="newClientSecret">Client secret</Label>
                    <Input
                        bind:value={clientSecret}
                        id="newClientSecret"
                        type="password"
                        autocomplete="off"
                    />
                </div>
            </div>

            <div class="flex items-center gap-2 justify-end">
                <Button
                    type="button"
                    variant="secondary"
                    onclick={() => credentialsModal?.close()}
                >
                    Cancel
                </Button>
                <Button type="submit" loading={savingCredentials}>
                    Update credentials
                </Button>
            </div>
        </form>
    </Modal>

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
