<script lang="ts">
    import Button from "./button.svelte";
    import Icon from "./icon.svelte";

    const {
        page,
        hasNext,
        loaded,
        onprevious,
        onnext,
    }: {
        page: number;
        hasNext: boolean;
        loaded: boolean;
        onprevious: () => void;
        onnext: () => void;
    } = $props();
</script>

{#if page > 1 || hasNext}
    <div class="flex items-center justify-end gap-2 -mt-4">
        <span class="mr-2 text-sm text-muted select-none">Page {page}</span>
        <Button
            variant="secondary"
            disabled={page <= 1 || !loaded}
            onclick={onprevious}
            aria-label="Previous page"
        >
            <Icon icon="chevron-left" />
        </Button>
        <Button
            variant="secondary"
            disabled={!hasNext || !loaded}
            onclick={onnext}
            aria-label="Next page"
        >
            <Icon icon="chevron-right" />
        </Button>
    </div>
{/if}
