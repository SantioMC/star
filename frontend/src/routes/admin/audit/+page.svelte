<script lang="ts">
    import { goto } from "$app/navigation";
    import * as star from "$lib/star";
    import AuditLogSettingsModal from "@/components/audit-log-settings-modal.svelte";
    import AuditLogView from "@/components/audit-log-view.svelte";
    import Button from "@/components/button.svelte";
    import Icon from "@/components/icon.svelte";
    import PageHeader from "@/components/page-header.svelte";
    import { hasPermission } from "@/index";
    import { onMount } from "svelte";

    const canWrite = hasPermission("admin:audit:write");

    let settings = $state<star.AuditSettings | null>(null);
    let settingsModal = $state<AuditLogSettingsModal>();

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

<PageHeader title="Audit Log" {description}>
    {#snippet actions()}
        {#if canWrite && settings}
            <Button variant="secondary" onclick={() => settingsModal?.open()}>
                <Icon icon="settings" />
                Retention & privacy
            </Button>
        {/if}
    {/snippet}
</PageHeader>

<AuditLogView />

{#if settings}
    <AuditLogSettingsModal bind:this={settingsModal} bind:settings />
{/if}
