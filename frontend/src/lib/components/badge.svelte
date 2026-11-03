<script lang="ts" module>
    const tones = {
        neutral: "border-background-200 text-foreground/80",
        primary: "border-primary/40 text-primary",
        success: "border-emerald-400/30 text-emerald-300",
        warning: "border-amber-400/30 text-amber-300",
        danger: "border-red-400/30 text-red-300",
        info: "border-sky-400/30 text-sky-300",
    } as const;

    export type BadgeTone = keyof typeof tones;
</script>

<script lang="ts">
    import { tooltip } from "@/tooltip";
    import { cn } from "@/utils";
    import type { Snippet } from "svelte";
    import type { ClassValue } from "svelte/elements";

    const {
        tone = "neutral",
        title,
        class: className,
        children,
    }: {
        tone?: BadgeTone;
        title?: string;
        class?: ClassValue;
        children?: Snippet;
    } = $props();
</script>

<span
    {@attach tooltip(title)}
    class={cn(
        "inline-flex items-center gap-1 rounded-md border px-1.5 py-0.5 font-mono text-[0.7rem] leading-none whitespace-nowrap select-none",
        tones[tone],
        className,
    )}
>
    {@render children?.()}
</span>
