<script lang="ts">
    import BackupCodes from "@/components/backup-codes.svelte";
    import Button from "@/components/button.svelte";
    import Card from "@/components/card.svelte";
    import ConfirmModal from "@/components/confirm-modal.svelte";
    import Badge from "@/components/badge.svelte";
    import Icon from "@/components/icon.svelte";
    import Input from "@/components/input.svelte";
    import Modal from "@/components/modal.svelte";
    import { syncPasskeys } from "@/passkey";
    import PasskeyModal from "@/components/modals/mfa/passkey-modal.svelte";
    import TotpModal from "@/components/modals/mfa/totp-modal.svelte";
    import * as star from "@/star";
    import { relativeTime, simpleDate } from "@/utils";
    import { browserSupportsWebAuthn } from "@simplewebauthn/browser";
    import { onMount, type Snippet } from "svelte";
    import { toast } from "svelte-sonner";

    let enableTotpModal = $state<TotpModal>();
    let disableTotpModal = $state<ConfirmModal>();
    let factors = $state<star.MFAMethod[]>([]);
    let totpEnabled = $derived(factors.includes("totp"));

    let addPasskeyModal = $state<PasskeyModal>();
    let renamePasskeyModal = $state<Modal>();
    let deletePasskeyModal = $state<ConfirmModal>();
    let passkeys = $state<star.Passkey[]>([]);
    let selectedPasskey = $state<star.Passkey | undefined>();
    let passkeyName = $state("");
    let renaming = $state(false);

    let regenerateModal = $state<ConfirmModal>();
    let codesModal = $state<Modal>();
    let backupCodes = $state<string[]>([]);
    let remainingCodes = $state(0);

    const refresh = async () => {
        const [data, keys, backup] = await Promise.all([
            star.getMfaFactors(),
            star.getPasskeys(),
            star.getBackupCodeCount(),
        ]);

        if (!data) return toast.error("failed to fetch MFA factors");

        factors = data.factors;
        passkeys = keys;
        remainingCodes = backup?.remaining ?? 0;

        syncPasskeys();
    };

    onMount(refresh);

    const openRename = (passkey: star.Passkey) => {
        selectedPasskey = passkey;
        passkeyName = passkey.name;
        renamePasskeyModal?.open();
    };

    const openDelete = (passkey: star.Passkey) => {
        selectedPasskey = passkey;
        deletePasskeyModal?.open();
    };

    const renamePasskey = async (event: Event) => {
        event.preventDefault();
        if (!selectedPasskey || renaming || !passkeyName.trim()) return;

        renaming = true;
        const response = await star.renamePasskey(
            selectedPasskey.id,
            passkeyName.trim(),
        );
        renaming = false;

        if ("error" in response) {
            toast.error(response.message);
            return;
        }

        passkeys = passkeys.map((passkey) =>
            passkey.id === response.id ? response : passkey,
        );
        renamePasskeyModal?.close();
    };

    const deletePasskey = async () => {
        if (!selectedPasskey) return false;

        const response = await star.deletePasskey(selectedPasskey.id);
        if ("error" in response) {
            toast.error(response.message);
            return false;
        }

        toast.success("Passkey has been removed");
        await refresh();
    };

    const regenerateCodes = async () => {
        const response = await star.regenerateBackupCodes();

        if ("error" in response) {
            toast.error(response.message);
            return false;
        }

        backupCodes = response.backup_codes;
        remainingCodes = backupCodes.length;
        codesModal?.open();
    };

    const disableTotp = async () => {
        const response = await star.deleteTotp();

        if ("error" in response) {
            toast.error(response.message);
            return false;
        }

        toast.success("Authenticator app has been disabled");
        await refresh();
    };
</script>

{#snippet factor(
    icon: string,
    title: string,
    description: string,
    actions: Snippet,
)}
    <div
        class="flex flex-row items-center gap-2 rounded-lg border border-background-200 p-4"
    >
        <Icon {icon} class="text-lg bg-background-200/80 p-2 rounded" />

        <div class="flex flex-col items-baseline leading-none">
            <h3>{title}</h3>
            <p class="text-muted text-sm">{description}</p>
        </div>

        <div class="ml-auto">
            {@render actions()}
        </div>
    </div>
{/snippet}

