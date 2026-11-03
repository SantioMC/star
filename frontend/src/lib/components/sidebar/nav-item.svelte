<script lang="ts">
    import { page } from "$app/state";
    import { user } from "@/index";
    import type { StarPermission } from "@/star";
    import { cn } from "@/utils";
    import type { Snippet } from "svelte";
    import type { ClassValue } from "svelte/elements";

    const isActive = (href: string): boolean => {
        if (href.includes("#")) {
            return page.url.pathname === href.split("#")[0];
        }

        if (href === "/") return page.url.pathname === "/";
        return page.url.pathname.startsWith(href);
    };

    const {
        href,
        permission,
        class: className,
        children,
    }: {
        href: string;
        permission?: StarPermission;
        children?: Snippet;
        class?: ClassValue;
    } = $props();
</script>

{#if !permission || ($user && $user.permissions.includes(permission))}
    <a
        data-selected={isActive(href)}
        class={cn(
            "flex items-center gap-1 px-2.5 py-1.5 data-[selected='true']:font-medium text-foreground/90 data-[selected='true']:text-foreground transition-colors hover:bg-hover data-[selected='true']:bg-hover text-sm rounded-r-lg border-transparent data-[selected='true']:border-primary border-l cursor-pointer",
            className,
        )}
        {href}
    >
        {@render children?.()}
    </a>
{/if}
