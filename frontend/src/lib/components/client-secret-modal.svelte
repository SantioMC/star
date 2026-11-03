<script lang="ts">
    import { branding } from "@/branding";
    import Button from "./button.svelte";
    import Hidden from "./hidden.svelte";
    import Modal from "./modal.svelte";
    import Notice from "./notice.svelte";

    let modal = $state<Modal>();
    let secret = $state("");

    export const show = (value: string) => {
        secret = value;
        modal?.open();
    };

    const close = () => {
        secret = "";
        modal?.close();
    };
</script>

<Modal bind:this={modal} button>
    <div class="flex flex-col gap-6">
        <div class="flex flex-col gap-1 pr-8">
            <h1 class="text-xl font-semibold">Save your client secret</h1>
            <h2 class="text-sm text-muted">
                {$branding.name} only stores a hash of this secret, so it can not be shown
                again. Rotate the secret if it is ever lost.
            </h2>
        </div>

        <Notice icon="alert-triangle" title="Copy it now.">
            Treat this value like a password and keep it out of client-side
            code.
        </Notice>

        <Hidden content={secret} />

        <div class="flex items-center justify-end">
            <Button onclick={close}>Done</Button>
        </div>
    </div>
</Modal>
