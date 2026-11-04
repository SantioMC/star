<script lang="ts">
    import * as star from "$lib/star";
    import Button from "@/components/button.svelte";
    import Input from "@/components/input.svelte";
    import Modal from "@/components/modal.svelte";
    import Segmented from "@/components/segmented.svelte";
    import { Label } from "@/components/ui/label";
    import { toast } from "svelte-sonner";

    const {
        user,
        onsuspended,
    }: {
        user: star.AdminUser;
        onsuspended: (user: star.AdminUser) => void;
    } = $props();

    type Duration = "temporary" | "permanent";

    const durations: { value: Duration; label: string }[] = [
        { value: "temporary", label: "Temporary" },
        { value: "permanent", label: "Permanent" },
    ];

    let modal = $state<Modal>();
    let duration = $state<Duration>("temporary");
    let reason = $state("");
    let until = $state("");
    let minimum = $state("");
    let saving = $state(false);

    const toLocalInput = (date: Date): string => {
        const offset = date.getTimezoneOffset() * 60_000;
        const local = new Date(date.getTime() - offset);

        return local.toISOString().slice(0, 16);
    };

    export const open = () => {
        const now = Date.now();

        duration = "temporary";
        reason = "";
        minimum = toLocalInput(new Date(now));
        until = toLocalInput(new Date(now + 7 * 24 * 60 * 60 * 1000));
        modal?.open();
    };

    const suspend = async (event: Event) => {
        event.preventDefault();
        if (saving) return;

        const expiresAt =
            duration === "temporary" ? new Date(until).toISOString() : null;

        saving = true;
        const result = await star.suspendUser(user.id, {
            reason,
            expires_at: expiresAt,
        });
        saving = false;

        if ("error" in result) {
            return toast.error(result.message);
        }

        modal?.close();
        onsuspended(result);
        toast.success(`Suspended ${result.username}`);
    };
</script>

<Modal bind:this={modal} button>
    <form class="flex flex-col gap-6" onsubmit={suspend}>
        <div class="flex flex-col gap-1 pr-8">
            <h1 class="text-xl font-semibold">Suspend user</h1>
            <h2 class="text-sm text-muted">
                {user.username} will be signed out everywhere and can not sign
                in or use any application until the suspension ends.
            </h2>
        </div>

        <div class="flex flex-col gap-5">
            <Segmented
                bind:value={duration}
                options={durations}
                label="Suspension duration"
            />

            {#if duration === "temporary"}
                <div class="grid gap-2">
                    <Label for="suspendedUntil">Suspended Until</Label>
                    <Input
                        bind:value={until}
                        id="suspendedUntil"
                        type="datetime-local"
                        min={minimum}
                        class="scheme-dark"
                        required
                    />
                </div>
            {/if}

            <div class="grid gap-2">
                <Label for="suspensionReason">Reason</Label>
                <Input
                    bind:value={reason}
                    id="suspensionReason"
                    type="text"
                    maxlength={500}
                    autocomplete="off"
                    required
                />
            </div>
        </div>

        <div class="flex items-center gap-2 justify-end">
            <Button
                type="button"
                variant="secondary"
                onclick={() => modal?.close()}
            >
                Cancel
            </Button>
            <Button
                type="submit"
                variant="danger"
                loading={saving}
                disabled={!reason.trim() ||
                    (duration === "temporary" && !until)}
            >
                Suspend user
            </Button>
        </div>
    </form>
</Modal>