{#snippet totpActions()}
    {#if totpEnabled}
        <Button variant="danger" onclick={() => disableTotpModal?.open()}>
            Disable
        </Button>
    {:else}
        <Button onclick={enableTotpModal?.open}>Enable</Button>
    {/if}
{/snippet}

{#snippet passkeyActions()}
    <Button
        disabled={!browserSupportsWebAuthn()}
        onclick={() => addPasskeyModal?.open()}
    >
        Add
    </Button>
{/snippet}

{#snippet backupActions()}
    <Button variant="secondary" onclick={() => regenerateModal?.open()}>
        Regenerate
    </Button>
{/snippet}

<Card class="min-w-1/3 w-full md:w-auto">
    <div>
        <h2 class="flex flex-row items-center gap-2 text-sm font-semibold">
            <Icon icon="shield-lock" /> Two-Factor Authentication
        </h2>
        <div class="max-w-xl text-muted">
            Add another step of verification to your login, increasing your
            account security.
        </div>
    </div>

    <div class="my-4 flex flex-col gap-2">
        {@render factor(
            "device-mobile",
            "Authenticator app",
            "Time-based one-time codes (TOTP)",
            totpActions,
        )}

        {@render factor(
            "fingerprint",
            "Passkeys",
            "Sign in with your device instead of a password",
            passkeyActions,
        )}

        {#if passkeys.length}
            <ul
                class="flex flex-col divide-y divide-background-200 rounded-lg border border-background-200"
            >
                {#each passkeys as passkey (passkey.id)}
                    <li class="flex flex-row items-center gap-3 px-4 py-3">
                        <div class="flex min-w-0 flex-col gap-1">
                            <div class="flex flex-row items-center gap-2">
                                <span class="truncate">{passkey.name}</span>
                                <Badge
                                    tone={passkey.synced ? "info" : "neutral"}
                                    class="select-none"
                                >
                                    {passkey.synced ? "Synced" : "This device"}
                                </Badge>
                            </div>
                            <p class="text-xs text-muted select-none">
                                Added {simpleDate(new Date(passkey.created_at))}
                                {#if passkey.last_used_at}
                                    · Used {relativeTime(
                                        passkey.last_used_at,
                                    )} ago
                                {:else}
                                    · Never used
                                {/if}
                            </p>
                        </div>

                        <div class="ml-auto flex flex-row items-center gap-1">
                            <Button
                                variant="icon"
                                aria-label="Rename passkey"
                                onclick={() => openRename(passkey)}
                            >
                                <Icon icon="pencil" />
                            </Button>
                            <Button
                                variant="icon"
                                aria-label="Remove passkey"
                                onclick={() => openDelete(passkey)}
                            >
                                <Icon icon="trash" class="text-red-300" />
                            </Button>
                        </div>
                    </li>
                {/each}
            </ul>
        {/if}

        {#if factors.length}
            {@render factor(
                "key",
                "Backup codes",
                `${remainingCodes} of 10 remaining, each can be used once`,
                backupActions,
            )}
        {/if}
    </div>
</Card>

<TotpModal bind:this={enableTotpModal} onsuccess={refresh} />

<PasskeyModal bind:this={addPasskeyModal} onsuccess={refresh} />

<Modal bind:this={renamePasskeyModal} button>
    <form class="flex flex-col gap-6" onsubmit={renamePasskey}>
        <div class="flex flex-col gap-1 pr-8">
            <h1 class="text-xl font-semibold">Rename passkey</h1>
            <h2 class="text-sm text-muted">
                Pick a name that helps you recognize where this passkey lives.
            </h2>
        </div>

        <Input
            bind:value={passkeyName}
            aria-label="Passkey name"
            maxlength={64}
            autocomplete="off"
            required
        />

        <div class="flex justify-end gap-2">
            <Button
                type="button"
                variant="secondary"
                onclick={() => renamePasskeyModal?.close()}
            >
                Cancel
            </Button>
            <Button
                type="submit"
                loading={renaming}
                disabled={!passkeyName.trim()}
            >
                Save
            </Button>
        </div>
    </form>
</Modal>

<ConfirmModal
    bind:this={deletePasskeyModal}
    title="Remove passkey"
    confirmLabel="Remove"
    onconfirm={deletePasskey}
>
    You will no longer be able to sign in with
    <span class="text-foreground">{selectedPasskey?.name}</span>. Remember to also
    delete it from your device or password manager.
</ConfirmModal>

<ConfirmModal
    bind:this={regenerateModal}
    title="Regenerate backup codes"
    confirmLabel="Regenerate"
    onconfirm={regenerateCodes}
>
    You'll get 10 new codes and every existing backup code will stop working.
</ConfirmModal>

<Modal bind:this={codesModal} button>
    <div class="flex flex-col gap-6">
        <div class="flex flex-col gap-1 pr-8">
            <h1 class="text-xl font-semibold">Your new backup codes</h1>
            <h2 class="text-sm text-muted">
                Use one of these to sign in if you lose access to your other
                sign in methods.
            </h2>
        </div>

        <BackupCodes codes={backupCodes} />

        <div class="flex justify-end">
            <Button onclick={() => codesModal?.close()}>Done</Button>
        </div>
    </div>
</Modal>

<ConfirmModal
    bind:this={disableTotpModal}
    title="Disable authenticator app"
    confirmLabel="Disable"
    onconfirm={disableTotp}
>
    Your account will no longer require a one-time code when signing in, making
    it less secure.
</ConfirmModal>
