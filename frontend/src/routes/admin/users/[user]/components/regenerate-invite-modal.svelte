<script lang="ts">
    import * as star from "$lib/star";
    import Button from "@/components/button.svelte";
    import InviteExpiryField, {
        defaultInviteExpiry,
    } from "@/components/invite-expiry-field.svelte";
    import Modal from "@/components/modal.svelte";
    import { toast } from "svelte-sonner";

    const {
        user,
        onregenerated,
    }: {
        user: star.AdminUser;
        onregenerated: (link: star.InviteLink) => void;
    } = $props();

    let modal = $state<Modal>();
    let expiry = $state(defaultInviteExpiry);
    let saving = $state(false);

    export const open = () => {
        expiry = defaultInviteExpiry;
        modal?.open();
    };

    const regenerate = async (event: Event) => {
        event.preventDefault();
        if (saving) return;

        saving = true;
        const result = await star.regenerateInvite(user.id, Number(expiry));
        saving = false;

        if ("error" in result) {
            return toast.error(result.message);
        }

        modal?.close();
        onregenerated(result);
    };
</script>

<Modal bind:this={modal} button>
    <form class="flex flex-col gap-6" onsubmit={regenerate}>
        <div class="flex flex-col gap-1 pr-8">
            <h1 class="text-xl font-semibold">Regenerate invite link</h1>
            <h2 class="text-sm text-muted">
                The current link for {user.username} stops working and a new one
                is created.
            </h2>
        </div>

        <InviteExpiryField bind:value={expiry} />

        <div class="flex items-center gap-2 justify-end">
            <Button
                type="button"
                variant="secondary"
                onclick={() => modal?.close()}
            >
                Cancel
            </Button>
            <Button type="submit" loading={saving}>Regenerate link</Button>
        </div>
    </form>
</Modal>
