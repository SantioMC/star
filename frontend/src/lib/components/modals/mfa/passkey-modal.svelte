<script lang="ts">
    import BackupCodes from "@/components/backup-codes.svelte";
    import Button from "@/components/button.svelte";
    import Input from "@/components/input.svelte";
    import Modal from "@/components/modal.svelte";
    import { Label } from "@/components/ui/label";
    import { registerPasskey } from "@/passkey";
    import { toast } from "svelte-sonner";

    const {
        onsuccess,
    }: {
        onsuccess?: () => void;
    } = $props();

    let modal = $state<Modal>();
    let name = $state("");
    let creating = $state(false);
    let error = $state("");
    let backupCodes = $state<string[]>([]);

    export const open = () => {
        name = "";
        error = "";
        backupCodes = [];

        modal?.open();
    };

    export const close = () => modal?.close();

    const create = async (event: Event) => {
        event.preventDefault();
        if (creating || !name.trim()) return;

        error = "";
        creating = true;
        const response = await registerPasskey(name.trim());
        creating = false;

        if (response === null) return;
        if ("error" in response) {
            error = response.message;
            return;
        }

        toast.success("Passkey added");
        onsuccess?.();

        if (!response.backup_codes?.length) return close();
        backupCodes = response.backup_codes;
    };
</script>

<Modal bind:this={modal} button class="sm:w-lg">
    <div class="flex flex-col gap-6">
        <div class="flex flex-col gap-1 pr-8">
            <h1 class="text-xl font-semibold">
                {backupCodes.length ? "Save your backup codes" : "Add a passkey"}
            </h1>
            <h2 class="text-sm text-muted">
                {backupCodes.length
                    ? "If you lose access to your passkeys, you can sign in with one of these instead."
                    : "Sign in with your fingerprint, face, screen lock or security key instead of a password."}
            </h2>
        </div>

        {#if backupCodes.length}
            <BackupCodes codes={backupCodes} />

            <div class="flex justify-end">
                <Button onclick={close}>Done</Button>
            </div>
        {:else}
            <form class="flex flex-col gap-6" onsubmit={create}>
                <div class="grid gap-2">
                    <Label for="passkey-name" class="select-none">Name</Label>
                    <Input
                        bind:value={name}
                        id="passkey-name"
                        placeholder="Work laptop"
                        maxlength={64}
                        autocomplete="off"
                        required
                    />
                </div>

                {#if error}
                    <p role="alert" class="text-sm text-red-300">{error}</p>
                {/if}

                <div class="flex justify-end gap-2">
                    <Button type="button" variant="secondary" onclick={close}>
                        Cancel
                    </Button>
                    <Button
                        type="submit"
                        loading={creating}
                        disabled={!name.trim()}
                    >
                        Continue
                    </Button>
                </div>
            </form>
        {/if}
    </div>
</Modal>
