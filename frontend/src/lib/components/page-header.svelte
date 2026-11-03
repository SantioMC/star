<script lang="ts">
    import type { Snippet } from "svelte";

    const {
        title,
        description,
        actions,
    }: {
        title: string;
        description?: string;
        actions?: Snippet;
    } = $props();
</script>

<div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
    <div class="min-w-0">
        <h1 class="leading-8 text-2xl font-semibold">{title}</h1>
        {#if description}
            <h2 class="leading-6 text-muted text-sm mt-1">{description}</h2>
        {/if}
    </div>

    {#if actions}
        <div class="flex items-center gap-2 flex-row shrink-0">
            {@render actions()}
        </div>
    {/if}
</div>
