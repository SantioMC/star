<script lang="ts">
    import { cn } from "@/utils";
    import type { ClassValue, HTMLInputAttributes } from "svelte/elements";
    import Icon from "./icon.svelte";

    let {
        checked = $bindable(false),
        class: className,
        ...rest
    }: {
        checked?: boolean;
        class?: ClassValue;
    } & Omit<HTMLInputAttributes, "type"> = $props();
</script>

<span class={cn("relative inline-flex size-4 shrink-0", className)}>
    <input
        type="checkbox"
        bind:checked
        class="peer size-4 appearance-none rounded border border-background-200 bg-background transition-colors not-disabled:cursor-pointer checked:border-primary checked:bg-primary focus-visible:ring-1 focus-visible:ring-foreground/50 focus-visible:outline-none disabled:opacity-50"
        {...rest}
    />
    <Icon
        icon="check"
        class="pointer-events-none absolute inset-0 hidden items-center justify-center text-xs text-[#111] peer-checked:flex"
    />
</span>
