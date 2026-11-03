<script lang="ts">
    import { tooltip } from "@/tooltip";
    import Copy from "@/components/copy.svelte";
    import Icon from "@/components/icon.svelte";
    import { cn } from "@/utils";
    import type { ClassValue } from "svelte/elements";

    let hidden = $state(true);

    const props: {
        content: string;
        class?: ClassValue;
    } = $props();
</script>

<div class={cn("flex min-w-0 flex-row items-center gap-1.5", props.class)}>
    <code class="grid min-w-0 font-mono text-sm">
        <span class={cn("col-start-1 row-start-1 truncate", !hidden && "invisible")}>
            {"•".repeat(Math.min(props.content.length, 32))}
        </span>
        <span class={cn("col-start-1 row-start-1 truncate", hidden && "invisible")}>
            {props.content}
        </span>
    </code>
    <button
        type="button"
        aria-label={hidden ? "Reveal value" : "Hide value"}
        {@attach tooltip(hidden ? "Reveal value" : "Hide value")}
        onclick={() => (hidden = !hidden)}
        class="inline-flex shrink-0 items-center justify-center rounded p-0.5 text-muted transition-colors hover:text-foreground"
    >
        <Icon icon={hidden ? "eye" : "eye-off"} />
    </button>
    <Copy content={props.content} />
</div>
