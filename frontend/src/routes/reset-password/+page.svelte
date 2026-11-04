<script lang="ts">
    import { goto, replaceState } from "$app/navigation";
    import { page } from "$app/state";
    import * as star from "$lib/star";
    import AuthShell from "@/components/auth-shell.svelte";
    import Button from "@/components/button.svelte";
    import Icon from "@/components/icon.svelte";
    import NewPasswordFields from "@/components/new-password-fields.svelte";
    import { onMount } from "svelte";

    let token = $state("");
    let verifying = $state(true);
    let invalid = $state<string | null>(null);

    let password = $state("");
    let confirmPassword = $state("");
    let loading = $state(false);
    let error = $state("");

    onMount(async () => {
        token = page.url.searchParams.get("token") ?? "";
        replaceState("/reset-password", {});

        if (!token) {
            invalid = "This password reset link is invalid, please request a new one";
            verifying = false;
            return;
        }

        const response = await star.verifyPasswordReset(token);
        verifying = false;

        if ("error" in response) {
            invalid = response.message;
        }
    });

    const submit = async (event: Event) => {
        event.preventDefault();
        if (loading) return;

        error = "";

        if (password !== confirmPassword) {
            error = "Your passwords do not match!";
            return;
        }

        loading = true;
        const response = await star.completePasswordReset(token, password);
        loading = false;

        if ("error" in response) {
            if (response.status === 403 || response.status === 410) {
                invalid = response.message;
                return;
            }

            error = response.message;
            return;
        }

        localStorage.setItem(
            "star-login-note",
            "Your password has been reset and all sessions were signed out, you may now login",
        );

        goto("/login");
    };
</script>

<svelte:head>
    <meta name="referrer" content="no-referrer" />
</svelte:head>

<AuthShell>
    {#if verifying}
        <div class="flex flex-col">
            <h1 class="text-lg">Checking your link</h1>
            <p class="text-muted text-sm">
                One moment while we verify your password reset link...
            </p>
        </div>
        <div class="flex items-center justify-center h-40 w-full">
            <Icon icon="loader-2" class="animate-spin text-3xl text-muted" />
        </div>
    {:else if invalid}
        <div class="flex flex-col">
            <h1 class="text-lg">Unable to reset your password</h1>
            <p class="text-red-300 text-sm mt-1">{invalid}</p>
        </div>
        <div class="mt-6 flex flex-col gap-2">
            <Button class="w-full" onclick={() => goto("/forgot-password")}>
                Request a new link
            </Button>
            <Button
                class="w-full"
                variant="secondary"
                onclick={() => goto("/login")}
            >
                Go to Login
            </Button>
        </div>
    {:else}
        <div class="flex flex-col mb-6">
            <h1 class="text-lg">Choose a new password</h1>
            <p class="text-muted text-sm">
                Resetting your password signs you out of every device
            </p>
        </div>

        {#if error}
            <div
                role="alert"
                class="mb-6 flex items-start gap-2 rounded-lg border border-red-400/30 bg-red-400/5 px-3 py-2 text-sm text-red-300"
            >
                <Icon icon="alert-triangle" class="mt-0.5 shrink-0" />
                <span>{error}</span>
            </div>
        {/if}

        <form onsubmit={submit}>
            <NewPasswordFields bind:password bind:confirmPassword />
            <div class="mt-6">
                <Button type="submit" class="w-full" {loading}>
                    Reset Password
                </Button>
            </div>
        </form>
    {/if}
</AuthShell>
