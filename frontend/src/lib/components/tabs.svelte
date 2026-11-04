<script lang="ts" module>
    export interface TabItem {
        label: string;
        href: string;
        icon?: string;
    }
</script>

<script lang="ts">
    import { page } from "$app/state";
    import { cn } from "@/utils";
    import Icon from "./icon.svelte";

    const { items, label }: { items: TabItem[]; label: string } = $props();
</script>

<nav
    aria-label={label}
    class="flex max-w-full items-center gap-1 overflow-x-auto"
>
    {#each items as item (item.href)}
        {@const current = page.url.pathname === item.href}
        <a
            href={item.href}
            aria-current={current ? "page" : undefined}
            class={cn(
                "flex shrink-0 items-center gap-1.5 rounded-md px-3 py-1.5 text-sm select-none transition-colors",
                current
                    ? "bg-background-200/70 text-foreground"
                    : "text-muted hover:bg-background-200/40 hover:text-foreground",
            )}
        >
            {#if item.icon}
                <Icon icon={item.icon} />
            {/if}
            {item.label}
        </a>
    {/each}
</nav>
