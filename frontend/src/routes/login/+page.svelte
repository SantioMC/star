<script lang="ts">
    import { env } from "$env/dynamic/public";
    import { afterNavigate, replaceState } from "$app/navigation";
    import { page } from "$app/state";
    import * as star from "$lib/star";
    import Card from "@/components/card.svelte";
    import Icon from "@/components/icon.svelte";
    import {
        continueFlow,
        getLoginIntent,
        startSession,
        user,
    } from "@/index";
    import { authenticateWithPasskey } from "@/passkey";
    import {
        browserSupportsWebAuthn,
        browserSupportsWebAuthnAutofill,
        WebAuthnAbortService,
    } from "@simplewebauthn/browser";
    import { onMount, untrack } from "svelte";
    import CaptchaStep from "./components/captcha-step.svelte";
    import CredentialsStep from "./components/credentials-step.svelte";
    import MfaStep from "./components/mfa-step.svelte";

    type Step =
        | { kind: "credentials" }
        | { kind: "captcha" }
        | { kind: "mfa"; flow: string; factors: star.MFAMethod[] };

    const intent = getLoginIntent();
    const knownUsername = $user?.username;
    const reauthenticating = intent?.reason === "reauthenticate" && !!knownUsername;

    let step = $state<Step>({ kind: "credentials" });
    let notice = $state<string | null>(null);
    let error = $state("");
    let loading = $state(false);
    let mfaMethod = $state<star.MFAMethod>("totp");
    let providers = $state<star.LoginProvider[]>([]);

    let username = $state(reauthenticating ? knownUsername! : "");
    let password = $state("");

    const passkeysEnabled = !reauthenticating && browserSupportsWebAuthn();

    const heading = $derived.by(() => {
        switch (step.kind) {
            case "captcha":
                return {
                    title: "Verify you're human",
                    description: "Complete the challenge below to continue signing in",
                };
            case "mfa":
                return {
                    title: "Two-factor authentication",
                    description: {
                        totp: "Enter the 6-digit code from your authenticator app to finish signing in",
                        passkey: "Use one of your passkeys to finish signing in",
                        backup_code:
                            "Enter one of the backup codes you saved when setting up two-factor authentication",
                    }[mfaMethod],
                };
        }

        if (reauthenticating) {
            return {
                title: "Confirm it's you",
                description: "For your security, enter your password to continue",
            };
        }

        if (intent?.reason === "authorize" && intent.label) {
            return {
                title: `Sign in to ${intent.label}`,
                description: `Use your Star account to continue to ${intent.label}`,
            };
        }

        return {
            title: "Login to your account",
            description: "Enter your account details below to login to your account",
        };
    });

    onMount(async () => {
        notice = localStorage.getItem("star-login-note");
        if (notice) localStorage.removeItem("star-login-note");

        const flow = page.url.searchParams.get("flow");
        error = page.url.searchParams.get("error") ?? "";

        if (!reauthenticating) {
            star.getLoginProviders().then((value) => (providers = value));
        }

        if (!flow) return;

        loading = true;
        const response = await star.completeProviderLogin(flow);
        loading = false;

        if ("error" in response) {
            return restart(response.message);
        }

        handleStage(response);
    });

    afterNavigate(() => {
        if (page.url.search) replaceState("/login", {});
    });

    $effect(() => {
        if (step.kind !== "credentials" || !passkeysEnabled) return;

        untrack(() => passkeyLogin(true));
        return () => WebAuthnAbortService.cancelCeremony();
    });

    const restart = (message: string = "") => {
        error = message;
        password = "";
        step = { kind: "credentials" };
    };

    const complete = (response: star.LoginSuccessResponse) => {
        user.set(null);
        startSession(response.access_token);

        continueFlow();
    };

    const handleStage = (response: star.LoginResponse) => {
        switch (response.next) {
            case "complete":
                return complete(response);
            case "mfa_required":
                if (response.factors.includes("passkey")) {
                    mfaMethod = "passkey";
                } else if (response.factors.includes("totp")) {
                    mfaMethod = "totp";
                } else {
                    return restart(
                        "Your account requires an authentication factor that is not supported yet",
                    );
                }

                step = {
                    kind: "mfa",
                    flow: response.flow,
                    factors: response.factors,
                };
                return;
            default:
                return restart(
                    "Your account requires a sign in step that is not supported yet",
                );
        }
    };

    const login = async (captcha?: string) => {
        if (loading) return;

        if (env.PUBLIC_SITE_KEY && !captcha) {
            step = { kind: "captcha" };
            return;
        }

        error = "";
        loading = true;
        const response = await star.login({ username, password, captcha });
        loading = false;

        if ("error" in response) {
            return restart(response.message);
        }

        handleStage(response);
    };

    const passkeyLogin = async (autofill: boolean = false) => {
        if (autofill && !(await browserSupportsWebAuthnAutofill())) return;
        if (!autofill) {
            if (loading) return;

            error = "";
            loading = true;
        }

        const response = await authenticateWithPasskey(undefined, autofill);
        if (!autofill) loading = false;

        if (response === null) {
            if (!autofill) passkeyLogin(true);
            return;
        }

        if ("error" in response) {
            error = response.message;
            return;
        }

        handleStage(response);
    };

    const submitCredentials = (event: Event) => {
        event.preventDefault();
        login();
    };
</script>

<div class="h-full w-full flex items-center justify-center px-4">
    <div class="w-full sm:w-sm flex flex-col gap-4 items-center">
        {#if notice}
            <div
                class="mb-4 w-full flex items-center gap-2 rounded-lg border border-blue-400/30 bg-blue-500/10 px-4 py-3 text-sm text-blue-200"
            >
                <Icon icon="alert-circle" class="text-lg shrink-0" />
                <span>{notice}</span>
            </div>
        {/if}

        <Card class="-my-4 w-full">
            <div class="flex flex-col mb-6">
                <h1 class="text-lg">{heading.title}</h1>
                <p class="text-muted text-sm">{heading.description}</p>
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

            {#if step.kind === "credentials"}
                <form onsubmit={submitCredentials}>
                    <CredentialsStep
                        bind:username
                        bind:password
                        {loading}
                        lockUsername={reauthenticating}
                        allowRegistration={!reauthenticating}
                        submitLabel={reauthenticating ? "Confirm" : "Login"}
                        {providers}
                        onpasskey={passkeysEnabled
                            ? () => passkeyLogin()
                            : undefined}
                    />
                </form>
            {:else if step.kind === "captcha"}
                <CaptchaStep onsolve={login} oncancel={() => restart()} />
            {:else if step.kind === "mfa"}
                <MfaStep
                    bind:method={mfaMethod}
                    flow={step.flow}
                    factors={step.factors}
                    onstage={handleStage}
                    oncancel={() => restart()}
                />
            {/if}
        </Card>

        {#if env.PUBLIC_TOS || env.PUBLIC_PRIVACY}
            <div
                class="flex w-full justify-center gap-4 mt-8 items-center text-sm text-primary/70"
            >
                {#if env.PUBLIC_TOS}
                    <a
                        href={env.PUBLIC_TOS}
                        class="transition-colors hover:text-primary"
                    >
                        Terms of Service
                    </a>
                {/if}

                {#if env.PUBLIC_TOS && env.PUBLIC_PRIVACY}
                    <span class="h-4 w-px bg-background-200"></span>
                {/if}

                {#if env.PUBLIC_PRIVACY}
                    <a
                        href={env.PUBLIC_PRIVACY}
                        class="transition-colors hover:text-primary"
                    >
                        Privacy Policy
                    </a>
                {/if}
            </div>
        {/if}
    </div>
</div>
