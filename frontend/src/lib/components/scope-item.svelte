<script lang="ts">
    import Icon from "@/components/icon.svelte";
    import type { Snippet } from "svelte";

    const {
        name,
        description,
        badges,
        actions,
    }: {
        name: string;
        description: string;
        badges?: Snippet;
        actions?: Snippet;
    } = $props();
</script>

<div
    class="flex items-center gap-3 px-1 py-2.5"
>
    <Icon class="shrink-0 text-lg text-muted" icon="lock-access" />
    <div class="min-w-0">
        <p class="flex flex-wrap items-center gap-2">
            <code class="truncate font-mono text-sm">{name}</code>
            {@render badges?.()}
        </p>
        <p class="text-xs text-muted mt-0.5">{description}</p>
    </div>

    {#if actions}
        <div class="ml-auto flex shrink-0 items-center gap-1">
            {@render actions()}
        </div>
    {/if}
</div>
