<script lang="ts">
    import { replaceState } from "$app/navigation";
    import { resolve } from "$app/paths";
    import { page } from "$app/state";
    import Card from "@/components/card.svelte";
    import ConfirmModal from "@/components/confirm-modal.svelte";
    import Icon from "@/components/icon.svelte";
    import * as star from "@/star";
    import { tooltip } from "@/tooltip";
    import { cn, titleCase } from "@/utils";
    import { onMount } from "svelte";
    import { toast } from "svelte-sonner";

    let connections = $state<star.Connection[]>([]);
    let highlighted = $state<string | null>(null);
    let starting = $state<string | null>(null);

    let removeModal = $state<ConfirmModal>();
    let pending = $state<star.Connection | null>(null);

    const sorted = $derived(
        [...connections].sort((a, b) => {
            if (!a.account !== !b.account) return a.account ? -1 : 1;
            return a.name.localeCompare(b.name);
        }),
    );

    onMount(async () => {
        const connected = page.url.searchParams.get("connected");
        const error = page.url.searchParams.get("connection_error");

        if (connected || error) replaceState(resolve("/"), {});
        if (error) toast.error(error);

        connections = await star.getConnections();

        const match = connections.find((it) => it.id === connected);
        if (!match) return;

        toast.success(`Connected with ${titleCase(match.name)}`);
        highlighted = match.id;
        setTimeout(() => (highlighted = null), 4000);
    });

    const connect = async (connection: star.Connection) => {
        if (starting) return;

        starting = connection.id;
        const response = await star.startConnection(connection.id);

        if ("error" in response) {
            starting = null;
            return toast.error(response.message);
        }

        location.assign(response.url);
    };

    const confirmRemove = (connection: star.Connection) => {
        pending = connection;
        removeModal?.open();
    };

    const remove = async () => {
        const connection = pending;
        if (!connection) return;

        const response = await star.deleteConnection(connection.id);
        if ("error" in response) {
            toast.error(response.message);
            return false;
        }

        connections = await star.getConnections();
        toast.success(`Removed your ${titleCase(connection.name)} connection`);
    };
</script>

{#if connections.length > 0}
    <Card
        title="Connections"
        description="Link your other accounts to sign in with them."
        class="flex flex-col gap-3"
    >
        <div class="grid gap-2 sm:grid-cols-2">
            {#each sorted as connection (connection.id)}
                <button
                    type="button"
                    disabled={starting != null}
                    aria-label={connection.account
                        ? `Connected with ${titleCase(connection.name)}, click to remove connection`
                        : undefined}
                    {@attach tooltip(
                        connection.account
                            ? "Click to remove connection"
                            : null,
                    )}
                    onclick={() =>
                        connection.account
                            ? confirmRemove(connection)
                            : connect(connection)}
                    class={cn(
                        "flex min-h-14 w-full min-w-0 items-center gap-3 rounded-lg border border-background-200 px-3 py-2 text-left text-sm transition-colors duration-500 not-disabled:cursor-pointer not-disabled:hover:bg-background-200 disabled:brightness-50",
                        highlighted === connection.id && "bg-background-200",
                    )}
                >
                    <Icon icon={connection.icon || "login"} />

                    {#if connection.account}
                        <span class="flex min-w-0 flex-col">
                            <span class="truncate">
                                Connected with {titleCase(connection.name)}
                            </span>
                            {#if connection.account.name}
                                <span
                                    class="truncate font-mono text-xs text-muted"
                                >
                                    {connection.account.name}
                                </span>
                            {/if}
                        </span>
                        <Icon
                            icon="check"
                            class="ml-auto shrink-0 text-emerald-300"
                        />
                    {:else}
                        <span class="truncate">
                            Connect with {titleCase(connection.name)}
                        </span>
                        {#if starting === connection.id}
                            <Icon
                                icon="loader-2"
                                class="ml-auto animate-spin"
                            />
                        {/if}
                    {/if}
                </button>
            {/each}
        </div>
    </Card>

    <ConfirmModal
        bind:this={removeModal}
        title="Remove connection"
        confirmLabel="Remove connection"
        onconfirm={remove}
    >
        You will no longer be able to sign in with
        <span class="text-foreground">{titleCase(pending?.name ?? "")}</span>,
        and other devices signed in with it will be signed out.
    </ConfirmModal>
{/if}
