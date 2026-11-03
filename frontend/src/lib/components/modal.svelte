<script lang="ts">
    import { cn } from "@/utils";
    import type { Snippet } from "svelte";
    import type { ClassValue } from "svelte/elements";
    import Button from "./button.svelte";
    import Icon from "./icon.svelte";

    let {
        showModal = $bindable(false),
        closable = true,
        button = false,
        class: className,
        children,
    }: {
        showModal?: boolean;
        children?: Snippet;
        closable?: boolean;
        button?: boolean;
        class?: ClassValue;
    } = $props();

    let dialog: HTMLDialogElement = $state()!;
    let panel: HTMLDivElement = $state()!;

    $effect(() => {
        if (showModal && !dialog.open) {
            dialog.showModal();
            panel.focus();
        } else if (!showModal && dialog.open) {
            dialog.close();
        }
    });

    export const close = () => {
        showModal = false;
    };

    export const open = () => {
        showModal = true;
    };

    const handleClick = (e: MouseEvent) => {
        if (e.target != dialog || !closable) return;
        close();
    };

    const handleCancel = (e: Event) => {
        if (closable) return;
        e.preventDefault();
    };
</script>

<!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_noninteractive_element_interactions -->
<dialog
    bind:this={dialog}
    onclose={() => (showModal = false)}
    oncancel={handleCancel}
    onclick={handleClick}
    class={cn(
        "fixed inset-0 m-auto h-fit max-h-[calc(100dvh-2rem)] w-[calc(100%-2rem)] rounded-lg border border-background-200 bg-background-100 p-0 text-foreground backdrop:bg-black/70 sm:w-lg",
        className,
    )}
>
    <div bind:this={panel} tabindex="-1" class="relative p-6 outline-none">
        {#if button}
            <Button
                type="button"
                variant="icon"
                aria-label="Close"
                class="absolute top-4 right-4 text-muted hover:text-foreground"
                onclick={close}
            >
                <Icon icon="x" />
            </Button>
        {/if}

        {@render children?.()}
    </div>
</dialog>

<style>
    @keyframes zoom {
        from {
            scale: 0.95;
        }
        to {
            scale: 1;
        }
    }

    @keyframes fade {
        from {
            opacity: 0;
        }
        to {
            opacity: 1;
        }
    }

    dialog[open] {
        animation: zoom 0.3s cubic-bezier(0.34, 1.56, 0.64, 1);
    }

    dialog[open]::backdrop {
        animation: fade 0.2s ease-out;
    }
</style>
