<script lang="ts">
    import { cn } from "@/utils";
    import type { ClassValue, HTMLInputAttributes } from "svelte/elements";
    import Icon from "./icon.svelte";
    import Input from "./input.svelte";

    let {
        value = $bindable(""),
        placeholder = "Search",
        class: className,
        ...rest
    }: {
        value?: string;
        placeholder?: string;
        class?: ClassValue;
    } & Omit<HTMLInputAttributes, "type" | "value"> = $props();
</script>

<div class={cn("relative w-full sm:max-w-xs", className)}>
    <Icon
        icon="search"
        class="pointer-events-none absolute top-1/2 left-3 -translate-y-1/2 text-muted"
    />
    <Input
        bind:value
        type="search"
        {placeholder}
        class="pl-9"
        autocomplete="off"
        {...rest}
    />
</div>
