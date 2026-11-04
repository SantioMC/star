<script lang="ts">
    import { goto } from "$app/navigation";
    import { resolve } from "$app/paths";
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

    const canRevoke = hasPermission("admin:token:revoke");

    const kindTones: Record<star.RevocationKind, BadgeTone> = {
        session: "info",
        user: "warning",
        global: "danger",
    };

    const kindIcons: Record<star.RevocationKind, string> = {
        session: "device-desktop",
        user: "user",
        global: "world",
    };

    const reasons: Record<star.RevocationReason, string> = {
        signed_out: "Signed out",
        sessions_cleared: "All sessions signed out",
        grant_revoked: "Application access revoked",
        token_replayed: "Refresh token replayed",
        code_replayed: "Authorization code replayed",
        roles_changed: "Roles or permissions changed",
        password_reset: "Password reset by an admin",
        password_change_required: "Password change required by an admin",
        user_deleted: "User deleted",
        user_suspended: "User suspended",
        application_deleted: "Application deleted",
        manual: "Revoked by an admin",
    };

    let revocations = $state<star.TokenRevocation[]>([]);
    let page = $state(1);
    let loaded = $state(false);
    let hasMore = $state(false);

    let revokeAllModal = $state<ConfirmModal>();

    const load = async (next: number) => {
        const result = await star.getTokenRevocations(next);

        revocations = next === 1 ? result : [...revocations, ...result];
        hasMore = result.length === 25;
        page = next;
        loaded = true;
    };

    onMount(() => {
        if (!hasPermission("admin:token:read")) return goto("/");
        load(1);
    });

    const revokeAll = async () => {
        const result = await star.revokeTokens({ kind: "global" });
        if ("error" in result) {
            toast.error(result.message);
            return false;
        }

        revocations = result;
        hasMore = result.length === 25;
        page = 1;
        toast.success("Revoked every issued access token");
    };
</script>

<PageHeader
    title="Token Revocations"
    description="Sessions and access tokens that stopped validating before they expired"
>
    {#snippet actions()}
        {#if canRevoke}
            <Button variant="danger" onclick={() => revokeAllModal?.open()}>
                <Icon icon="ban" />
                Revoke all tokens
            </Button>
        {/if}
    {/snippet}
</PageHeader>

<div class="flex flex-col gap-3">
    <Card class="p-2 sm:p-3">
        <table class="w-full text-sm">
            <thead>
                <tr class="hover:bg-transparent">
                    <th class="px-3 text-xs font-medium text-muted select-none">
                        Subject
                    </th>
                    <th
                        class="px-3 text-xs font-medium text-muted select-none max-sm:hidden"
                    >
                        Reason
                    </th>
                    <th
                        class="px-3 text-xs font-medium text-muted select-none text-right"
                    >
                        Revoked at
                    </th>
                </tr>
            </thead>
            <tbody>
                {#each revocations as revocation (revocation.subject)}
                    <tr>
                        <td class="px-3 py-2.5">
                            <div class="flex items-center gap-3">
                                <div
                                    class="flex size-8 shrink-0 items-center justify-center rounded-md bg-background-200/80"
                                >
                                    <Icon
                                        icon={kindIcons[revocation.kind]}
                                        class="text-lg"
                                    />
                                </div>
                                <div class="min-w-0">
                                    <div class="flex items-center gap-2">
                                        {#if revocation.kind === "global"}
                                            <span>Every user</span>
                                        {:else if revocation.user_id}
                                            <a
                                                href={resolve(
                                                    "/admin/users/[user]",
                                                    { user: revocation.user_id },
                                                )}
                                                class="truncate hover:text-primary transition-colors"
                                            >
                                                {revocation.user_name ??
                                                    revocation.user_id}
                                            </a>
                                        {/if}
                                        <Badge tone={kindTones[revocation.kind]}>
                                            {revocation.kind}
                                        </Badge>
                                    </div>
                                    {#if revocation.kind === "session"}
                                        <p
                                            class="truncate font-mono text-xs text-muted"
                                        >
                                            {revocation.subject}
                                        </p>
                                    {/if}
                                    <p class="text-xs text-muted sm:hidden">
                                        {reasons[revocation.reason] ??
                                            revocation.reason}
                                    </p>
                                </div>
                            </div>
                        </td>
                        <td class="px-3 text-muted max-sm:hidden">
                            {reasons[revocation.reason] ?? revocation.reason}
                        </td>
                        <td
                            class="px-3 text-right text-muted whitespace-nowrap"
                        >
                            <span
                                {@attach tooltip(
                                    `Tracked until ${formatTime(new Date(revocation.expires_at))}`,
                                )}
                            >
                                {relativeTime(revocation.revoked_at)} ago
                            </span>
                        </td>
                    </tr>
                {:else}
                    <tr class="hover:bg-transparent">
                        <td colspan="3" class="py-10 text-center text-muted">
                            {loaded
                                ? "No tokens have been revoked recently."
                                : "Loading revocations..."}
                        </td>
                    </tr>
                {/each}
            </tbody>
        </table>
    </Card>

    {#if hasMore}
        <Button
            variant="secondary"
            class="self-center"
            onclick={() => load(page + 1)}
        >
            Load more
        </Button>
    {/if}
</div>

<ConfirmModal
    bind:this={revokeAllModal}
    title="Revoke all access tokens"
    confirmLabel="Revoke tokens"
    onconfirm={revokeAll}
>
    Every access token issued until now stops validating within a few seconds,
    including your own. Sessions stay signed in, so applications holding a
    refresh token receive a new access token with up to date roles and
    permissions on their next refresh.
</ConfirmModal>
