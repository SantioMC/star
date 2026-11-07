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

    const visible = $derived(scopes.filter((it) => it.scope !== "openid"));
</script>

<ul class={cn("flex flex-col gap-3", className)}>
    {#each visible as entry (entry.scope)}
        <li class="flex items-start gap-3 text-sm">
            <Icon icon="check" class="text-emerald-300 mt-0.5" />
            <span class="min-w-0">{entry.description}</span>
        </li>
    {:else}
        <li class="text-sm text-muted">
            Basic access to your account identity.
        </li>
    {/each}
</ul>
