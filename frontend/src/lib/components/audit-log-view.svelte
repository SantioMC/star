<script lang="ts">
    import { goto } from "$app/navigation";
    import { resolve } from "$app/paths";
    import * as star from "$lib/star";
    import { hasPermission } from "@/index";
    import { formatTime } from "@/utils";
    import { onMount, type Snippet } from "svelte";
    import { toast } from "svelte-sonner";
    import Badge, { type BadgeTone } from "./badge.svelte";
    import Card from "./card.svelte";
    import Icon from "./icon.svelte";
    import Input from "./input.svelte";
    import Modal from "./modal.svelte";
    import Pagination from "./pagination.svelte";
    import SearchInput from "./search-input.svelte";
    import Select, { type SelectOption } from "./select.svelte";

    const {
        filter,
    }: {
        filter?: Pick<star.AuditLogFilter, "user" | "application" | "provider">;
    } = $props();

    const categoryIcons: Record<string, string> = {
        Sessions: "login-2",
        Passwords: "password",
        Emails: "mail",
        "Multi-factor": "shield-lock",
        Authorizations: "plug-connected",
        Users: "user",
        Roles: "shield",
        Applications: "apps",
        Providers: "world",
        "Signing Keys": "key",
        Tokens: "ban",
        Security: "alert-triangle",
        Settings: "settings",
    };

    const severityTones: Record<star.AuditSeverity, BadgeTone> = {
        info: "neutral",
        notice: "warning",
        critical: "danger",
    };

    const severityIcons: Record<star.AuditSeverity, string> = {
        info: "bg-background-200 text-foreground/80",
        notice: "bg-amber-400/10 text-amber-300",
        critical: "bg-red-400/10 text-red-300",
    };

    const ranges: SelectOption[] = [
        { value: "all", label: "Any time" },
        { value: "1", label: "Last 24 hours" },
        { value: "7", label: "Last 7 days" },
        { value: "30", label: "Last 30 days" },
        { value: "90", label: "Last 90 days" },
        { value: "custom", label: "Custom range" },
    ];

    let definitions = $state<star.AuditActionDefinition[]>([]);
    let logs = $state<star.AuditLog[]>([]);
    let next = $state<string | null>(null);
    let loaded = $state(false);

    let query = $state("");
    let category = $state("");
    let action = $state("");
    let range = $state("all");
    let from = $state("");
    let to = $state("");

    let cursors = $state<string[]>([]);
    let selected = $state<star.AuditLog | null>(null);
    let detailModal = $state<Modal>();

    let searchTask: ReturnType<typeof setTimeout> | null = null;
    let latestRequest = 0;

    const definitionIndex = $derived(
        new Map(definitions.map((it) => [it.name, it])),
    );

    const categoryOptions = $derived<SelectOption[]>([
        { value: "", label: "All categories" },
        ...[...new Set(definitions.map((it) => it.category))].map((it) => ({
            value: it,
            label: it,
        })),
    ]);

    const actionOptions = $derived<SelectOption[]>([
        { value: "", label: "All actions" },
        ...definitions
            .filter((it) => !category || it.category === category)
            .map((it) => ({ value: it.name, label: it.label })),
    ]);

    const filtered = $derived(
        query.trim() !== "" || category !== "" || action !== "" || range !== "all",
    );

    const rows = $derived.by(() => {
        const grouped: { day: string; items: star.AuditLog[] }[] = [];

        for (const log of logs) {
            const day = dayLabel(log.created_at);
            const group = grouped.at(-1);

            if (group?.day === day) group.items.push(log);
            else grouped.push({ day, items: [log] });
        }

        return grouped;
    });

    const dayLabel = (time: string) =>
        new Date(time).toLocaleDateString(undefined, {
            weekday: "long",
            month: "long",
            day: "numeric",
            year: "numeric",
        });

    const clockLabel = (time: string) =>
        new Date(time).toLocaleTimeString(undefined, {
            hour: "numeric",
            minute: "2-digit",
            second: "2-digit",
        });

    const definitionOf = (log: star.AuditLog) =>
        definitionIndex.get(log.action);

    const timeBounds = () => {
        if (range === "custom") {
            return {
                since: from ? new Date(from).toISOString() : undefined,
                until: to ? new Date(to).toISOString() : undefined,
            };
        }

        if (range === "all") return {};

        const age = Number(range) * 86_400_000;
        return { since: new Date(Date.now() - age).toISOString() };
    };

    const load = async () => {
        const request = ++latestRequest;
        loaded = false;

        const result = await star
            .getAuditLogs({
                ...filter,
                query: query.trim(),
                category,
                action,
                ...timeBounds(),
                before: cursors.at(-1),
            })
            .catch(() => null);

        if (request !== latestRequest) return;
        loaded = true;

        if (result === null || "error" in result) {
            logs = [];
            next = null;
            return toast.error(result?.message ?? "Failed to load the audit log");
        }

        logs = result.items;
        next = result.next;
    };

    const reload = () => {
        cursors = [];
        load();
    };

    const handleSearch = () => {
        if (searchTask) clearTimeout(searchTask);
        searchTask = setTimeout(reload, 250);
    };

    const handleCategory = () => {
        action = "";
        reload();
    };

    const clearFilters = () => {
        query = "";
        category = "";
        action = "";
        range = "all";
        from = "";
        to = "";
        reload();
    };

    const nextPage = () => {
        if (!next) return;

        cursors = [...cursors, next];
        load();
    };

    const previousPage = () => {
        cursors = cursors.slice(0, -1);
        load();
    };

    const openDetails = (log: star.AuditLog) => {
        selected = log;
        detailModal?.open();
    };

    const displayValue = (value: unknown) =>
        typeof value === "string" ? value : JSON.stringify(value);

    onMount(async () => {
        if (!hasPermission("admin:audit:read")) return goto("/");

        load();
        definitions = await star.getAuditActions();
    });
