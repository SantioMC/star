<script lang="ts">
    import { hasPermission } from "@/index";
    import type { Snippet } from "svelte";
    import Button from "./button.svelte";
    import PageHeader from "./page-header.svelte";

    const {
        title,
        description,
        dirty,
        issue = null,
        onsave,
        ondiscard,
        children,
        preview,
    }: {
        title: string;
        description?: string;
        dirty: boolean;
        issue?: string | null;
        onsave: () => Promise<unknown>;
        ondiscard: () => void;
        children?: Snippet;
        preview?: Snippet;
    } = $props();

    const canModify = hasPermission("admin:settings:write");

    let saving = $state(false);

    const save = async (event: Event) => {
        event.preventDefault();
        if (saving || !dirty || issue) return;

        saving = true;
        await onsave();
        saving = false;
    };
</script>

<PageHeader {title} {description} />

<form onsubmit={save} class="flex flex-col gap-8">
    <fieldset disabled={!canModify} class="contents">
        {@render children?.()}
    </fieldset>

    {@render preview?.()}

    {#if canModify}
        <div class="flex items-center justify-end gap-2">
            {#if dirty}
                <span class="mr-auto text-sm text-muted select-none">
                    You have unsaved changes
                </span>
                <Button type="button" variant="secondary" onclick={ondiscard}>
                    Discard
                </Button>
            {/if}
            <Button
                type="submit"
                loading={saving}
                disabled={!dirty || issue != null}
            >
                Save settings
            </Button>
        </div>
    {/if}
</form>
