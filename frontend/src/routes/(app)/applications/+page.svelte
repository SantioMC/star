<script lang="ts">
    import Button from "@/components/button.svelte";
    import Card from "@/components/card.svelte";
    import ConfirmModal from "@/components/confirm-modal.svelte";
    import GrantedScopes from "@/components/granted-scopes.svelte";
    import Icon from "@/components/icon.svelte";
    import PageHeader from "@/components/page-header.svelte";
    import { hasPermission } from "@/index";
    import * as star from "@/star";
    import { tooltip } from "@/tooltip";
    import { formatTime, relativeTime } from "@/utils";
    import { onMount } from "svelte";
    import { toast } from "svelte-sonner";

    let applications = $state<star.AuthorizedApplication[] | null>(null);
    let pending = $state<star.AuthorizedApplication | null>(null);
    let revokeModal = $state<ConfirmModal>();

    const canRevoke = hasPermission("user:application:revoke");

    const hostOf = (url: string) =>
        URL.canParse(url) ? new URL(url).host : url;

    onMount(async () => {
        applications = (await star.getAuthorizedApps()) ?? [];
    });

    const confirmRevoke = (application: star.AuthorizedApplication) => {
        pending = application;
        revokeModal?.open();
    };

    const revoke = async () => {
        if (!pending) return;

        const response = await star.revokeAuthorizedApp(pending.id);
        if ("error" in response) {
            toast.error(response.message);
            return false;
        }

        const revokedId = pending.id;
        applications = applications?.filter((it) => it.id !== revokedId) ?? [];

        toast.success(`Revoked access for ${pending.name}`);
    };
</script>

<PageHeader
    title="Authorized Applications"
    description="Applications you have allowed to access your account"
/>

{#if applications == null}
    <Card>
        <p class="py-6 text-center text-sm text-muted">
            Loading applications...
        </p>
    </Card>
{:else}
    {#each applications as application (application.id)}
        <Card title={application.name} icon="apps" class="flex flex-col gap-4">
            {#snippet actions()}
                {#if canRevoke}
                    <Button
                        variant="danger"
                        onclick={() => confirmRevoke(application)}
                    >
                        Revoke access
                    </Button>
                {/if}
            {/snippet}

            <p
                class="-mt-2 flex flex-wrap items-center gap-x-1.5 text-xs text-muted"
            >
                <span class="font-mono">{hostOf(application.url)}</span>
                &middot;
                <span
                    {@attach tooltip(
                        formatTime(new Date(application.authorized_at)),
                    )}
                >
                    authorized {relativeTime(application.authorized_at)} ago
                </span>
                &middot;
                {#if application.last_used_at}
                    <span
                        {@attach tooltip(
                            formatTime(new Date(application.last_used_at)),
                        )}
                    >
                        last used {relativeTime(application.last_used_at)} ago
                    </span>
                {:else}
                    <span>no active sessions</span>
                {/if}
            </p>

            <GrantedScopes scopes={application.scopes} />
        </Card>
    {:else}
        <Card>
            <div class="flex flex-col items-center gap-2 py-8 text-center">
                <Icon icon="apps" class="text-2xl text-muted" />
                <p class="text-sm">No authorized applications</p>
                <p class="text-sm text-muted">
                    Applications you sign in to with your account will appear
                    here.
                </p>
            </div>
        </Card>
    {/each}
{/if}

<ConfirmModal
    bind:this={revokeModal}
    title={`Revoke access for ${pending?.name ?? "application"}`}
    confirmLabel="Revoke access"
    onconfirm={revoke}
>
    {pending?.name ?? "This application"} will be signed out and lose access to your
    account. Access tokens it already holds stop working within a few seconds.
</ConfirmModal>
