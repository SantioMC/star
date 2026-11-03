<script lang="ts">
    import { page } from "$app/state";
    import favicon from "$lib/assets/favicon.svg";
    import { branding, loadBranding } from "@/branding";
    import { Toaster } from "@/components/ui/sonner";
    import { isLoggedIn, restoreSession } from "@/index";
    import { onMount } from "svelte";
    import "./layout.css";

    let { children } = $props();

    const PUBLIC_ROUTES = [
        "/login",
        "/register",
        "/verify",
        "/forgot-password",
        "/reset-password",
        "/oauth/authorize",
    ];

    let ready = $state(false);

    onMount(async () => {
        await Promise.all([restoreSession(), loadBranding()]);

        if (!PUBLIC_ROUTES.includes(page.url.pathname) && !isLoggedIn()) {
            location.href = "/login";
            return;
        }

        ready = true;
    });
</script>

<svelte:head>
    <title>{$branding.name}</title>
    <link rel="icon" href={$branding.logo_url ?? favicon} />
</svelte:head>
<Toaster
    position="bottom-right"
    duration={3000}
    theme="dark"
    richColors
    style="--normal-bg: var(--color-background-100); --normal-text: var(--color-text); --normal-border: var(--border); --success-bg: var(--normal-bg); --success-border: var(--normal-border); --info-bg: var(--normal-bg); --info-border: var(--normal-border); --warning-bg: var(--normal-bg); --warning-border: var(--normal-border); --error-bg: var(--normal-bg); --error-border: var(--normal-border);"
/>

{#if ready}
    {@render children?.()}
{/if}
