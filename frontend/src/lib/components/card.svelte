<script lang="ts">
    import { cn } from "@/utils";
    import type { Snippet } from "svelte";
    import type { ClassValue } from "svelte/elements";
    import Icon from "./icon.svelte";

    const {
        children,
        title,
        icon,
        description,
        actions,
        class: className,
    }: {
        children?: Snippet;
        title?: string;
        icon?: string;
        description?: string;
        actions?: Snippet;
        class?: ClassValue;
    } = $props();
</script>

<div
    class={cn(
        "bg-background-100 border border-background-200 p-6 rounded-lg",
        className,
    )}
>
    {#if title}
        <div class="flex flex-row items-start justify-between gap-4">
            <div class="min-w-0">
                <h2
                    class="flex flex-row items-center gap-2 text-sm font-semibold"
                >
                    {#if icon}
                        <Icon {icon} />
                    {/if}
                    {title}
                </h2>
                {#if description}
                    <p class="max-w-xl text-muted text-sm mt-0.5">
                        {description}
                    </p>
                {/if}
            </div>

            {#if actions}
                <div class="flex items-center gap-2 shrink-0">
                    {@render actions()}
                </div>
            {/if}
        </div>
    {/if}

    {@render children?.()}
</div>
