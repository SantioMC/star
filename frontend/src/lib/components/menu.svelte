<script lang="ts" module>
    export interface MenuItem {
        label: string;
        icon?: string;
        danger?: boolean;
        onclick: () => void;
    }
</script>

<script lang="ts">
    import { closeOnViewportChange, positionPopover } from "@/popover";
    import { cn } from "@/utils";
    import Icon from "./icon.svelte";

    const {
        items,
        label = "More actions",
        icon = "dots",
    }: {
        items: MenuItem[];
        label?: string;
        icon?: string;
    } = $props();

    const uid = $props.id();
    const menuId = `${uid}-menu`;

    let trigger = $state<HTMLButtonElement>();
    let menu = $state<HTMLDivElement>();
    let expanded = $state(false);

    $effect(() => {
        if (!expanded || !menu) return;
        return closeOnViewportChange(menu);
    });

    const entries = () =>
        Array.from(menu?.querySelectorAll<HTMLElement>("[role=menuitem]") ?? []);

    const handleBeforeToggle = (event: ToggleEvent) => {
        if (event.newState !== "open" || !trigger || !menu) return;
        positionPopover(trigger, menu, "end");
    };

    const handleToggle = (event: ToggleEvent) => {
        expanded = event.newState === "open";
        if (!expanded || !trigger || !menu) return;

        positionPopover(trigger, menu, "end");
        entries()[0]?.focus();
    };

    const handleKeydown = (event: KeyboardEvent) => {
        const all = entries();
        const index = all.indexOf(document.activeElement as HTMLElement);

        switch (event.key) {
            case "ArrowDown":
                event.preventDefault();
                all[(index + 1) % all.length]?.focus();
                break;
            case "ArrowUp":
                event.preventDefault();
                all[(index - 1 + all.length) % all.length]?.focus();
                break;
            case "Home":
                event.preventDefault();
                all[0]?.focus();
                break;
            case "End":
                event.preventDefault();
                all[all.length - 1]?.focus();
                break;
            case "Tab":
                menu?.hidePopover();
                break;
        }
    };

    const select = (item: MenuItem) => {
        menu?.hidePopover();
        trigger?.focus();
        item.onclick();
    };
</script>

{#if items.length > 0}
    <button
        bind:this={trigger}
        type="button"
        aria-label={label}
        aria-haspopup="menu"
        aria-expanded={expanded}
        aria-controls={menuId}
        popovertarget={menuId}
        class="rounded-lg p-2 transition-all hover:bg-background-200 active:scale-95 aria-expanded:bg-background-200"
    >
        <Icon {icon} class="text-lg" />
    </button>

    <div
        bind:this={menu}
        id={menuId}
        popover="auto"
        role="menu"
        tabindex="-1"
        onbeforetoggle={handleBeforeToggle}
        ontoggle={handleToggle}
        onkeydown={handleKeydown}
        class="fixed inset-auto m-0 min-w-48 rounded-lg border border-background-200 bg-background-100 p-1 text-sm text-foreground shadow-xl shadow-black/40"
    >
        {#each items as item (item.label)}
            <button
                type="button"
                role="menuitem"
                onclick={() => select(item)}
                class={cn(
                    "flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-left outline-none hover:bg-background-200 focus-visible:bg-background-200",
                    item.danger && "text-red-300 hover:bg-red-400/10 focus-visible:bg-red-400/10",
                )}
            >
                {#if item.icon}
                    <Icon icon={item.icon} />
                {/if}
                {item.label}
            </button>
        {/each}
    </div>
{/if}
