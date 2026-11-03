<script lang="ts">
    import { page } from "$app/state";
    import * as star from "$lib/star";
    import Button from "@/components/button.svelte";
    import AuthShell from "@/components/auth-shell.svelte";
    import GrantedScopes from "@/components/granted-scopes.svelte";
    import Icon from "@/components/icon.svelte";
    import { isLoggedIn, requestLogin } from "@/index";
    import { onMount } from "svelte";

    let error = $state<string | null>(null);
    let loading = $state<boolean>(true);
    let loggingIn = $state<boolean>(false);
    let application = $state<star.PublicApplication | null>(null);

    const host = $derived.by(() => {
        if (!application) return null;
        if (!URL.canParse(application.url)) return application.url;

        return new URL(application.url).host;
    });

    const prompted = (value: string) => {
        const prompt = page.url.searchParams.get("prompt") ?? "";
        return prompt.split(" ").includes(value);
    };

    onMount(async () => {
        error = page.url.searchParams.get("error");
        if (error) return;

        const clientId = page.url.searchParams.get("client_id");
        const redirectUri = page.url.searchParams.get("redirect_uri");

        if (!clientId || !redirectUri) {
            return fail("Missing client_id or redirect_uri parameter");
        }

        const scope = page.url.searchParams.get("scope") ?? "openid";
        const result = await star.getOAuthApplication(
            clientId,
            redirectUri,
            scope,
        );

        if ("error" in result) {
            return fail(result.message);
        }

        application = result;

        if (!isLoggedIn()) {
            if (prompted("none")) return redirect({ error: "login_required" });
            return signIn();
        }

        if (prompted("consent")) {
            loading = false;
            return;
        }

        submit(false);
    });

    const signIn = () =>
        requestLogin({
            returnTo: page.url.pathname + page.url.search,
            reason: "authorize",
            label: application?.name,
        });

    const fail = (message: string) => {
        error = message;
        loading = false;
        loggingIn = false;
    };

    const redirect = (values: Record<string, string>) => {
        const redirectUri = page.url.searchParams.get("redirect_uri");
        if (!application || !redirectUri) {
            return fail(
                "The redirect_uri could not be verified for this application",
            );
        }

        const url = new URL(redirectUri);
        if (["javascript:", "data:", "vbscript:"].includes(url.protocol)) {
            return fail("The redirect_uri uses a scheme that is not allowed");
        }

        for (const [key, value] of Object.entries(values)) {
            url.searchParams.set(key, value);
        }

        const state = page.url.searchParams.get("state");
        if (state) {
            url.searchParams.set("state", state);
        }

        window.location.href = url.toString();
    };

    const handleSubmit = (event: Event) => {
        event.preventDefault();

        if (!!error || !application) {
            history.back();
            return;
        }

        submit(true);
    };

    const deny = () => redirect({ error: "access_denied" });

    const submit = async (consent: boolean) => {
        if (!application) return;

        loggingIn = true;

        const redirectUri = page.url.searchParams.get("redirect_uri");
        const scope = page.url.searchParams.get("scope") ?? "openid";

        if (!redirectUri) return fail("Missing redirect_uri parameter");

        const authCode = await star.getOAuthAuthCode(
            application.id,
            scope,
            redirectUri,
            page.url.searchParams.get("state"),
            page.url.searchParams.get("code_challenge"),
            page.url.searchParams.get("code_challenge_method"),
            page.url.searchParams.get("nonce"),
            consent,
        );

        if ("error" in authCode) {
            if (authCode.message === "Forbidden") {
                return signIn();
            }

            if (authCode.message === "consent_required") {
                if (prompted("none")) {
                    return redirect({ error: "consent_required" });
                }

                loading = false;
                loggingIn = false;
                return;
            }

            return fail(authCode.message);
        }

        redirect({ code: authCode.code });
    };
</script>

{#snippet spinner()}
    <div class="flex items-center justify-center h-40 w-full">
        <Icon icon="loader-2" class="animate-spin text-3xl text-muted" />
    </div>
{/snippet}

<AuthShell>
    <form onsubmit={handleSubmit}>
        {#if error}
            <div class="flex flex-col">
                <h1 class="text-lg">An error has occurred</h1>
                <p class="text-muted text-sm">
                    Failed to authorize your request, please try again or
                    contact the site administrator if the problem persists.
                </p>
            </div>
            <p
                class="my-6 rounded-lg border border-red-400/30 bg-red-400/5 px-3 py-2 text-sm text-red-300 break-words"
            >
                {error}
            </p>
            <Button type="submit" class="w-full">Go back</Button>
        {:else if loading}
            <div class="flex flex-col">
                <h1 class="text-lg">One moment...</h1>
                <p class="text-muted text-sm">
                    Loading application information, please wait...
                </p>
            </div>
            {@render spinner()}
        {:else if !application}
            <div class="flex flex-col">
                <h1 class="text-lg">Failed to get application</h1>
                <p class="text-muted text-sm">
                    Failed to get the application information, please try
                    again or contact the site administrator if the problem
                    persists.
                </p>
            </div>
            <Button type="submit" class="w-full mt-6">Go back</Button>
        {:else if loggingIn}
            <div class="flex flex-col">
                <h1 class="text-lg">One moment...</h1>
                <p class="text-muted text-sm">
                    You are being logged in, please wait...
                </p>
            </div>
            {@render spinner()}
        {:else}
            <div class="flex flex-col gap-1">
                <h1 class="text-lg">Authorize {application.name}</h1>
                <p class="text-muted text-sm">
                    <span class="text-foreground">{application.name}</span>
                    is asking for access to your account, review what it will
                    be able to do before continuing.
                </p>
                {#if host}
                    <p
                        class="mt-1 flex items-center gap-1.5 text-xs text-muted"
                    >
                        <Icon icon="world" />
                        <span class="font-mono truncate">{host}</span>
                    </p>
                {/if}
            </div>

            <GrantedScopes scopes={application.scopes} class="my-6" />

            <div class="flex-col gap-2 flex">
                <Button type="submit" class="w-full">Allow access</Button>
                <Button
                    type="button"
                    variant="ghost"
                    class="w-full"
                    onclick={deny}
                >
                    Cancel
                </Button>
            </div>

            <p class="mt-4 text-center text-xs text-muted select-none">
                You can revoke this access at any time from your
                <a href="/applications" class="text-foreground underline">
                    authorized applications
                </a>.
            </p>
        {/if}
    </form>
</AuthShell>
