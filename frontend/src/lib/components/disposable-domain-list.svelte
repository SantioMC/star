<script lang="ts">
    import * as star from "$lib/star";
    import { onMount } from "svelte";
    import Icon from "./icon.svelte";
    import SearchInput from "./search-input.svelte";

    let domains = $state<string[] | null>(null);
    let failed = $state(false);
    let query = $state("");

    const matches = $derived.by(() => {
        const search = query.trim().toLowerCase();
        return (domains ?? []).filter((domain) => domain.includes(search));
    });

    const visible = $derived(matches.slice(0, 200));

    onMount(async () => {
        domains = await star.getDisposableDomains();
        failed = domains === null;
    });

    const onkeydown = (event: KeyboardEvent) => {
        if (event.key === "Enter") event.preventDefault();
    };
</script>

<div class="mt-6 flex flex-col gap-3">
    {#if domains}
        <div class="flex flex-wrap items-center justify-between gap-2">
            <SearchInput
                bind:value={query}
                placeholder="Search domains"
                aria-label="Search disposable domains"
                {onkeydown}
            />
            <span class="text-sm text-muted">
                Showing {visible.length.toLocaleString()} of
                {matches.length.toLocaleString()}
            </span>
        </div>

        {#if visible.length > 0}
            <div class="max-h-80 overflow-y-auto rounded-lg border border-background-200 p-3">
                <ul class="grid gap-x-4 gap-y-1 font-mono text-sm sm:grid-cols-2">
                    {#each visible as domain (domain)}
                        <li class="truncate">{domain}</li>
                    {/each}
                </ul>
            </div>
        {:else}
            <p class="text-sm text-muted">No known disposable domain matches this search.</p>
        {/if}
    {:else if failed}
        <p role="alert" class="text-sm text-red-300">
            Failed to load the disposable domain list.
        </p>
    {:else}
        <Icon icon="loader-2" class="animate-spin text-2xl text-muted" />
    {/if}
</div>
