<script lang="ts">
    import { tooltip } from "@/tooltip";
    import ConfirmModal from "@/components/confirm-modal.svelte";
    import Copy from "@/components/copy.svelte";
    import Icon from "@/components/icon.svelte";

    let {
        redirectUri,
        onremove,
    }: {
        redirectUri: string;
        onremove?: () => Promise<boolean | void> | boolean | void;
    } = $props();

    let removeModal = $state<ConfirmModal>();
</script>

<div
    class="flex min-w-0 items-center gap-2 px-1 py-2.5"
>
    <Icon icon="world" class="shrink-0 text-muted" />
    <code class="truncate font-mono text-sm" {@attach tooltip(redirectUri)}>
        {redirectUri}
    </code>
    <div class="ml-auto flex shrink-0 items-center gap-1">
        <Copy content={redirectUri} />
        {#if onremove}
            <button
                type="button"
                aria-label="Remove {redirectUri}"
                {@attach tooltip("Remove redirect URI")}
                onclick={() => removeModal?.open()}
                class="inline-flex items-center justify-center rounded p-0.5 text-muted transition-colors hover:text-red-300"
            >
                <Icon icon="trash" />
            </button>
        {/if}
    </div>
</div>

{#if onremove}
    <ConfirmModal
        bind:this={removeModal}
        title="Remove redirect URI"
        confirmLabel="Remove"
        onconfirm={() => onremove?.()}
    >
        <span class="font-mono text-foreground break-all">{redirectUri}</span>
        will no longer be accepted during authorization. Logins redirecting to it
        will fail.
    </ConfirmModal>
{/if}
