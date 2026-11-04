<script lang="ts">
    import Icon from "@/components/icon.svelte";
    import { user } from "@/index";
    import type { StarPermission } from "@/star";
    import { cn } from "@/utils";
    import type { Snippet } from "svelte";
    import type { ClassValue } from "svelte/elements";
    import { isActive, navRowClass } from "./nav";

    const {
        href,
        permission,
        icon,
        class: className,
        children,
    }: {
        href: string;
        permission?: StarPermission;
        icon?: string;
        children?: Snippet;
        class?: ClassValue;
    } = $props();
</script>

{#if !permission || ($user && $user.permissions.includes(permission))}
    <a
        data-selected={isActive(href)}
        class={cn(navRowClass, icon && "gap-2", className)}
        {href}
    >
        {#if icon}
            <Icon {icon} class="text-base text-muted" />
        {/if}
        {@render children?.()}
    </a>
{/if}
