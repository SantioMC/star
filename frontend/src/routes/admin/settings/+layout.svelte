<script lang="ts">
    import { goto } from "$app/navigation";
    import { hasPermission } from "@/index";
    import { onMount } from "svelte";
    import { SettingsContext, setSettingsContext } from "./context.svelte";

    let { children } = $props();

    const context = new SettingsContext();
    setSettingsContext(context);

    onMount(() => {
        if (!hasPermission("admin:settings:read")) return goto("/");
        context.load();
    });
</script>

{#if context.current}
    {@render children()}
{/if}
