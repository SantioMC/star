<script lang="ts">
    import { goto } from "$app/navigation";
    import * as star from "$lib/star";
    import AuditLogView from "@/components/audit-log-view.svelte";
    import PageHeader from "@/components/page-header.svelte";
    import { hasPermission } from "@/index";
    import { onMount } from "svelte";

    let settings = $state<star.AuditSettings | null>(null);

    const description = $derived.by(() => {
        const days = settings?.retention_days;
        const kept = days ? `kept for ${days} days` : "kept forever";

        return `Every sign in, security change and administrative action, ${kept}`;
    });

    onMount(async () => {
        if (!hasPermission("admin:audit:read")) return goto("/");
        settings = await star.getAuditSettings();
    });
</script>

<PageHeader title="Audit Log" {description} />

<AuditLogView />
