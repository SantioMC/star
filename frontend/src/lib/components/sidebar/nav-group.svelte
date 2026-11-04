<script lang="ts">
    import Icon from "@/components/icon.svelte";
    import { user } from "@/index";
    import type { StarPermission } from "@/star";
    import { cn } from "@/utils";
    import { isActive, navRowClass } from "./nav";
    import NavItem from "./nav-item.svelte";

    const {
        label,
        icon,
        items,
    }: {
        label: string;
        icon: string;
        items: {
            href: string;
            label: string;
            permission?: StarPermission;
        }[];
    } = $props();

    const id = $props.id();

    const visibleItems = $derived(
        items.filter(
            (item) =>
                !item.permission ||
                $user?.permissions.includes(item.permission),
        ),
    );

    const containsActive = $derived(
        visibleItems.some((item) => isActive(item.href)),
    );

    let open = $derived(containsActive);
</script>

{#if visibleItems.length > 0}
    <button
        type="button"
        aria-expanded={open}
        aria-controls={id}
        data-selected={containsActive && !open}
        class={cn(navRowClass, "w-full gap-2 select-none")}
        onclick={() => (open = !open)}
    >
        <Icon {icon} class="text-base text-muted" />
        {label}
        <Icon
            icon="chevron-down"
            class={cn(
                "ml-auto text-sm text-muted transition-transform motion-reduce:transition-none",
                open && "rotate-180",
            )}
        />
    </button>

    <div
        {id}
        inert={!open}
        class={cn(
            "grid grid-rows-[0fr] transition-[grid-template-rows] duration-200 motion-reduce:transition-none",
            open && "grid-rows-[1fr]",
        )}
    >
        <div class="min-h-0 overflow-hidden">
            <div
                class="ml-4.5 my-0.5 flex flex-col gap-0.5 border-l border-background-200"
            >
                {#each visibleItems as item (item.href)}
                    <NavItem href={item.href} class="-ml-px pl-3">
                        {item.label}
                    </NavItem>
                {/each}
            </div>
        </div>
    </div>
{/if}
