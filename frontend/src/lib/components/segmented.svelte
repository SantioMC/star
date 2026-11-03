<script lang="ts" generics="T extends string">
    import type { SelectOption } from "./select.svelte";

    let {
        value = $bindable(),
        options,
        label,
    }: {
        value: T;
        options: SelectOption<T>[];
        label: string;
    } = $props();

    const name = $props.id();
</script>

<div
    role="radiogroup"
    aria-label={label}
    class="flex w-fit max-w-full items-center gap-1 overflow-x-auto rounded-lg border border-background-200 bg-background p-1"
>
    {#each options as option (option.value)}
        <label
            class="shrink-0 rounded-md px-3 py-1 text-sm text-muted select-none transition-colors cursor-pointer hover:bg-background-200/60 hover:text-foreground has-checked:bg-background-200 has-checked:text-foreground has-focus-visible:outline-2 has-focus-visible:outline-primary has-focus-visible:-outline-offset-2"
        >
            <input
                type="radio"
                {name}
                value={option.value}
                bind:group={value}
                class="sr-only"
            />
            {option.label}
        </label>
    {/each}
</div>
