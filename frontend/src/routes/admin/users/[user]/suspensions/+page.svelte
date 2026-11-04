<script lang="ts">
    import { page } from "$app/state";
    import Badge, { type BadgeTone } from "@/components/badge.svelte";
    import Card from "@/components/card.svelte";
    import Pagination from "@/components/pagination.svelte";
    import { getResource } from "@/resource.svelte";
    import * as star from "@/star";
    import { formatTime } from "@/utils";
    import { onMount, untrack } from "svelte";
    import { toast } from "svelte-sonner";

    const context = getResource<star.AdminUser>();

    let suspensions = $state<star.UserSuspension[]>([]);
    let next = $state<string | null>(null);
    let cursors = $state<string[]>([]);
    let loaded = $state(false);

    let latestRequest = 0;

    const load = async () => {
        const request = ++latestRequest;
        loaded = false;

        const result = await star.getUserSuspensions(
            page.params.user!,
            cursors.at(-1),
        );

        if (request !== latestRequest) return;
        loaded = true;

        if ("error" in result) {
            suspensions = [];
            next = null;
            return toast.error(result.message);
        }

        suspensions = result.items;
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

    const status = (
        suspension: star.UserSuspension,
    ): { label: string; tone: BadgeTone } => {
        if (suspension.lifted_at) return { label: "lifted", tone: "success" };

        if (suspension.id === context.current?.suspension?.id) {
            return { label: "active", tone: "danger" };
        }

        return { label: "expired", tone: "neutral" };
    };

    onMount(load);

    $effect(() => {
        context.current?.suspension;

        untrack(() => {
            if (loaded) load();
        });
    });
</script>

<Card
    title="Suspension History"
    icon="user-off"
    description="Every suspension placed on this account and how it ended."
    class="flex flex-col gap-4"
>
    <table class="w-full table-fixed">
        <tbody>
            {#each suspensions as suspension (suspension.id)}
                {@const state = status(suspension)}
                <tr>
                    <td class="py-2.5 pl-1 break-words">
                        <div class="flex flex-wrap items-center gap-2">
                            <span class="min-w-0 break-words">
                                {suspension.reason}
                            </span>
                            <Badge tone={state.tone}>{state.label}</Badge>
                        </div>
                        <p class="text-muted text-xs mt-0.5">
                            Suspended by
                            {suspension.suspended_by_name ?? "a deleted user"}
                            on {formatTime(new Date(suspension.created_at))}
                            &middot;
                            {suspension.expires_at
                                ? `until ${formatTime(new Date(suspension.expires_at))}`
                                : "permanent"}
                        </p>
                        {#if suspension.lifted_at}
                            <p class="text-muted text-xs mt-0.5">
                                Lifted by
                                {suspension.lifted_by_name ?? "a deleted user"}
                                on {formatTime(new Date(suspension.lifted_at))}:
                                {suspension.lift_reason}
                            </p>
                        {/if}
                    </td>
                </tr>
            {:else}
                <tr class="hover:bg-transparent">
                    <td class="py-6 text-center text-sm text-muted">
                        {loaded
                            ? "This user has never been suspended."
                            : "Loading suspensions..."}
                    </td>
                </tr>
            {/each}
        </tbody>
    </table>
</Card>

<Pagination
    page={cursors.length + 1}
    hasNext={next !== null}
    {loaded}
    onprevious={previousPage}
    onnext={nextPage}
/>
