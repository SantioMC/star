<script lang="ts">
    import { cn } from "@/utils";
    import type { ClassValue } from "svelte/elements";

    let {
        length = 6,
        value = $bindable(""),
        oncomplete,
        onenter,
        type = "number",
        class: className,
    }: {
        length?: number;
        value?: string;
        type?: "text" | "number";
        oncomplete?: (value: string) => void;
        onenter?: (value: string) => void;
        class?: ClassValue;
    } = $props();

    // svelte-ignore state_referenced_locally
    let parts = $state<string[]>(new Array(length).fill(""));
    let container = $state<HTMLDivElement>();

    $effect(() => {
        value = parts.join("");
    });

    const handleKeydown = (event: KeyboardEvent, index: number) => {
        if (event.key === "Backspace" && !parts[index] && index > 0) {
            const previous = container?.children.item(
                index - 1,
            ) as HTMLInputElement | null;

            if (!previous) return;

            previous.focus();
            previous.setSelectionRange(
                previous.value.length,
                previous.value.length,
            );
        } else if (event.key === "Enter") {
            if (oncomplete) oncomplete?.(value);
            else onenter?.(value);
        }
    };

    const handleInput = (event: Event, index: number) => {
        const input = event.currentTarget as HTMLInputElement;

        if (type === "number") {
            input.value = input.value.replace(/\D/g, "");
        }

        parts[index] = input.value;
        if (input.value && index < length - 1) {
            const next = container?.children.item(
                index + 1,
            ) as HTMLInputElement | null;

            next?.focus();
        }

        if (parts.every(Boolean)) {
            oncomplete?.(parts.join(""));
        }
    };

    const handleFocus = (event: FocusEvent) => {
        const input = event.target as HTMLInputElement;

        const index = Array.from(container?.children ?? []).indexOf(input);
        if (index === -1) return;

        const firstEmpty = parts.findIndex((part) => !part);
        if (firstEmpty !== -1 && index > firstEmpty) {
            const target = container?.children.item(
                firstEmpty,
            ) as HTMLInputElement | null;

            target?.focus();
        }
    };

    const handlePaste = (event: ClipboardEvent) => {
        event.preventDefault();

        const data = event.clipboardData;
        if (!data) return;

        let content = data.getData("text").toUpperCase();
        if (type === "number") {
            content = content.replace(/\D/g, "");
        }

        let element: HTMLInputElement | null = null;
        for (let i = 0; i < Math.min(content.length, length); i++) {
            element = container?.children.item(i) as HTMLInputElement | null;
            if (!element) continue;

            element.value = content[i];
            parts[i] = content[i];
        }

        if (element) {
            element.focus();
            element.setSelectionRange(
                element.value.length,
                element.value.length,
            );
        }
    };
</script>

<div
    class={cn("flex flex-row gap-2 items-center justify-center", className)}
    bind:this={container}
>
    {#each Array(length) as _, index}
        <input
            type="text"
            inputmode="numeric"
            maxlength="1"
            onpaste={handlePaste}
            onfocusin={handleFocus}
            oninput={(event) => handleInput(event, index)}
            onkeydown={(event) => handleKeydown(event, index)}
            autocomplete={index === 0 ? "one-time-code" : "off"}
            class="p-0 bg-background size-11 sm:size-12 text-lg outline-none input-clean text-center rounded-lg border border-background-200 transition-[color,box-shadow] focus-visible:border-foreground/80 focus-visible:ring-1 focus-visible:ring-foreground/50"
        />
    {/each}
</div>
