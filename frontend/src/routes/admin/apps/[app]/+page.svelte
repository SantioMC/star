<script lang="ts">
    import { goto } from "$app/navigation";
    import { page } from "$app/state";
    import * as star from "$lib/star";
    import {
        CLIENT_KIND_DESCRIPTIONS,
        CLIENT_KIND_OPTIONS,
        CONSENT_DESCRIPTIONS,
        CONSENT_OPTIONS,
        SIGNING_ALGORITHM_DESCRIPTIONS,
        SIGNING_ALGORITHM_OPTIONS,
        type ClientKind,
    } from "@/applications";
    import ApplicationScope from "@/components/application-scope.svelte";
    import Badge from "@/components/badge.svelte";
    import Button from "@/components/button.svelte";
    import Card from "@/components/card.svelte";
    import Categorizied from "@/components/categorizied.svelte";
    import Checkbox from "@/components/checkbox.svelte";
    import ClientSecretModal from "@/components/client-secret-modal.svelte";
    import ConfirmModal from "@/components/confirm-modal.svelte";
    import Copy from "@/components/copy.svelte";
    import Icon from "@/components/icon.svelte";
    import Notice from "@/components/notice.svelte";
    import Input from "@/components/input.svelte";
    import Menu, { type MenuItem } from "@/components/menu.svelte";
    import Modal from "@/components/modal.svelte";
    import RedirectUri from "@/components/provider/redirect-uri.svelte";
    import ScopeFields from "@/components/scope-fields.svelte";
    import ScopeItem from "@/components/scope-item.svelte";
    import Select from "@/components/select.svelte";
    import { Label } from "@/components/ui/label";
    import { hasPermission } from "@/index";
    import { formatTime } from "@/utils";
    import { onMount } from "svelte";
    import { toast } from "svelte-sonner";

    const MIN_DURATION = 60;

    let app = $state<star.AdminApplication | null>(null);

    let name = $state("");
    let namespace = $state("");
    let url = $state("");
    let kind = $state<ClientKind>("public");
    let consent = $state<star.ApplicationConsent>("default");
    let developmentMode = $state(false);
    let privilegedScopes = $state(false);
    let signingAlgorithm = $state<star.SigningAlgorithm>("ES512");
    let tokenDuration = $state("");
    let sessionDuration = $state("");
    let saving = $state(false);

    let redirectUri = $state("");
    let addingRedirectUri = $state(false);

    let scopeName = $state("");
    let scopeDescription = $state("");
    let scopeRequestable = $state(false);
    let addingScope = $state(false);

    let builtinExpanded = $state(false);

    let deleteAppModal = $state<ConfirmModal>();
    let rotateSecretModal = $state<ConfirmModal>();
    let addRedirectUriModal = $state<Modal>();
    let addScopeModal = $state<Modal>();
    let secretModal = $state<ClientSecretModal>();

    const canModify = $derived(
        app?.internal === false && hasPermission("application:write"),
    );

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

    const resetForm = (data: star.AdminApplication) => {
        name = data.name;
        namespace = data.namespace;
        url = data.application_url;
        kind = data.kind;
        consent = data.consent;
        developmentMode = data.development_mode;
        privilegedScopes = data.privileged_scopes;
        signingAlgorithm = data.signing_algorithm;
        tokenDuration = data.token_duration.toString();
        sessionDuration = data.session_duration.toString();
    };

    const refresh = async () => {
        if (!app) return;

        const data = await star.getApplication(app.id);
        if (data) app = data;
    };

    onMount(async () => {
        if (!hasPermission("application:read")) return goto("/");

        const data = await star.getApplication(page.params.app!);
        if (!data) return goto("/admin/apps");

        app = data;
        resetForm(data);
    });

    const parseDuration = (value: string, label: string): number | null => {
        const parsed = Number(value);

        if (!Number.isInteger(parsed) || parsed < MIN_DURATION) {
            toast.error(`${label} must be a whole number of at least ${MIN_DURATION} seconds`);
            return null;
        }

        return parsed;
    };

    const updateApp = async (event: Event) => {
        event.preventDefault();
        if (!app || saving) return;

        const token = parseDuration(tokenDuration, "Token duration");
        if (token == null) return;

        if (token > 3600) {
            return toast.error("Token duration may not exceed 3600 seconds (1 hour)");
        }

        const session = parseDuration(sessionDuration, "Session duration");
        if (session == null) return;

        saving = true;
        const result = await star.updateApplication(app.id, {
            name: name.trim(),
            namespace: namespace.trim(),
            url: url.trim(),
            kind,
            consent,
            development_mode: developmentMode,
            privileged_scopes: privilegedScopes,
            signing_algorithm: signingAlgorithm,
            token_duration: token,
            session_duration: session,
        });
        saving = false;

        if ("error" in result) {
            return toast.error(result.message);
        }

        await refresh();
        if (app) resetForm(app);
        if (result.client_secret) secretModal?.show(result.client_secret);

        toast.success("Successfully updated application");
    };

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

    const rotateSecret = async () => {
        if (!app) return;

        const result = await star.rotateApplicationSecret(app.id);
        if ("error" in result) {
            toast.error(result.message);
            return false;
        }

        await refresh();
        if (result.client_secret) secretModal?.show(result.client_secret);

        toast.success("Successfully rotated the client secret");
    };

    const openAddRedirectUri = () => {
        redirectUri = "";
        addRedirectUriModal?.open();
    };

    const addRedirectUri = async (event: Event) => {
        event.preventDefault();
        if (!app || addingRedirectUri) return;

        const trimmed = redirectUri.trim();
        if (!trimmed) return;

        if (app.redirect_uris.includes(trimmed)) {
            return toast.error("This redirect URI is already registered");
        }

        addingRedirectUri = true;
        const result = await star.updateApplication(app.id, {
            redirect_uris: [...app.redirect_uris, trimmed],
        });
        addingRedirectUri = false;

        if ("error" in result) {
            return toast.error(result.message);
        }

        await refresh();
        addRedirectUriModal?.close();
        toast.success("Successfully added redirect URI");
    };

    const removeRedirectUri = async (uri: string) => {
        if (!app) return false;

        const result = await star.updateApplication(app.id, {
            redirect_uris: app.redirect_uris.filter((it) => it !== uri),
        });

        if ("error" in result) {
            toast.error(result.message);
            return false;
        }

        await refresh();
        toast.success("Removed redirect URI");
    };

    const openAddScope = () => {
        scopeName = `${app?.namespace ?? ""}:`;
        scopeDescription = "";
        scopeRequestable = false;
        addScopeModal?.open();
    };

    const addScope = async (event: Event) => {
        event.preventDefault();
        if (!app || addingScope) return;

        addingScope = true;
        const result = await star.createApplicationScope(app.id, {
            name: scopeName.trim(),
            description: scopeDescription.trim(),
            requestable: scopeRequestable,
        });
        addingScope = false;

        if ("error" in result) {
            return toast.error(result.message);
        }

        await refresh();
        addScopeModal?.close();
        toast.success("Successfully added scope");
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

    {#if app.internal}
        <Notice icon="lock" title="Built-in application.">
            This application is managed by Star and can not be modified or
            deleted.
        </Notice>
    {/if}

    <Card
        title="OAuth 2.0 Credentials"
        icon="key"
        description="The credentials this application uses to identify itself against Star's OAuth 2.0 endpoints."
        class="flex flex-col gap-5"
    >
        {#snippet actions()}
            {#if canModify && app!.kind === "confidential"}
                <Button
                    variant="secondary"
                    onclick={() => rotateSecretModal?.open()}
                >
                    <Icon icon="refresh" />
                    Rotate secret
                </Button>
            {/if}
        {/snippet}

        <div class="grid gap-4 sm:grid-cols-2">
            <Categorizied title="Client ID">
                <div class="flex min-w-0 items-center gap-1.5">
                    <code class="truncate font-mono text-sm">
                        {app.client_id}
                    </code>
                    <Copy content={app.client_id} />
                </div>
            </Categorizied>

            <Categorizied title="Client Secret">
                {#if app.kind === "public"}
                    <p class="text-sm text-muted">
                        None, public clients must use PKCE
                    </p>
                {:else if app.has_client_secret}
                    <p class="text-sm text-muted">
                        Hidden, rotate the secret to issue a new one
                    </p>
                {:else}
                    <p class="text-sm text-muted">Unset</p>
                {/if}
            </Categorizied>
        </div>
    </Card>

    <Card
        title="Redirect URIs"
        icon="world"
        description="Exact URIs this application may redirect to during the authorization code flow."
        class="flex flex-col gap-3"
    >
        {#snippet actions()}
            {#if canModify}
                <Button variant="secondary" onclick={openAddRedirectUri}>
                    Add URI
                </Button>
            {/if}
        {/snippet}

        <div class="flex flex-col [&>div~div]:border-t [&>div~div]:border-background-200">
            {#each app.redirect_uris as uri (uri)}
                <RedirectUri
                    redirectUri={uri}
                    onremove={canModify
                        ? () => removeRedirectUri(uri)
                        : undefined}
                />
            {:else}
                <p class="py-6 text-center text-sm text-muted">
                    No redirect URIs are registered, authorization will fail
                    until one is added.
                </p>
            {/each}
        </div>
    </Card>

    <Card
        title="Application Scopes"
        icon="shield-lock"
        description="Scopes this application can request on top of Star's built-in ones. Descriptions are shown to users on the consent page."
        class="flex flex-col gap-3"
    >
        {#snippet actions()}
            {#if canModify}
                <Button variant="secondary" onclick={openAddScope}>
                    Add scope
                </Button>
            {/if}
        {/snippet}

        <div class="flex flex-col [&>div~div]:border-t [&>div~div]:border-background-200">
            {#each app.scopes as scope (scope.id)}
                <ApplicationScope
                    application={app.id}
                    namespace={app.namespace}
                    {scope}
                    {canModify}
                    onchange={refresh}
                />
            {:else}
                <p class="py-6 text-center text-sm text-muted">
                    This application has no scopes of its own. It can still
                    request Star's built-in scopes.
                </p>
            {/each}
        </div>

        {#if app.builtin_scopes && app.builtin_scopes.length > 0}
            <button
                type="button"
                aria-expanded={builtinExpanded}
                onclick={() => (builtinExpanded = !builtinExpanded)}
                class="flex w-fit items-center gap-1.5 rounded-md px-1 py-1 text-sm text-muted transition-colors hover:text-foreground"
            >
                <Icon
                    icon="chevron-right"
                    class="transition-transform {builtinExpanded
                        ? 'rotate-90'
                        : ''}"
                />
                {builtinExpanded ? "Hide" : "Show"} built-in scopes
                <span class="font-mono text-xs">
                    ({app.builtin_scopes.length})
                </span>
            </button>

            {#if builtinExpanded}
                <div
                    class="flex flex-col [&>div~div]:border-t [&>div~div]:border-background-200 rounded-lg border border-background-200 px-3 animate-in fade-in-0"
                >
                    {#each app.builtin_scopes as scope (scope.name)}
                        <ScopeItem
                            name={scope.name}
                            description={scope.description}
                        >
                            {#snippet badges()}
                                {#if scope.restricted}
                                    <Badge tone="danger">restricted</Badge>
                                {/if}
                                {#if scope.privileged}
                                    <Badge tone="warning">privileged</Badge>
                                {/if}
                                {#each scope.aliases ?? [] as alias (alias)}
                                    <Badge>{alias}</Badge>
                                {/each}
                            {/snippet}
                        </ScopeItem>
                    {/each}
                </div>
            {/if}
        {/if}
    </Card>

    <Card title="Application Settings" icon="settings">
        <form onsubmit={updateApp} class="mt-6">
            <fieldset disabled={!canModify} class="flex flex-col gap-5">
                <div class="grid gap-2">
                    <Label for="name">Name</Label>
                    <Input bind:value={name} id="name" type="text" required />
                </div>

                <div class="grid gap-2">
                    <Label for="namespace">Scope namespace</Label>
                    <Input
                        bind:value={namespace}
                        id="namespace"
                        type="text"
                        class="font-mono"
                        required
                    />
                    <p class="text-xs text-muted">
                        Every scope this application registers must start with
                        <span class="font-mono text-foreground"
                            >{namespace || "<namespace>"}:</span
                        >
                    </p>
                </div>

                <div class="grid gap-2">
                    <Label for="url">Application URL</Label>
                    <Input
                        bind:value={url}
                        id="url"
                        type="url"
                        class="font-mono"
                        required
                    />
                </div>

                <div class="grid gap-2">
                    <Label for="kind">Client type</Label>
                    <Select
                        bind:value={kind}
                        id="kind"
                        options={CLIENT_KIND_OPTIONS}
                        disabled={!canModify}
                    />
                    <p class="text-xs text-muted">
                        {CLIENT_KIND_DESCRIPTIONS[kind]}
                    </p>
                </div>

                <div class="grid gap-2">
                    <Label for="consent">Consent</Label>
                    <Select
                        bind:value={consent}
                        id="consent"
                        options={CONSENT_OPTIONS}
                        disabled={!canModify}
                    />
                    <p class="text-xs text-muted">
                        {CONSENT_DESCRIPTIONS[consent]}
                    </p>
                </div>

                <div class="grid gap-2">
                    <Label for="signingAlgorithm">Signing algorithm</Label>
                    <Select
                        bind:value={signingAlgorithm}
                        id="signingAlgorithm"
                        options={SIGNING_ALGORITHM_OPTIONS}
                        disabled={!canModify}
                    />
                    <p class="text-xs text-muted">
                        {SIGNING_ALGORITHM_DESCRIPTIONS[signingAlgorithm]}
                    </p>
                </div>

                <div class="flex flex-col gap-5 sm:flex-row sm:gap-6">
                    <div class="flex-1/2 gap-2 grid">
                        <Label for="tokenDuration">Token duration</Label>
                        <div class="relative">
                            <Input
                                bind:value={tokenDuration}
                                id="tokenDuration"
                                type="number"
                                min={MIN_DURATION}
                                max={3600}
                                class="pr-20 font-mono input-clean"
                            />
                            <span
                                class="pointer-events-none absolute top-1/2 right-3 -translate-y-1/2 text-xs text-muted select-none"
                            >
                                seconds
                            </span>
                        </div>
                    </div>

                    <div class="flex-1/2 gap-2 grid">
                        <Label for="sessionDuration">Session duration</Label>
                        <div class="relative">
                            <Input
                                bind:value={sessionDuration}
                                id="sessionDuration"
                                type="number"
                                min={MIN_DURATION}
                                class="pr-20 font-mono input-clean"
                            />
                            <span
                                class="pointer-events-none absolute top-1/2 right-3 -translate-y-1/2 text-xs text-muted select-none"
                            >
                                seconds
                            </span>
                        </div>
                    </div>
                </div>

                <label
                    class="flex items-start gap-3 rounded-lg border border-background-200 p-3 transition-colors has-checked:border-primary/40 has-checked:bg-hover has-enabled:cursor-pointer has-enabled:hover:bg-hover"
                >
                    <Checkbox class="mt-0.5" bind:checked={developmentMode} />
                    <div class="flex flex-col gap-0.5">
                        <span class="text-sm select-none">Development mode</span>
                        <span class="text-sm text-muted select-none">
                            Flags this application as under development.
                        </span>
                    </div>
                </label>

                {#if !app.internal}
                    <label
                        class="flex items-start gap-3 rounded-lg border border-background-200 p-3 transition-colors has-checked:border-red-400/40 has-checked:bg-red-400/5 has-enabled:cursor-pointer has-enabled:hover:bg-hover"
                    >
                        <Checkbox class="mt-0.5" bind:checked={privilegedScopes} />
                        <div class="flex flex-col gap-0.5">
                            <span class="text-sm select-none">
                                Allow privileged scopes
                            </span>
                            <span class="text-sm text-muted select-none">
                                Lets this application request Star scopes that
                                modify accounts, such as changing usernames, emails,
                                and sessions. Only enable this for applications you
                                fully trust.
                            </span>
                        </div>
                    </label>
                {/if}
            </fieldset>

            {#if canModify}
                <div class="mt-4 flex items-center gap-2 justify-end">
                    <Button type="submit" loading={saving}>Update</Button>
                </div>
            {/if}
        </form>
    </Card>

    {#if canModify}
        <Modal bind:this={addRedirectUriModal} button>
            <form class="flex flex-col gap-6" onsubmit={addRedirectUri}>
                <div class="flex flex-col gap-1 pr-8">
                    <h1 class="text-xl font-semibold">Add redirect URI</h1>
                    <h2 class="text-sm text-muted">
                        Users may be redirected to this exact URI after they
                        authorize the application.
                    </h2>
                </div>

                <div class="grid gap-2">
                    <Label for="redirectUri">Redirect URI</Label>
                    <Input
                        bind:value={redirectUri}
                        id="redirectUri"
                        type="url"
                        class="font-mono"
                        placeholder="https://example.com/callback"
                        required
                    />
                </div>

                <div class="flex items-center gap-2 justify-end">
                    <Button
                        type="button"
                        variant="secondary"
                        onclick={() => addRedirectUriModal?.close()}
                    >
                        Cancel
                    </Button>
                    <Button type="submit" loading={addingRedirectUri}>
                        Add URI
                    </Button>
                </div>
            </form>
        </Modal>

        <Modal bind:this={addScopeModal} button>
            <form class="flex flex-col gap-6" onsubmit={addScope}>
                <div class="flex flex-col gap-1 pr-8">
                    <h1 class="text-xl font-semibold">Add scope</h1>
                    <h2 class="text-sm text-muted">
                        Scopes must start with
                        <span class="font-mono text-foreground"
                            >{app.namespace}:</span
                        >. The description is shown to users when they
                        authorize the application.
                    </h2>
                </div>

                <ScopeFields
                    id="add"
                    namespace={app.namespace}
                    bind:name={scopeName}
                    bind:description={scopeDescription}
                    bind:requestable={scopeRequestable}
                />

                <div class="flex items-center gap-2 justify-end">
                    <Button
                        type="button"
                        variant="secondary"
                        onclick={() => addScopeModal?.close()}
                    >
                        Cancel
                    </Button>
                    <Button type="submit" loading={addingScope}>
                        Add scope
                    </Button>
                </div>
            </form>
        </Modal>

        <ConfirmModal
            bind:this={rotateSecretModal}
            title="Rotate client secret"
            confirmLabel="Rotate secret"
            onconfirm={rotateSecret}
        >
            The current secret stops working immediately.
            <span class="text-foreground">{app.name}</span> will fail to
            authenticate until it is updated with the new secret.
        </ConfirmModal>
    {/if}

    <ClientSecretModal bind:this={secretModal} />

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
