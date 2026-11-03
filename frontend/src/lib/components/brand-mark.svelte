<script lang="ts">
    import { branding } from "@/branding";
    import { cn } from "@/utils";
    import type { ClassValue } from "svelte/elements";
    import Icon from "./icon.svelte";

    const {
        name = true,
        class: className,
    }: {
        name?: boolean;
        class?: ClassValue;
    } = $props();
</script>

<div class={cn("flex items-center gap-2.5 select-none", className)}>
    {#if $branding.logo_url}
        <img
            src={$branding.logo_url}
            alt={name ? "" : $branding.name}
            class="size-8 shrink-0 rounded-lg object-cover"
        />
    {:else}
        <span
            class="flex size-8 shrink-0 items-center justify-center rounded-lg bg-primary text-primary-foreground"
            aria-hidden={name}
        >
            <Icon icon="star" filled />
        </span>
    {/if}

    {#if name}
        <span class="truncate font-serif font-semibold tracking-tight">
            {$branding.name}
        </span>
    {/if}
</div>
