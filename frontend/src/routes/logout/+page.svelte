<script lang="ts">
    import * as star from "$lib/star";
    import AuthShell from "@/components/auth-shell.svelte";
    import Icon from "@/components/icon.svelte";
    import { clearSession } from "@/index";
    import { onMount } from "svelte";

    onMount(async () => {
        await star.endSession();
        clearSession();

        location.href = "/";
    });
</script>

<AuthShell>
    <div class="flex flex-col">
        <h1 class="text-lg">Signing out</h1>
        <p class="text-muted text-sm">
            One moment while we sign you out...
        </p>
    </div>
    <div class="flex items-center justify-center h-40 w-full">
        <Icon icon="loader-2" class="animate-spin text-3xl text-muted" />
    </div>
</AuthShell>
