<script lang="ts" module>
    const tones = {
        warning: {
            container: "border-amber-400/30 bg-amber-400/5",
            icon: "text-amber-300",
            title: "text-amber-200",
        },
        info: {
            container: "border-sky-400/30 bg-sky-400/5",
            icon: "text-sky-300",
            title: "text-sky-200",
        },
        danger: {
            container: "border-red-400/30 bg-red-400/5",
            icon: "text-red-300",
            title: "text-red-200",
        },
    } as const;

    export type NoticeTone = keyof typeof tones;
</script>

<script lang="ts">
    import { cn } from "@/utils";
    import type { Snippet } from "svelte";
    import type { ClassValue } from "svelte/elements";
    import Icon from "./icon.svelte";

    const {
        tone = "warning",
        icon,
        title,
        class: className,
        children,
        actions,
    }: {
        tone?: NoticeTone;
        icon: string;
        title?: string;
        class?: ClassValue;
        children?: Snippet;
        actions?: Snippet;
    } = $props();

    const styles = $derived(tones[tone]);
</script>

<div
    class={cn(
        "flex items-start gap-3 rounded-lg border px-4 py-3 text-sm leading-5",
        styles.container,
        className,
    )}
>
    <span class="flex h-5 shrink-0 items-center">
        <Icon {icon} class={cn("leading-none", styles.icon)} />
    </span>
    <p>
        {#if title}
            <span class={cn("font-medium", styles.title)}>{title}</span>
        {/if}
        <span class="text-foreground/80">{@render children?.()}</span>
    </p>
    {#if actions}
        <div class="ml-auto flex shrink-0 items-center gap-2 self-center">
            {@render actions()}
        </div>
    {/if}
</div>
