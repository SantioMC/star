<script lang="ts">
    import { goto } from "$app/navigation";
    import * as star from "$lib/star";
    import Badge, { type BadgeTone } from "@/components/badge.svelte";
    import Button from "@/components/button.svelte";
    import Card from "@/components/card.svelte";
    import ConfirmModal from "@/components/confirm-modal.svelte";
    import Icon from "@/components/icon.svelte";
    import PageHeader from "@/components/page-header.svelte";
    import { hasPermission } from "@/index";
    import { tooltip } from "@/tooltip";
    import { formatTime, relativeTime } from "@/utils";
    import { onMount } from "svelte";
    import { toast } from "svelte-sonner";

    const canRotate = hasPermission("admin:key:rotate");
    const canDelete = hasPermission("admin:key:delete");

    const statusTones: Record<star.SigningKey["status"], BadgeTone> = {
        active: "success",
        next: "info",
        retiring: "warning",
    };

    let keys = $state<star.SigningKey[]>([]);
    let loaded = $state(false);
    let deleteTarget = $state<star.SigningKey | null>(null);

    let rotateModal = $state<ConfirmModal>();
    let deleteModal = $state<ConfirmModal>();

    onMount(async () => {
        if (!hasPermission("admin:key:read")) return goto("/");

        keys = await star.getSigningKeys();
        loaded = true;
    });

    const rotateKeys = async () => {
        const result = await star.rotateSigningKeys();
        if ("error" in result) {
            toast.error(result.message);
            return false;
        }

        keys = result;
        toast.success("Rotated the signing key");
    };

    const confirmDelete = (key: star.SigningKey) => {
        deleteTarget = key;
        deleteModal?.open();
    };

    const deleteKey = async () => {
        if (!deleteTarget) return false;

        const result = await star.deleteSigningKey(deleteTarget.id);
        if ("error" in result) {
            toast.error(result.message);
            return false;
        }

        keys = result;
        deleteTarget = null;
        toast.success("Deleted the signing key");
    };
</script>

<PageHeader
    title="Signing Keys"
    description="Keys used to sign access and ID tokens, published through the JWKS endpoint"
>
    {#snippet actions()}
        {#if canRotate}
            <Button onclick={() => rotateModal?.open()}>
                <Icon icon="refresh" />
                Rotate keys
            </Button>
        {/if}
    {/snippet}
</PageHeader>

<Card class="p-2 sm:p-3">
    <table class="w-full text-sm">
        <thead>
            <tr class="hover:bg-transparent">
                <th class="px-3 text-xs font-medium text-muted select-none">
                    Key
                </th>
                <th
                    class="px-3 text-xs font-medium text-muted select-none max-sm:hidden"
                >
                    Expires
                </th>
                <th
                    class="px-3 text-xs font-medium text-muted select-none text-right max-sm:hidden"
                >
                    Created
                </th>
                {#if canDelete}
                    <th class="w-0"></th>
                {/if}
            </tr>
        </thead>
        <tbody>
            {#each keys as key (key.id)}
                <tr>
                    <td class="px-3 py-2.5">
                        <div class="flex items-center gap-3">
                            <div
                                class="flex size-8 shrink-0 items-center justify-center rounded-md bg-background-200/80"
                            >
                                <Icon icon="key" class="text-lg" />
                            </div>
                            <div class="min-w-0">
                                <div class="flex items-center gap-2">
                                    <span class="truncate font-mono">
                                        {key.id}
                                    </span>
                                    <Badge tone={statusTones[key.status]}>
                                        {key.status}
                                    </Badge>
                                    <Badge class="max-sm:hidden">
                                        {key.algorithm}
                                    </Badge>
                                </div>
                                <p
                                    class="truncate font-mono text-xs text-muted max-sm:hidden"
                                    {@attach tooltip("SHA-256 JWK thumbprint")}
                                >
                                    {key.fingerprint}
                                </p>
                            </div>
                        </div>
                    </td>
                    <td class="px-3 text-muted whitespace-nowrap max-sm:hidden">
                        {#if key.expires_at}
                            <span
                                {@attach tooltip(
                                    formatTime(new Date(key.expires_at)),
                                )}
                            >
                                {#if Date.parse(key.expires_at) < Date.now()}
                                    {relativeTime(key.expires_at)} ago
                                {:else}
                                    in {relativeTime(key.expires_at)}
                                {/if}
                            </span>
                        {:else}
                            Never
                        {/if}
                    </td>
                    <td
                        class="px-3 text-right text-muted whitespace-nowrap max-sm:hidden"
                    >
                        {relativeTime(key.created_at)} ago
                    </td>
                    {#if canDelete}
                        <td class="px-3">
                            {#if key.status === "retiring"}
                                <Button
                                    variant="icon"
                                    aria-label="Delete {key.id}"
                                    onclick={() => confirmDelete(key)}
                                >
                                    <Icon icon="trash" class="text-red-300" />
                                </Button>
                            {/if}
                        </td>
                    {/if}
                </tr>
            {:else}
                <tr class="hover:bg-transparent">
                    <td colspan="4" class="py-10 text-center text-muted">
                        {loaded ? "No signing keys found." : "Loading keys..."}
                    </td>
                </tr>
            {/each}
        </tbody>
    </table>
</Card>

<ConfirmModal
    bind:this={rotateModal}
    title="Rotate signing keys"
    confirmLabel="Rotate keys"
    variant="primary"
    onconfirm={rotateKeys}
>
    The next key becomes the active signing key and a new next key is
    generated. The current key moves to retiring and stays in the JWKS until
    every token it signed has expired.
</ConfirmModal>

<ConfirmModal
    bind:this={deleteModal}
    title="Delete signing key"
    confirmLabel="Delete key"
    onconfirm={deleteKey}
>
    Tokens signed by this key will immediately stop validating. Retiring keys
    are removed automatically a week after they expire.
</ConfirmModal>
