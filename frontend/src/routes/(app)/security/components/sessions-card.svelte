<script lang="ts">
    import { goto } from "$app/navigation";
    import Badge from "@/components/badge.svelte";
    import Button from "@/components/button.svelte";
    import Card from "@/components/card.svelte";
    import ConfirmModal from "@/components/confirm-modal.svelte";
    import SessionTable from "@/components/session-table.svelte";
    import { getJWT, hasPermission } from "@/index";
    import * as star from "@/star";
    import { isExpired } from "@/utils";
    import { onMount } from "svelte";
    import { toast } from "svelte-sonner";

    let sessions = $state<star.Session[]>([]);
    let loaded = $state(false);
    let pendingSession = $state<string | null>(null);
    let signOutAllModal = $state<ConfirmModal>();

    const currentSession = getJWT()?.sid;
    const canDelete = hasPermission("user:session:delete");

    const activeSessions = $derived(
        sessions
            .filter((it) => !isExpired(it))
            .sort((a, b) => {
                if (a.id === currentSession) return -1;
                if (b.id === currentSession) return 1;
                return Date.parse(b.created_at) - Date.parse(a.created_at);
            }),
    );

    onMount(async () => {
        sessions = await star.getSessions();
        loaded = true;
    });

    const invalidate = async (session: star.Session) => {
        if (pendingSession) return;

        pendingSession = session.id;
        const response = await star.invalidateSession(session.id);
        pendingSession = null;

        if ("error" in response) {
            return toast.error(response.message);
        }

        if (session.id === currentSession) {
            return goto("/logout");
        }

        sessions = sessions.filter((it) => it.id !== session.id);
        toast.success("Session signed out");
    };

    const invalidateAll = async () => {
        const response = await star.invalidateAllSessions();

        if ("error" in response) {
            toast.error(response.message);
            return false;
        }

        goto("/logout");
    };
</script>

<Card
    title="Active Sessions"
    icon="device-desktop"
    description="Devices currently signed in to your account. Signing out a session ends its access immediately."
    class="flex flex-col gap-4"
>
    {#snippet actions()}
        {#if canDelete && activeSessions.length > 0}
            <Button variant="danger" onclick={() => signOutAllModal?.open()}>
                Sign out all
            </Button>
        {/if}
    {/snippet}

    <SessionTable sessions={activeSessions} {loaded}>
        {#snippet badge(session)}
            {#if session.id === currentSession}
                <Badge tone="success">this device</Badge>
            {/if}
        {/snippet}

        {#snippet action(session)}
            {#if canDelete}
                <Button
                    variant="ghost"
                    class="text-xs text-red-300 whitespace-nowrap"
                    loading={pendingSession === session.id}
                    disabled={pendingSession != null}
                    onclick={() => invalidate(session)}
                >
                    Sign out
                </Button>
            {/if}
        {/snippet}
    </SessionTable>
</Card>

<ConfirmModal
    bind:this={signOutAllModal}
    title="Sign out all sessions"
    confirmLabel="Sign out everywhere"
    onconfirm={invalidateAll}
>
    Every device signed in to your account, including this one, will be signed
    out. You will need to log in again.
</ConfirmModal>
