<script lang="ts">
    import { cn } from "@/utils";
    import type { Snippet } from "svelte";
    import type { ClassValue } from "svelte/elements";

    const props: {
        title: string;
        children?: Snippet;
        class?: ClassValue;
    } = $props();
</script>

<div class={cn("flex min-w-0 flex-col gap-1 items-stretch", props.class)}>
    <p class="text-xs font-medium text-muted select-none">
        {props.title}
    </p>
    {@render props.children?.()}
</div>
