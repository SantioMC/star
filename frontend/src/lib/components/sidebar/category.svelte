<script lang="ts">
    import { cn } from "@/utils";
    import type { Snippet } from "svelte";
    import type { ClassValue } from "svelte/elements";

    const {
        class: className,
        children,
    }: { children?: Snippet; class?: ClassValue } = $props();
</script>

<p
    class={cn(
        "font-mono text-xs text-muted uppercase select-none px-1 my-2",
        className,
    )}
>
    {@render children?.()}
</p>
