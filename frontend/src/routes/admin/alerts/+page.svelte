<script lang="ts">
    import { goto } from "$app/navigation";
    import { resolve } from "$app/paths";
    import * as star from "$lib/star";
    import Badge, { type BadgeTone } from "@/components/badge.svelte";
    import Button from "@/components/button.svelte";
    import Card from "@/components/card.svelte";
    import Icon from "@/components/icon.svelte";
    import PageHeader from "@/components/page-header.svelte";
    import { hasPermission } from "@/index";
    import { tooltip } from "@/tooltip";
    import { formatTime, relativeTime } from "@/utils";
    import { onMount } from "svelte";
    import { toast } from "svelte-sonner";

    const canResolve = hasPermission("admin:alert:resolve");

    const severityTones: Record<star.SecurityAlertSeverity, BadgeTone> = {
        high: "danger",
        medium: "warning",
        low: "info",
    };

    let alerts = $state<star.SecurityAlert[]>([]);
    let status = $state<"open" | "all">("open");
    let page = $state(1);
    let loaded = $state(false);
    let hasMore = $state(false);
    let expanded = $state<string | null>(null);
    let resolving = $state<string | null>(null);

    const load = async (next: number) => {
        const result = await star.getSecurityAlerts(next, status);

        alerts = next === 1 ? result : [...alerts, ...result];
        hasMore = result.length === 25;
        page = next;
        loaded = true;
    };

    onMount(() => {
        if (!hasPermission("admin:alert:read")) return goto("/");
        load(1);
    });

    const switchStatus = (value: "open" | "all") => {
        if (status === value) return;

        status = value;
        loaded = false;
        load(1);
    };

    const resolveAlert = async (alert: star.SecurityAlert) => {
        resolving = alert.id;
        const result = await star.resolveSecurityAlert(alert.id);
        resolving = null;

        if ("error" in result) {
            return toast.error(result.message);
        }

        alerts =
            status === "open"
                ? alerts.filter((it) => it.id !== alert.id)
                : alerts.map((it) => (it.id === alert.id ? result : it));

        toast.success("Marked the alert as resolved");
    };
</script>

<PageHeader
    title="Security Alerts"
    description="Events Star detected that should never happen during normal use and need review"
>
    {#snippet actions()}
        <Button
            variant={status === "open" ? "primary" : "secondary"}
            onclick={() => switchStatus("open")}
        >
            Open
        </Button>
        <Button
            variant={status === "all" ? "primary" : "secondary"}
            onclick={() => switchStatus("all")}
        >
            All
        </Button>
    {/snippet}
</PageHeader>

<div class="flex flex-col gap-3">
    {#each alerts as alert (alert.id)}
        <Card class="flex flex-col gap-3 p-4">
            <div class="flex items-start gap-3">
                <Icon
                    icon={alert.resolved_at ? "shield-check" : "shield-exclamation"}
                    class="mt-0.5 shrink-0 text-lg {alert.resolved_at
                        ? 'text-muted'
                        : 'text-red-300'}"
                />
                <div class="min-w-0 flex-1">
                    <div class="flex flex-wrap items-center gap-2">
                        <Badge tone={severityTones[alert.severity]}>
                            {alert.severity}
                        </Badge>
                        <Badge>{alert.kind}</Badge>
                        {#if alert.resolved_at}
                            <Badge
                                tone="success"
                                title="Resolved {formatTime(
                                    new Date(alert.resolved_at),
                                )}"
                            >
                                resolved
                            </Badge>
                        {/if}
                    </div>
                    <p class="mt-2 text-sm">{alert.message}</p>
                    <div
                        class="mt-2 flex flex-wrap items-center gap-x-4 gap-y-1 text-xs text-muted"
                    >
                        <span
                            {@attach tooltip(
                                formatTime(new Date(alert.created_at)),
                            )}
                        >
                            {relativeTime(alert.created_at)} ago
                        </span>
                        {#if alert.user_id}
                            <a
                                href={resolve("/admin/users/[user]", {
                                    user: alert.user_id,
                                })}
                                class="font-mono hover:text-foreground transition-colors"
                            >
                                {alert.user_id}
                            </a>
                        {/if}
                        {#if alert.application}
                            <a
                                href={resolve("/admin/apps/[app]", {
                                    app: alert.application,
                                })}
                                class="font-mono hover:text-foreground transition-colors"
                            >
                                {alert.application}
                            </a>
                        {/if}
                        {#if alert.data}
                            <button
                                type="button"
                                class="hover:text-foreground transition-colors"
                                onclick={() =>
                                    (expanded =
                                        expanded === alert.id
                                            ? null
                                            : alert.id)}
                            >
                                {expanded === alert.id
                                    ? "Hide details"
                                    : "Show details"}
                            </button>
                        {/if}
                    </div>
                </div>
                {#if canResolve && !alert.resolved_at}
                    <Button
                        variant="secondary"
                        loading={resolving === alert.id}
                        onclick={() => resolveAlert(alert)}
                    >
                        Resolve
                    </Button>
                {/if}
            </div>

            {#if expanded === alert.id && alert.data}
                <pre
                    class="overflow-x-auto rounded-md bg-background-200/60 p-3 font-mono text-xs">{JSON.stringify(
                        alert.data,
                        null,
                        2,
                    )}</pre>
            {/if}
        </Card>
    {:else}
        <Card class="py-10 text-center text-sm text-muted">
            {#if !loaded}
                Loading alerts...
            {:else if status === "open"}
                No open security alerts.
            {:else}
                No security alerts have been raised.
            {/if}
        </Card>
    {/each}

    {#if hasMore}
        <Button variant="secondary" class="self-center" onclick={() => load(page + 1)}>
            Load more
        </Button>
    {/if}
</div>
