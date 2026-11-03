<script lang="ts" module>
    export interface SelectOption<T extends string = string> {
        value: T;
        label: string;
        hint?: string;
    }
</script>

<script lang="ts" generics="T extends string">
    import { closeOnViewportChange, positionPopover } from "@/popover";
    import { cn } from "@/utils";
    import type { ClassValue } from "svelte/elements";
    import Icon from "./icon.svelte";

    let {
        value = $bindable(),
        options,
        placeholder = "Select an option",
        id,
        disabled = false,
        onchange,
        class: className,
    }: {
        value?: T;
        options: SelectOption<T>[];
        placeholder?: string;
        id?: string;
        disabled?: boolean;
        onchange?: (value: T) => void;
        class?: ClassValue;
    } = $props();

    const uid = $props.id();
    const listId = `${uid}-list`;

    let trigger = $state<HTMLButtonElement>();
    let list = $state<HTMLDivElement>();
    let expanded = $state(false);
    let active = $state(-1);

    const selected = $derived(options.find((it) => it.value === value));

    $effect(() => {
        if (!expanded || !list) return;
        return closeOnViewportChange(list);
    });

    const pick = (option: SelectOption<T>) => {
        value = option.value;
        onchange?.(option.value);
        list?.hidePopover();
        trigger?.focus();
    };

    const handleBeforeToggle = (event: ToggleEvent) => {
        if (event.newState !== "open" || !trigger || !list) return;

        const index = options.findIndex((it) => it.value === value);
        active = Math.max(index, 0);
        positionPopover(trigger, list, "start", true);
    };

    const handleToggle = (event: ToggleEvent) => {
        expanded = event.newState === "open";
        if (!expanded || !trigger || !list) return;

        positionPopover(trigger, list, "start", true);
        list.querySelector<HTMLElement>(`[data-index="${active}"]`)?.scrollIntoView({
            block: "nearest",
        });
    };

    const move = (index: number) => {
        const count = options.length;
        if (count === 0) return;

        active = (index + count) % count;
        list?.querySelector<HTMLElement>(`[data-index="${active}"]`)?.scrollIntoView({
            block: "nearest",
        });
    };

    const handleKeydown = (event: KeyboardEvent) => {
        if (disabled) return;

        if (!expanded) {
            const opens = ["ArrowDown", "ArrowUp", "Enter", " "];
            if (!opens.includes(event.key)) return;

            event.preventDefault();
            list?.showPopover();
            return;
        }

        switch (event.key) {
            case "ArrowDown":
                event.preventDefault();
                move(active + 1);
                break;
            case "ArrowUp":
                event.preventDefault();
                move(active - 1);
                break;
            case "Home":
                event.preventDefault();
                move(0);
                break;
            case "End":
                event.preventDefault();
                move(options.length - 1);
                break;
            case "Enter":
            case " ": {
                event.preventDefault();
                const option = options[active];
                if (option) pick(option);
                break;
            }
            case "Escape":
                event.preventDefault();
                list?.hidePopover();
                break;
            case "Tab":
                list?.hidePopover();
                break;
        }
    };

    const handleOptionClick = (event: MouseEvent, option: SelectOption<T>) => {
        event.preventDefault();
        pick(option);
    };
</script>

<button
    bind:this={trigger}
    {id}
    {disabled}
    type="button"
    role="combobox"
    aria-controls={listId}
    aria-expanded={expanded}
    aria-haspopup="listbox"
    aria-activedescendant={expanded && active >= 0
        ? `${uid}-option-${active}`
        : undefined}
    popovertarget={listId}
    onkeydown={handleKeydown}
    class={cn(
        "border-background-200 bg-background flex h-9 w-full min-w-0 items-center justify-between gap-2 rounded-md border px-3 py-1 text-left text-base transition-[color,box-shadow] outline-none md:text-sm",
        "focus-visible:border-foreground/80 focus-visible:ring-foreground/50 focus-visible:ring-1",
        "disabled:cursor-not-allowed disabled:opacity-50",
        className,
    )}
>
    <span class={cn("truncate", !selected && "text-muted")}>
        {selected?.label ?? placeholder}
    </span>
    <Icon
        icon="chevron-down"
        class="text-muted transition-transform {expanded ? 'rotate-180' : ''}"
    />
</button>

<div
    bind:this={list}
    id={listId}
    popover="auto"
    role="listbox"
    tabindex="-1"
    onbeforetoggle={handleBeforeToggle}
    ontoggle={handleToggle}
    class="fixed inset-auto m-0 max-h-64 overflow-y-auto overscroll-contain rounded-lg border border-background-200 bg-background-100 p-1 text-sm text-foreground shadow-xl shadow-black/40"
>
    {#each options as option, index (option.value)}
        <button
            id="{uid}-option-{index}"
            data-index={index}
            type="button"
            role="option"
            tabindex="-1"
            aria-selected={option.value === value}
            onmousemove={() => (active = index)}
            onclick={(event) => handleOptionClick(event, option)}
            class={cn(
                "flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-left",
                index === active && "bg-background-200",
            )}
        >
            <Icon
                icon="check"
                class={cn(
                    "text-primary",
                    option.value !== value && "invisible",
                )}
            />
            <span class="truncate">{option.label}</span>
            {#if option.hint}
                <span class="ml-auto pl-4 font-mono text-xs text-muted select-none">
                    {option.hint}
                </span>
            {/if}
        </button>
    {/each}
</div>