</script>

{#snippet person(id: string, name: string | null)}
    {#if name}
        <span class="block truncate">{name}</span>
    {:else}
        <span class="block truncate font-mono text-xs">{id}</span>
    {/if}
{/snippet}

{#snippet fact(label: string, body: Snippet, wide = false)}
    <div
        class="bg-background-100 px-4 py-3 min-w-0 {wide
            ? 'sm:col-span-2'
            : ''}"
    >
        <dt class="text-xs text-muted select-none">{label}</dt>
        <dd class="mt-0.5 text-sm wrap-anywhere">{@render body()}</dd>
    </div>
{/snippet}

{#snippet userLink(id: string, name: string | null)}
    <a
        href={resolve("/admin/users/[user]", { user: id })}
        class="text-primary hover:underline"
    >
        {#if name}
            {name}
        {:else}
            <span class="font-mono text-xs">{id}</span>
        {/if}
    </a>
    {#if name}
        <span class="block font-mono text-xs text-muted">{id}</span>
    {/if}
{/snippet}

<div class="flex flex-col gap-3">
    <div class="flex flex-col gap-2">
        <SearchInput
            bind:value={query}
            oninput={handleSearch}
            placeholder="Search by user, email, ID or IP address"
            class="sm:max-w-none"
        />
        <div class="grid grid-cols-1 gap-2 sm:grid-cols-3">
            <Select
                bind:value={category}
                options={categoryOptions}
                onchange={handleCategory}
            />
            <Select
                bind:value={action}
                options={actionOptions}
                onchange={reload}
            />
            <Select bind:value={range} options={ranges} onchange={reload} />
        </div>
    </div>

    {#if range === "custom"}
        <div class="grid gap-2 sm:grid-cols-2">
            <label class="grid gap-1 text-xs text-muted select-none">
                From
                <Input
                    bind:value={from}
                    type="datetime-local"
                    max={to || undefined}
                    onchange={reload}
                    class="scheme-dark"
                />
            </label>
            <label class="grid gap-1 text-xs text-muted select-none">
                To
                <Input
                    bind:value={to}
                    type="datetime-local"
                    min={from || undefined}
                    onchange={reload}
                    class="scheme-dark"
                />
            </label>
        </div>
    {/if}
</div>

<Card class="overflow-hidden p-0">
    <div class="overflow-x-auto">
        <table class="w-full text-sm">
            <thead>
                <tr class="hover:bg-transparent">
                    <th class="px-4 py-2.5 text-xs font-medium text-muted select-none">
                        Event
                    </th>
                    <th class="px-4 py-2.5 text-xs font-medium text-muted select-none max-sm:hidden">
                        Actor
                    </th>
                    <th class="px-4 py-2.5 text-xs font-medium text-muted select-none max-lg:hidden">
                        Target
                    </th>
                    <th class="px-4 py-2.5 text-right! text-xs font-medium text-muted select-none max-sm:hidden">
                        Time
                    </th>
                </tr>
            </thead>
            <tbody aria-busy={!loaded}>
                {#each rows as group (group.day)}
                    <tr class="bg-background hover:bg-background!">
                        <td
                            colspan="4"
                            class="px-4 py-1.5 text-xs font-medium text-muted select-none"
                        >
                            {group.day}
                        </td>
                    </tr>

                    {#each group.items as log (log.id)}
                        {@const definition = definitionOf(log)}
                        {@const severity = definition?.severity ?? "info"}
                        <tr
                            class="cursor-pointer hover:bg-background-200/50! {loaded
                                ? ''
                                : 'opacity-60'}"
                            onclick={() => openDetails(log)}
                        >
                            <td class="px-4 py-2.5">
                                <div class="flex items-center gap-3 min-w-0">
                                    <span
                                        class="grid size-8 shrink-0 place-items-center rounded-md {severityIcons[
                                            severity
                                        ]}"
                                    >
                                        <Icon
                                            icon={categoryIcons[
                                                definition?.category ?? ""
                                            ] ?? "activity"}
                                        />
                                    </span>
                                    <div class="min-w-0">
                                        <button
                                            type="button"
                                            class="flex items-center gap-2 text-left font-medium"
                                            onclick={(event) => {
                                                event.stopPropagation();
                                                openDetails(log);
                                            }}
                                        >
                                            <span class="truncate">
                                                {definition?.label ?? log.action}
                                            </span>
                                            {#if severity === "critical"}
                                                <Badge tone="danger">critical</Badge>
                                            {/if}
                                        </button>
                                        <p class="truncate font-mono text-xs text-muted">
                                            <span class="sm:hidden">
                                                {clockLabel(log.created_at)} &middot;
                                            </span>
                                            {log.action}
                                        </p>
                                    </div>
                                </div>
                            </td>
                            <td class="px-4 max-sm:hidden max-w-52">
                                {#if log.actor_id}
                                    {@render person(log.actor_id, log.actor_name)}
                                {:else}
                                    <span class="text-muted">Unknown</span>
                                {/if}
                            </td>
                            <td class="px-4 max-lg:hidden max-w-52">
                                {#if log.target_id}
                                    {@render person(log.target_id, log.target_name)}
                                {:else}
                                    <span class="text-muted">&mdash;</span>
                                {/if}
                            </td>
                            <td class="px-4 text-right text-muted whitespace-nowrap max-sm:hidden">
                                {clockLabel(log.created_at)}
                            </td>
                        </tr>
                    {/each}
                {:else}
                    <tr class="hover:bg-transparent">
                        <td colspan="4" class="py-14 text-center text-muted">
                            {#if !loaded}
                                Loading audit logs...
                            {:else if filtered}
                                <p>No audit logs match these filters.</p>
                                <button
                                    type="button"
                                    class="mt-1 text-primary hover:underline"
                                    onclick={clearFilters}
                                >
                                    Clear filters
                                </button>
                            {:else}
                                No audit logs have been recorded yet.
                            {/if}
                        </td>
                    </tr>
                {/each}
            </tbody>
        </table>
    </div>
</Card>

<Pagination
    page={cursors.length + 1}
    hasNext={next !== null}
    {loaded}
    onprevious={previousPage}
    onnext={nextPage}
/>

<Modal bind:this={detailModal} button class="sm:w-xl">
    {#if selected}
        {@const log = selected}
        {@const definition = definitionOf(log)}
        <div class="flex flex-col gap-5">
            <div class="flex flex-col gap-1 pr-8">
                <h1 class="text-xl font-semibold">
                    {definition?.label ?? log.action}
                </h1>
                <div class="flex flex-wrap items-center gap-2 text-sm text-muted">
                    <span class="font-mono text-xs">{log.action}</span>
                    {#if definition}
                        <Badge tone={severityTones[definition.severity]}>
                            {definition.severity}
                        </Badge>
                    {/if}
                </div>
                <p class="text-sm text-muted">
                    {formatTime(new Date(log.created_at))}
                </p>
            </div>

            <dl
                class="grid gap-px overflow-hidden rounded-lg border border-background-200 bg-background-200 sm:grid-cols-2"
            >
                {#snippet actorFact()}
                    {#if log.actor_id}
                        {@render userLink(log.actor_id, log.actor_name)}
                    {:else}
                        <span class="text-muted">Unknown</span>
                    {/if}
                {/snippet}
                {@render fact("Actor", actorFact)}

                {#snippet targetFact()}
                    {#if log.target_id && log.target_type === "user"}
                        {@render userLink(log.target_id, log.target_name)}
                    {:else if log.target_id}
                        {log.target_name ?? ""}
                        <span class="block font-mono text-xs text-muted">
                            {log.target_id}
                        </span>
                    {:else}
                        <span class="text-muted">None</span>
                    {/if}
                {/snippet}
                {@render fact("Target", targetFact)}

                {#snippet addressFact()}
                    <span class="font-mono text-xs">
                        {log.ip_address ?? "Not recorded"}
                    </span>
                {/snippet}
                {@render fact("IP Address", addressFact)}

                {#snippet applicationFact()}
                    {#if log.application_id}
                        <a
                            href={resolve("/admin/apps/[app]", {
                                app: log.application_id,
                            })}
                            class="font-mono text-xs text-primary hover:underline"
                        >
                            {log.application_id}
                        </a>
                    {:else}
                        <span class="text-muted">None</span>
                    {/if}
                {/snippet}
                {@render fact("Application", applicationFact)}

                {#if log.provider_id}
                    {@const providerId = log.provider_id}
                    {#snippet providerFact()}
                        <a
                            href={resolve("/admin/providers/[provider]", {
                                provider: providerId,
                            })}
                            class="font-mono text-xs text-primary hover:underline"
                        >
                            {providerId}
                        </a>
                    {/snippet}
                    {@render fact("Social Provider", providerFact, true)}
                {/if}
            </dl>

            {#if log.user_agent}
                <div>
                    <h2 class="text-xs text-muted select-none">Browser</h2>
                    <p class="mt-1 font-mono text-xs wrap-anywhere">
                        {log.user_agent}
                    </p>
                </div>
            {/if}

            {#if log.metadata}
                <div>
                    <h2 class="text-xs text-muted select-none">Details</h2>
                    <dl class="mt-1 divide-y divide-background-200">
                        {#each Object.entries(log.metadata) as [key, value] (key)}
                            <div class="flex items-baseline justify-between gap-6 py-1.5">
                                <dt class="font-mono text-xs text-muted select-none">
                                    {key}
                                </dt>
                                <dd class="font-mono text-xs text-right wrap-anywhere">
                                    {displayValue(value)}
                                </dd>
                            </div>
                        {/each}
                    </dl>
                </div>
            {/if}

            <p class="font-mono text-xs text-muted">{log.id}</p>
        </div>
    {/if}
</Modal>
