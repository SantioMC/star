<script lang="ts">
  import { goto } from "$app/navigation";
  import * as star from "$lib/star";
  import Card from "@/components/card.svelte";
  import PageHeader from "@/components/page-header.svelte";
  import { hasPermission } from "@/index";
  import { relativeTime } from "@/utils";
  import { onMount } from "svelte";

  let metrics = $state<star.Metrics | null>(null);

  const stats = $derived([
    { label: "Users", icon: "users", value: metrics?.users },
    {
      label: "Active sessions",
      icon: "device-desktop",
      value: metrics?.active_sessions,
    },
    { label: "Applications", icon: "apps", value: metrics?.applications },
    {
      label: "Open security alerts",
      icon: "shield-exclamation",
      value: metrics?.open_alerts,
      href: "/admin/alerts",
    },
  ]);

  onMount(async () => {
    if (!hasPermission("admin:metrics:read")) return goto("/");
    metrics = await star.getMetrics();
  });
</script>

<PageHeader
  title="Overview"
  description="Server metrics, this page is ugly and needs changing"
/>

<div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
  {#each stats as stat (stat.label)}
    <Card title={stat.label} icon={stat.icon}>
      <div class="mt-3 flex items-baseline justify-between gap-2">
        <span class="text-3xl font-semibold">
          {stat.value?.toLocaleString() ?? "-"}
        </span>
        {#if stat.href && stat.value && hasPermission("admin:alert:read")}
          <a
            href={stat.href}
            class="text-sm text-primary hover:text-primary/80 transition-colors"
          >
            Review
          </a>
        {/if}
      </div>
    </Card>
  {/each}
</div>

<Card title="Server" icon="server">
  <dl class="mt-3 grid gap-4 sm:grid-cols-2 text-sm">
    <div>
      <dt class="text-xs text-muted select-none">Version</dt>
      <dd class="mt-1 font-mono">{metrics?.version ?? "-"}</dd>
    </div>
    <div>
      <dt class="text-xs text-muted select-none">Uptime</dt>
      <dd class="mt-1">
        {metrics ? relativeTime(metrics.started_at) : "-"}
      </dd>
    </div>
  </dl>
</Card>
