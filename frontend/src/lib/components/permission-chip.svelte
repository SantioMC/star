<script lang="ts">
    import { tooltip } from "@/tooltip";
    import { cn } from "@/utils";
    import Icon from "./icon.svelte";

    const {
        name,
        label,
        description,
        removable = false,
        pending = false,
        disabled = false,
        onremove,
    }: {
        name: string;
        label: string;
        description?: string;
        removable?: boolean;
        pending?: boolean;
        disabled?: boolean;
        onremove?: () => void;
    } = $props();

    const base =
        "inline-flex items-center rounded-md border border-background-200 bg-background px-2 py-0.5 font-mono text-xs";
</script>

{#if removable}
    <button
        type="button"
        {@attach tooltip(description ?? name)}
        aria-label="Revoke {name}"
        disabled={disabled || pending}
        onclick={onremove}
        class={cn(
            base,
            "group transition-colors outline-none",
            "not-disabled:hover:border-red-400/60 not-disabled:hover:text-red-200",
            "focus-visible:border-red-400/60 focus-visible:text-red-200",
            pending && "opacity-50",
        )}
    >
        {label}
        <span
            class="grid grid-cols-[0fr] transition-[grid-template-columns] duration-150 group-not-disabled:group-hover:grid-cols-[1fr] group-focus-visible:grid-cols-[1fr]"
        >
            <span class="overflow-hidden">
                <Icon icon="x" class="ml-1 text-red-300" />
            </span>
        </span>
    </button>
{:else}
    <span {@attach tooltip(description ?? name)} class={base}>{label}</span>
{/if}
