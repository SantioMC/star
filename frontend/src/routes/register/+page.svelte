<script lang="ts">
    import { goto } from "$app/navigation";
    import { env } from "$env/dynamic/public";
    import * as star from "$lib/star";
    import Button from "@/components/button.svelte";
    import Captcha from "@/components/captcha.svelte";
    import AuthShell from "@/components/auth-shell.svelte";
    import Icon from "@/components/icon.svelte";
    import Input from "@/components/input.svelte";
    import { Label } from "@/components/ui/label";
    import { validateUsername, type UsernameRules } from "@/username";
    import { onMount } from "svelte";
    import { toast } from "svelte-sonner";

    let registered = $state(false);
    let registeredEmail = $state("");

    let loading = $state(false);
    let promptCaptcha = $state(false);
    let error = $state("");

    let countdown = $state<number>(0);
    let countdownInterval = $state<number | undefined>();

    let email = $state("");
    let username = $state("");
    let password = $state("");
    let confirmPassword = $state("");

    let captchaToken = $state("");
    let captchaError = $state("");

    let usernameRules = $state<UsernameRules | null>(null);

    const usernameIssue = $derived(
        usernameRules && username ? validateUsername(username, usernameRules) : null,
    );

    onMount(async () => {
        const requirements = await star.getRegisterRequirements();
        usernameRules = requirements?.username ?? null;
    });

    const register = async (event: Event | null = null) => {
        event?.preventDefault();
        if (usernameIssue) return;

        if (env.PUBLIC_SITE_KEY && !captchaToken) {
            promptCaptcha = true;
            return;
        }

        error = "";

        if (password !== confirmPassword) {
            error = "Your passwords do not match!";
            return;
        }

        loading = true;

        const data = await star.register({
            captcha: captchaToken ? captchaToken : undefined,
            username,
            email,
            password,
        });

        loading = false;

        if ("error" in data) {
            captchaToken = "";
            error = data.message;
            return;
        }

        if (!data.verification_required) {
            toast.success("Your account has been created, you may now login");
            await goto("/login");
            return;
        }

        resetCountdown();
        registeredEmail = email;
        registered = true;
    };

    const resetCountdown = () => {
        countdown = 60;

        if (countdownInterval) {
            clearInterval(countdownInterval);
        }

        countdownInterval = setInterval(() => {
            if (--countdown <= 0) {
                clearInterval(countdownInterval);
                countdownInterval = undefined;
            }
        }, 1000);
    };

    const captchaCallback = async (token: string) => {
        captchaToken = token;
        await register();
        promptCaptcha = false;
    };

    const resendEmail = async () => {
        if (!registeredEmail) return;

        loading = true;
        const data = await star.resendVerification(registeredEmail);

        resetCountdown();
        loading = false;

        if ("error" in data) {
            toast.error(data.message);
            return;
        }

        if (!data.success) {
            toast.error("Failed to resend the verification email");
            return;
        }

        toast.success("The email has been resent");
    };
</script>

<AuthShell>
    <form onsubmit={register}>
        {#if promptCaptcha}
            <div class="flex flex-col">
                <h1 class="text-lg">Solve the Captcha</h1>
                <p class="text-muted text-sm">
                    Please wait or manually solve the provided captcha to
                    prove you are a human
                </p>
            </div>
            <div class="mt-6 flex items-center flex-col gap-2">
                <Captcha
                    onsuccess={captchaCallback}
                    onerror={(e) => (captchaError = e)}
                />
                {#if captchaError}
                    <span class="text-red-300 text-center text-sm">
                        {captchaError}
                    </span>
                {/if}
            </div>
        {:else if registered}
            <div class="flex flex-col">
                <h1 class="text-lg flex items-center gap-2">
                    <Icon icon="mail" />
                    Verify your email
                </h1>
                <p class="text-muted text-sm mt-1">
                    We sent a verification email to
                    <span class="text-foreground">{registeredEmail}</span>.
                    Please check your inbox and spam folder, it may take a
                    minute to arrive.
                </p>
            </div>
            <div class="mt-6 flex justify-end">
                <Button
                    type="button"
                    variant="secondary"
                    onclick={resendEmail}
                    {loading}
                    disabled={countdown > 0}
                >
                    {#if countdown > 0}
                        Resend email ({countdown}s)
                    {:else}
                        Resend email
                    {/if}
                </Button>
            </div>
        {:else}
            <div class="flex flex-col">
                <h1 class="text-lg">Create an account</h1>
                <p class="text-muted text-sm">
                    Enter your account details below to create your account
                </p>
            </div>
            <div class="my-6">
                <div class="flex flex-col gap-6">
                    <div class="grid gap-2">
                        <Label for="username" class="select-none">
                            Username
                        </Label>
                        <Input
                            bind:value={username}
                            id="username"
                            type="text"
                            autocomplete="username"
                            aria-invalid={usernameIssue != null}
                            aria-describedby={usernameIssue
                                ? "username-issue"
                                : undefined}
                            required
                        />
                        {#if usernameIssue}
                            <p
                                id="username-issue"
                                class="text-xs text-red-300"
                            >
                                {usernameIssue}
                            </p>
                        {/if}
                    </div>
                    <div class="grid gap-2">
                        <Label for="email" class="select-none">Email</Label>
                        <Input
                            bind:value={email}
                            id="email"
                            type="email"
                            autocomplete="email"
                            required
                        />
                    </div>
                    <div class="grid gap-2">
                        <Label for="password" class="select-none">
                            Password
                        </Label>
                        <Input
                            bind:value={password}
                            id="password"
                            type="password"
                            autocomplete="new-password"
                            required
                        />
                    </div>
                    <div class="grid gap-2">
                        <Label for="confirm-password" class="select-none">
                            Confirm Password
                        </Label>
                        <Input
                            bind:value={confirmPassword}
                            id="confirm-password"
                            type="password"
                            autocomplete="new-password"
                            required
                        />
                    </div>
                </div>

                {#if error}
                    <span class="text-red-300 mt-2 block">{error}</span>
                {/if}
            </div>
            <div class="flex-col gap-2 flex">
                <Button type="submit" class="w-full" {loading}>
                    Create Account
                </Button>

                <div class="flex items-center gap-1 text-sm text-muted">
                    Already have an account?
                    <a
                        href="/login"
                        class="text-primary hover:text-primary/80 transition-colors"
                    >
                        Log in
                    </a>
                </div>
            </div>
        {/if}
    </form>
</AuthShell>
