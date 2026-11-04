<script lang="ts">
    import { cn } from "@/utils";
    import type { ClassValue, HTMLInputAttributes } from "svelte/elements";

    let {
        value = $bindable(),
        class: className,
        ...rest
    }: {
        class?: ClassValue;
        value?: string;
    } & HTMLInputAttributes = $props();
</script>

<input
    class={cn(
        "border-background-200 bg-background dark:bg-input/30 selection:bg-primary selection:text-primary-foreground ring-offset-background placeholder:text-muted-foreground flex h-9 w-full min-w-0 rounded-md border px-3 py-1 text-base shadow-xs transition-[color,box-shadow] outline-none disabled:cursor-not-allowed disabled:opacity-50 md:text-sm",
        "focus-visible:border-foreground/80 focus-visible:ring-foreground/50 focus-visible:ring-1",
        "aria-invalid:ring-destructive/20 dark:aria-invalid:ring-destructive/40 aria-invalid:border-destructive",
        className,
    )}
    {...rest}
    bind:value
/>
