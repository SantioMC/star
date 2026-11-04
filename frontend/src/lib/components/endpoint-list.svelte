<script lang="ts" module>
    export interface Endpoint {
        label: string;
        value: string;
    }
</script>

<script lang="ts">
    import Copy from "./copy.svelte";

    const { endpoints }: { endpoints: Endpoint[] } = $props();
</script>

<dl class="flex flex-col gap-4">
    {#each endpoints as endpoint (endpoint.label)}
        <div class="min-w-0">
            <dt class="text-xs text-muted select-none">{endpoint.label}</dt>
            <dd class="mt-1 flex items-center gap-2 font-mono text-sm break-all">
                {endpoint.value}
                <Copy content={endpoint.value} class="text-muted" />
            </dd>
        </div>
    {/each}
</dl>
