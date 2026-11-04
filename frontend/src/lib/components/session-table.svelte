<script lang="ts">
    import { tooltip } from "@/tooltip";
    import Icon from "@/components/icon.svelte";
    import type { Session } from "@/star";
    import { getIcon, relativeTime } from "@/utils";
    import type { Snippet } from "svelte";

    const {
        sessions,
        loaded,
        badge,
        action,
    }: {
        sessions: Session[];
        loaded: boolean;
        badge?: Snippet<[Session]>;
        action?: Snippet<[Session]>;
    } = $props();
</script>

<table>
    <tbody>
        {#each sessions as session (session.id)}
            <tr>
                <td class="w-10 pl-1 align-middle">
                    <Icon icon={getIcon(session)} class="text-lg text-muted" />
                </td>
                <td class="py-2.5">
                    <div class="flex flex-wrap items-center gap-2">
                        <span {@attach tooltip(session.user_agent)}>
                            {session.geolocation ?? session.ip_address}
                        </span>
                        {@render badge?.(session)}
                    </div>
                    <p class="text-muted text-xs mt-0.5">
                        {session.browser !== "Unknown"
                            ? session.browser
                            : session.user_agent}
                        &middot;
                        <span class="font-mono">{session.ip_address}</span>
                        &middot; signed in {relativeTime(session.created_at)} ago
                    </p>
                </td>
                {#if action}
                    <td class="w-px pr-1 text-right">
                        {@render action(session)}
                    </td>
                {/if}
            </tr>
        {:else}
            <tr>
                <td class="py-6 text-center text-sm text-muted">
                    {loaded ? "No active sessions found." : "Loading sessions..."}
                </td>
            </tr>
        {/each}
    </tbody>
</table>
