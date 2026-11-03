<script lang="ts">
    import type { Snippet } from "svelte";
    import Button from "./button.svelte";
    import Input from "./input.svelte";
    import Modal from "./modal.svelte";

    const {
        title,
        confirmLabel = "Confirm",
        variant = "danger",
        confirmText,
        onconfirm,
        children,
    }: {
        title: string;
        confirmLabel?: string;
        variant?: "danger" | "primary";
        confirmText?: string;
        onconfirm: () => Promise<boolean | void> | boolean | void;
        children?: Snippet;
    } = $props();

    let modal = $state<Modal>();
    let loading = $state(false);
    let typed = $state("");

    const matches = $derived(!confirmText || typed === confirmText);

    export const open = () => {
        typed = "";
        modal?.open();
    };

    export const close = () => modal?.close();

    const confirm = async (event: Event) => {
        event.preventDefault();
        if (!matches || loading) return;

        loading = true;
        const result = await onconfirm();
        loading = false;

        if (result !== false) close();
    };
</script>

<Modal bind:this={modal} button>
    <form class="flex flex-col gap-6" onsubmit={confirm}>
        <div class="flex flex-col gap-1 pr-8">
            <h1 class="text-xl font-semibold">{title}</h1>
            <div class="text-sm text-muted">
                {@render children?.()}
            </div>
        </div>

        {#if confirmText}
            <label class="grid gap-2 text-sm">
                <span class="select-none">
                    Type <span class="font-mono text-foreground"
                        >{confirmText}</span
                    > to confirm
                </span>
                <Input bind:value={typed} autocomplete="off" />
            </label>
        {/if}

        <div class="flex items-center gap-2 justify-end">
            <Button type="button" variant="secondary" onclick={close}>
                Cancel
            </Button>
            <Button type="submit" {variant} {loading} disabled={!matches}>
                {confirmLabel}
            </Button>
        </div>
    </form>
</Modal>
