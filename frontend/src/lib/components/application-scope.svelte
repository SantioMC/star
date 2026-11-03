<script lang="ts">
    import { tooltip } from "@/tooltip";
    import * as star from "$lib/star";
    import Badge from "@/components/badge.svelte";
    import Button from "@/components/button.svelte";
    import ConfirmModal from "@/components/confirm-modal.svelte";
    import Icon from "@/components/icon.svelte";
    import Modal from "@/components/modal.svelte";
    import ScopeFields from "@/components/scope-fields.svelte";
    import ScopeItem from "@/components/scope-item.svelte";
    import { relativeTime } from "@/utils";
    import { toast } from "svelte-sonner";

    const {
        application,
        scope,
        namespace,
        canModify = true,
        onchange,
    }: {
        application: string;
        scope: star.ApplicationScope;
        namespace: string;
        canModify?: boolean;
        onchange?: () => void;
    } = $props();

    let loading = $state(false);
    let editModal = $state<Modal>();
    let deleteModal = $state<ConfirmModal>();

    let name = $state("");
    let description = $state("");
    let requestable = $state(false);

    const openEdit = () => {
        name = scope.name;
        description = scope.description;
        requestable = scope.requestable;
        editModal?.open();
    };

    const edit = async (event: Event) => {
        event.preventDefault();
        if (loading) return;

        loading = true;
        const result = await star.updateApplicationScope(application, scope.id, {
            name: name.trim(),
            description: description.trim(),
            requestable,
        });
        loading = false;

        if ("error" in result) {
            return toast.error(result.message);
        }

        editModal?.close();
        onchange?.();
        toast.success("Successfully updated scope");
    };

    const remove = async () => {
        const result = await star.deleteApplicationScope(application, scope.id);

        if ("error" in result) {
            toast.error(result.message);
            return false;
        }

        onchange?.();
        toast.success("Removed scope");
    };
</script>

<ScopeItem
    name={scope.name}
    description="{scope.description} · added {relativeTime(
        scope.created_at,
    )} ago"
>
    {#snippet badges()}
        {#if scope.requestable}
            <Badge tone="warning">requestable</Badge>
        {/if}
    {/snippet}

    {#snippet actions()}
        {#if canModify}
            <Button
                variant="icon"
                aria-label="Edit {scope.name}"
                {@attach tooltip("Edit scope")}
                onclick={openEdit}
            >
                <Icon icon="pencil" />
            </Button>
            <Button
                variant="icon"
                aria-label="Remove {scope.name}"
                {@attach tooltip("Remove scope")}
                onclick={() => deleteModal?.open()}
            >
                <Icon icon="trash" class="text-red-300" />
            </Button>
        {/if}
    {/snippet}
</ScopeItem>

{#if canModify}
    <Modal bind:this={editModal} button>
        <form class="flex flex-col gap-6" onsubmit={edit}>
            <div class="flex flex-col gap-1 pr-8">
                <h1 class="text-xl font-semibold">Edit scope</h1>
                <h2 class="text-sm text-muted">
                    Scopes must stay inside this application's
                    <span class="font-mono text-foreground">{namespace}:</span>
                    namespace. The description is shown to users when they authorize
                    the application.
                </h2>
            </div>

            <ScopeFields
                id="edit_{scope.id}"
                {namespace}
                bind:name
                bind:description
                bind:requestable
            />

            <div class="flex items-center gap-2 justify-end">
                <Button
                    type="button"
                    variant="secondary"
                    onclick={() => editModal?.close()}
                >
                    Cancel
                </Button>
                <Button type="submit" {loading}>Save scope</Button>
            </div>
        </form>
    </Modal>

    <ConfirmModal
        bind:this={deleteModal}
        title="Remove scope"
        confirmLabel="Remove scope"
        onconfirm={remove}
    >
        <span class="font-mono text-foreground">{scope.name}</span> will be
        removed from this application. Tokens issued with this scope lose it the
        next time they are refreshed.
    </ConfirmModal>
{/if}
