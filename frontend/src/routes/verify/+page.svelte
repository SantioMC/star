<script lang="ts">
    import { browser } from "$app/environment";
    import { goto } from "$app/navigation";
    import { page } from "$app/state";
    import * as star from "$lib/star";
    import Button from "@/components/button.svelte";
    import Card from "@/components/card.svelte";
    import Icon from "@/components/icon.svelte";
    import { isLoggedIn, requestLogin } from "@/index";
    import { onMount } from "svelte";
    import { toast } from "svelte-sonner";

    let loading = $state(true);
    let error = $state<string | null>(null);
    let ownerRequired = $state(false);

    const signIn = () =>
        requestLogin({
            returnTo: page.url.pathname + page.url.search,
        });

    onMount(async () => {
        if (!browser) return;

        const email = page.url.searchParams.get("email");
        const code = page.url.searchParams.get("code");

        if (!email || !code) {
            error = "Invalid verification link provided";
            return;
        }

        const data = await star.verifyEmail(email, code);
        if ("error" in data) {
            ownerRequired = data.status === 403;
            error = data.message;
            return;
        }

        loading = false;

        if (isLoggedIn()) {
            toast.success(`Verified ${email}`);
            goto("/");
            return;
        }

        localStorage.setItem(
            "star-login-note",
            "Your account has been verified, you may now login",
        );

        goto("/login");
    });
</script>

<div class="h-full w-full flex items-center justify-center px-4">
    <Card class="-my-4 w-full sm:w-sm">
        {#if error}
            <div class="flex flex-col">
                <h1 class="text-lg">
                    {ownerRequired
                        ? "Sign in to verify this email"
                        : "Failed to verify your account"}
                </h1>
                <p
                    class={ownerRequired
                        ? "text-muted text-sm mt-1"
                        : "text-red-300 text-sm mt-1"}
                >
                    {error}
                </p>
            </div>
            <div class="mt-6 flex">
                {#if ownerRequired}
                    <Button class="w-full" onclick={signIn}>Sign in</Button>
                {:else}
                    <Button class="w-full" onclick={() => goto("/login")}>
                        Go to Login
                    </Button>
                {/if}
            </div>
        {:else if loading}
            <div class="flex flex-col">
                <h1 class="text-lg">Verifying your account</h1>
                <p class="text-muted text-sm">
                    One moment while we verify your account...
                </p>
            </div>
            <div class="flex items-center justify-center h-40 w-full">
                <Icon icon="loader-2" class="animate-spin text-3xl text-muted" />
            </div>
        {/if}
    </Card>
</div>
