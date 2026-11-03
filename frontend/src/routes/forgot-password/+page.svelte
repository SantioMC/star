<script lang="ts">
    import { env } from "$env/dynamic/public";
    import { goto } from "$app/navigation";
    import * as star from "$lib/star";
    import AuthShell from "@/components/auth-shell.svelte";
    import Button from "@/components/button.svelte";
    import Captcha from "@/components/captcha.svelte";
    import Icon from "@/components/icon.svelte";
    import Input from "@/components/input.svelte";
    import { Label } from "@/components/ui/label";

    let email = $state("");
    let loading = $state(false);
    let error = $state("");
    let submitted = $state(false);

    let promptCaptcha = $state(false);
    let captchaError = $state("");

    const submit = async (captcha?: string) => {
        if (loading) return;

        if (env.PUBLIC_SITE_KEY && !captcha) {
            promptCaptcha = true;
            return;
        }

        error = "";
        loading = true;
        const response = await star.requestPasswordReset(email, captcha);

        loading = false;
        promptCaptcha = false;

        if ("error" in response) {
            error = response.message;
            return;
        }

        submitted = true;
    };

    const submitForm = (event: Event) => {
        event.preventDefault();
        submit();
    };
</script>

<AuthShell>
    {#if promptCaptcha}
        <div class="flex flex-col">
            <h1 class="text-lg">Verify you're human</h1>
            <p class="text-muted text-sm">
                Complete the challenge below to continue
            </p>
        </div>
        <div class="mt-6 flex items-center flex-col gap-2">
            <Captcha onsuccess={submit} onerror={(e) => (captchaError = e)} />
            {#if captchaError}
                <span role="alert" class="text-red-300 text-center text-sm">
                    {captchaError}
                </span>
            {/if}
        </div>
        <div class="mt-6 flex justify-start">
            <Button
                type="button"
                variant="ghost"
                onclick={() => (promptCaptcha = false)}
            >
                Back
            </Button>
        </div>
    {:else if submitted}
        <div class="flex flex-col">
            <h1 class="text-lg flex items-center gap-2">
                <Icon icon="mail" />
                Check your email
            </h1>
            <p class="text-muted text-sm mt-1">
                If <span class="text-foreground">{email}</span> is a verified
                email on an account with a password, we sent it a link to reset
                your password. The link expires in 30 minutes.
            </p>
        </div>
        <div class="mt-6 flex">
            <Button class="w-full" onclick={() => goto("/login")}>
                Back to Login
            </Button>
        </div>
    {:else}
        <div class="flex flex-col mb-6">
            <h1 class="text-lg">Reset your password</h1>
            <p class="text-muted text-sm">
                Enter the verified email on your account and we'll send you a
                link to choose a new password
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

        <form onsubmit={submitForm}>
            <div class="grid gap-2">
                <Label for="email" class="select-none">Email</Label>
                <Input
                    bind:value={email}
                    id="email"
                    type="email"
                    autocomplete="email"
                    maxlength={255}
                    required
                />
            </div>
            <div class="mt-6 flex flex-col gap-2">
                <Button type="submit" class="w-full" {loading}>
                    Send reset link
                </Button>
                <div class="flex items-center gap-1 text-sm text-muted select-none">
                    Remembered it?
                    <a
                        href="/login"
                        class="text-primary hover:text-primary/80 transition-colors"
                    >
                        Back to login
                    </a>
                </div>
            </div>
        </form>
    {/if}
</AuthShell>
