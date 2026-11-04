<script lang="ts">
    import { goto } from "$app/navigation";
    import { page } from "$app/state";
    import Badge from "@/components/badge.svelte";
    import Button from "@/components/button.svelte";
    import Card from "@/components/card.svelte";
    import Pagination from "@/components/pagination.svelte";
    import SessionTable from "@/components/session-table.svelte";
    import { getJWT, hasPermission } from "@/index";
    import * as star from "@/star";
    import { onMount } from "svelte";
    import { toast } from "svelte-sonner";

    let sessions = $state<star.Session[]>([]);
    let next = $state<string | null>(null);
    let cursors = $state<string[]>([]);
    let loaded = $state(false);
    let pendingSession = $state<string | null>(null);

    const currentSession = getJWT()?.sid;
    const canDelete = hasPermission("admin:user:delete");

    let latestRequest = 0;

    const load = async () => {
        const request = ++latestRequest;
        loaded = false;

        const result = await star.getUserSessions(
            page.params.user!,
            cursors.at(-1),
        );

        if (request !== latestRequest) return;
        loaded = true;

        if ("error" in result) {
            sessions = [];
            next = null;
            return toast.error(result.message);
        }

        if (result.items.length === 0 && cursors.length > 0) {
            return previousPage();
        }

        sessions = result.items;
        next = result.next;
    };

    const nextPage = () => {
        if (!next) return;

        cursors = [...cursors, next];
        load();
    };

    const previousPage = () => {
        cursors = cursors.slice(0, -1);
        load();
    };

    onMount(load);

    const invalidate = async (session: star.Session) => {
        if (pendingSession) return;

        pendingSession = session.id;
        const response = await star.invalidateUserSession(
            page.params.user!,
            session.id,
        );
        pendingSession = null;

        if ("error" in response) {
            return toast.error(response.message);
        }

        if (session.id === currentSession) {
            return goto("/logout");
        }

        toast.success("Session signed out");
        load();
    };
</script>

<Card
    title="Active Sessions"
    icon="device-desktop"
    description="Devices currently signed in to this account. Signing out a session ends its access immediately."
    class="flex flex-col gap-4"
>
    <SessionTable {sessions} {loaded}>
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

<Pagination
    page={cursors.length + 1}
    hasNext={next !== null}
    {loaded}
    onprevious={previousPage}
    onnext={nextPage}
/>
