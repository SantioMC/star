<script lang="ts">
    import * as star from "$lib/star";
    import Button from "@/components/button.svelte";
    import Input from "@/components/input.svelte";
    import Modal from "@/components/modal.svelte";
    import { Label } from "@/components/ui/label";
    import { toast } from "svelte-sonner";

    const {
        user,
        onunsuspended,
    }: {
        user: star.AdminUser;
        onunsuspended: (user: star.AdminUser) => void;
    } = $props();

    let modal = $state<Modal>();
    let reason = $state("");
    let saving = $state(false);

    export const open = () => {
        reason = "";
        modal?.open();
    };

    const unsuspend = async (event: Event) => {
        event.preventDefault();
        if (saving) return;

        saving = true;
        const result = await star.unsuspendUser(user.id, reason);
        saving = false;

        if ("error" in result) {
            return toast.error(result.message);
        }

        modal?.close();
        onunsuspended(result);
        toast.success(`Unsuspended ${result.username}`);
    };
</script>

<Modal bind:this={modal} button>
    <form class="flex flex-col gap-6" onsubmit={unsuspend}>
        <div class="flex flex-col gap-1 pr-8">
            <h1 class="text-xl font-semibold">Unsuspend user</h1>
            <h2 class="text-sm text-muted">
                {user.username} will be able to sign in again immediately.
            </h2>
        </div>

        <div class="grid gap-2">
            <Label for="unsuspendReason">Reason</Label>
            <Input
                bind:value={reason}
                id="unsuspendReason"
                type="text"
                maxlength={500}
                autocomplete="off"
                required
            />
        </div>

        <div class="flex items-center gap-2 justify-end">
            <Button
                type="button"
                variant="secondary"
                onclick={() => modal?.close()}
            >
                Cancel
            </Button>
            <Button type="submit" loading={saving} disabled={!reason.trim()}>
                Unsuspend user
            </Button>
        </div>
    </form>
</Modal>
