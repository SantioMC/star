<script lang="ts">
    import Sidebar from "@/components/sidebar/index.svelte";
    import { user } from "@/index";
    import { onMount } from "svelte";
    import * as star from "@/star";
    import { goto } from "$app/navigation";

    let { children } = $props();

    onMount(async () => {
        if (!!$user) return;

        const data = await star.getSelf();
        if (!data) return goto("/login");

        user.set(data);
    });
</script>

{#if $user}
    <Sidebar>
        <div class="flex flex-col md:w-3xl mx-auto py-12 px-4 gap-8">
            {@render children()}
        </div>
    </Sidebar>
{/if}
