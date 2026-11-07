<script lang="ts">
    import { page } from "$app/state";
    import * as star from "$lib/star";
    import AuthShell from "@/components/auth-shell.svelte";
    import Button from "@/components/button.svelte";
    import GrantedScopes from "@/components/granted-scopes.svelte";
    import Icon from "@/components/icon.svelte";
    import Input from "@/components/input.svelte";
    import Label from "@/components/ui/label/label.svelte";
    import Notice from "@/components/notice.svelte";
    import { isLoggedIn, requestLogin } from "@/index";
    import { onMount } from "svelte";

    let code = $state<string>(page.url.searchParams.get("user_code") ?? "");
    let error = $state<string | null>(null);
    let loading = $state<boolean>(false);
    let verification = $state<star.DeviceVerification | null>(null);
    let outcome = $state<"approved" | "denied" | null>(null);

    const describeError = (message: string) => {
        if (message === "application_access_denied") {
            return "Your account does not have access to this application";
        }

        return message;
    };

    const signIn = () =>
        requestLogin({
            returnTo: page.url.pathname + page.url.search,
            reason: "authorize",
            label: "your device",
        });

    const verify = async () => {
        error = null;
        loading = true;

        const result = await star.verifyDeviceCode(code);
        loading = false;

        if ("error" in result) {
            error = describeError(result.message);
            return;
        }

        verification = result;
    };

    const decide = async (approve: boolean) => {
        if (!verification) return;

        error = null;
        loading = true;

        const result = await star.decideDeviceAuthorization(
            verification.user_code,
            approve,
        );

        loading = false;

        if ("error" in result) {
            error = describeError(result.message);
            verification = null;
            return;
        }

        outcome = approve ? "approved" : "denied";
    };

    const handleSubmit = (event: Event) => {
        event.preventDefault();
        verify();
    };

    onMount(() => {
        if (!isLoggedIn()) {
            signIn();
            return;
        }

        if (code) verify();
    });
</script>

<AuthShell>
    {#if outcome === "approved"}
        <div class="flex flex-col items-center gap-2 py-4 text-center">
            <Icon icon="check" class="text-3xl text-emerald-300" />
            <h1 class="text-lg">Device connected</h1>
            <p class="text-muted text-sm">
                You can close this page and return to your device.
            </p>
        </div>
    {:else if outcome === "denied"}
        <div class="flex flex-col items-center gap-2 py-4 text-center">
            <Icon icon="x" class="text-3xl text-muted" />
            <h1 class="text-lg">Request denied</h1>
            <p class="text-muted text-sm">
                The device was not given access to your account. You can close
                this page.
            </p>
        </div>
    {:else if verification}
        <div class="flex flex-col gap-1">
            <h1 class="text-lg">Connect {verification.application.name}</h1>
            <p class="text-muted text-sm">
                A device is asking for access to your account. Confirm that this
                code matches the one shown on your device.
            </p>
        </div>

        <p
            class="my-6 rounded-lg border border-background-200 py-3 text-center font-mono text-2xl tracking-widest select-all"
        >
            {verification.user_code}
        </p>

        <Notice icon="alert-triangle" title="Only continue if you started this.">
            If someone sent you this code or link, or the device is not in front
            of you, deny this request. Approving gives that device access to
            your account.
        </Notice>

        <div class="mt-4 flex flex-col gap-1 text-xs text-muted">
            <p class="flex items-center gap-1.5">
                <Icon icon="device-desktop" />
                <span>Reported device type: {verification.device}</span>
            </p>
            {#if verification.geolocation}
                <p class="flex items-center gap-1.5">
                    <Icon icon="world" />
                    <span>Requested from {verification.geolocation}</span>
                </p>
            {/if}
        </div>

        <GrantedScopes scopes={verification.application.scopes} class="my-6" />

        {#if error}
            <p class="mb-4 text-sm text-red-300 break-words">{error}</p>
        {/if}

        <div class="flex flex-col gap-2">
            <Button
                type="button"
                class="w-full"
                {loading}
                onclick={() => decide(true)}
            >
                Allow access
            </Button>
            <Button
                type="button"
                variant="ghost"
                class="w-full"
                disabled={loading}
                onclick={() => decide(false)}
            >
                Deny
            </Button>
        </div>
    {:else}
        <form onsubmit={handleSubmit}>
            <div class="flex flex-col gap-1">
                <h1 class="text-lg">Connect a device</h1>
                <p class="text-muted text-sm">
                    Enter the code shown on your device.
                </p>
            </div>

            <div class="my-6 grid gap-2">
                <Label for="user_code">Device code</Label>
                <Input
                    bind:value={code}
                    id="user_code"
                    type="text"
                    class="font-mono uppercase"
                    placeholder="XXXX-XXXX"
                    autocomplete="off"
                    autocapitalize="characters"
                    spellcheck="false"
                    maxlength={32}
                    required
                />
                {#if error}
                    <p class="text-sm text-red-300 break-words">{error}</p>
                {/if}
            </div>

            <Button type="submit" class="w-full" {loading}>Continue</Button>
        </form>
    {/if}
</AuthShell>
