<script lang="ts">
    import type { Snippet } from "svelte";
    import Checkbox from "./checkbox.svelte";

    let {
        checked = $bindable(false),
        title,
        children,
    }: {
        checked?: boolean;
        title: string;
        children?: Snippet;
    } = $props();
</script>

<label
    class="flex items-start gap-3 rounded-lg border border-background-200 p-3 transition-colors has-checked:border-primary/40 has-enabled:cursor-pointer has-enabled:hover:bg-hover"
>
    <Checkbox class="mt-0.5" bind:checked />
    <div class="flex flex-col gap-0.5">
        <span class="text-sm">{title}</span>
        {#if children}
            <span class="text-sm text-muted">{@render children()}</span>
        {/if}
    </div>
</label>
