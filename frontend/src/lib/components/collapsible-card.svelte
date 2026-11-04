<script lang="ts">
    import type { Snippet } from "svelte";
    import Button from "./button.svelte";
    import Card from "./card.svelte";
    import Icon from "./icon.svelte";

    const {
        title,
        icon,
        description,
        children,
    }: {
        title: string;
        icon?: string;
        description?: string;
        children?: Snippet;
    } = $props();

    let expanded = $state(false);
</script>

<Card {title} {icon} {description}>
    {#snippet actions()}
        <Button
            type="button"
            variant="secondary"
            aria-expanded={expanded}
            onclick={() => (expanded = !expanded)}
        >
            {expanded ? "Hide" : "Show"}
            <Icon icon={expanded ? "chevron-up" : "chevron-down"} />
        </Button>
    {/snippet}

    {#if expanded}
        {@render children?.()}
    {/if}
</Card>
