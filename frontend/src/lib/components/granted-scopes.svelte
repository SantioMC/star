<script lang="ts">
    import Icon from "@/components/icon.svelte";
    import type { GrantedScope } from "@/star";
    import { cn } from "@/utils";
    import type { ClassValue } from "svelte/elements";

    const {
        scopes,
        class: className,
    }: {
        scopes: GrantedScope[];
        class?: ClassValue;
    } = $props();
</script>

<ul
    class={cn(
        "flex flex-col rounded-lg border border-background-200 divide-y divide-background-200",
        className,
    )}
>
    {#each scopes as entry (entry.scope)}
        <li class="flex items-start gap-3 px-3 py-2.5 text-sm">
            <Icon icon="check" class="text-emerald-300 mt-0.5" />
            <div class="flex min-w-0 flex-col">
                <span>{entry.description}</span>
                <span class="font-mono text-xs text-muted truncate select-none">
                    {entry.scope}
                </span>
            </div>
        </li>
    {:else}
        <li class="px-3 py-4 text-center text-sm text-muted">
            Basic access to your account identity.
        </li>
    {/each}
</ul>
