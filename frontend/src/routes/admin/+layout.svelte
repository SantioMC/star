<script lang="ts">
    import Sidebar from "@/components/sidebar/index.svelte";
    import { user, syncUser } from "@/index";
    import { onMount } from "svelte";
    import { goto } from "$app/navigation";

    let { children } = $props();

    onMount(async () => {
        const synced = await syncUser();
        if (!synced) return goto("/login");
    });
</script>

{#if $user}
    <Sidebar>
        <div class="flex flex-col md:w-3xl mx-auto py-12 px-4 gap-8">
            {@render children()}
        </div>
    </Sidebar>
{/if}
