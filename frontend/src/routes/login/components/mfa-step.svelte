<script lang="ts">
    import Button from "@/components/button.svelte";
    import CodeInput from "@/components/code-input.svelte";
    import Icon from "@/components/icon.svelte";
    import Input from "@/components/input.svelte";
    import { authenticateWithPasskey } from "@/passkey";
    import * as star from "@/star";
    import { tick } from "svelte";

    let {
        flow,
        factors,
        method = $bindable("totp"),
        onstage,
        oncancel,
    }: {
        flow: string;
        factors: star.MFAMethod[];
        method?: star.MFAMethod;
        onstage: (response: star.LoginResponse) => void;
        oncancel: () => void;
    } = $props();

    let container = $state<HTMLDivElement>();
    let code = $state("");
    let attempt = $state(0);
    let loading = $state(false);
    let error = $state("");

    const usingBackupCode = $derived(method === "backup_code");
    const minimumLength = $derived(usingBackupCode ? 10 : 6);
    const alternatives = $derived(
        factors.filter((factor) => factor !== method),
    );

    $effect(() => {
        attempt;
        method;
        tick().then(() => container?.querySelector("input")?.focus());
    });

    const switchMethod = (target: star.MFAMethod) => {
        method = target;
        code = "";
        error = "";
    };

    const verifyPasskey = async () => {
        if (loading) return;

        error = "";
        loading = true;
        const response = await authenticateWithPasskey(flow);
        loading = false;

        if (response === null) return;
        if ("error" in response) {
            error = response.message;
            return;
        }

        onstage(response);
    };

    const verify = async (value: string = code) => {
        if (loading || value.trim().length < minimumLength) return;

        error = "";
        loading = true;
        const response = usingBackupCode
            ? await star.loginBackupCode(flow, value)
            : await star.loginTotp(flow, value);
        loading = false;

        if ("error" in response) {
            error = response.message;
            attempt++;
            return;
        }

        onstage(response);
    };

    const submitBackupCode = (event: Event) => {
        event.preventDefault();
        verify();
    };
</script>

<div bind:this={container} class="flex flex-col items-center gap-3">
    {#if method === "passkey"}
        <div
            class="flex size-16 items-center justify-center rounded-full bg-background-200/80"
        >
            <Icon icon="fingerprint" class="text-3xl" />
        </div>
    {:else if usingBackupCode}
        <form class="w-full" onsubmit={submitBackupCode}>
            {#key attempt}
                <Input
                    bind:value={code}
                    aria-label="Backup code"
                    placeholder="xxxxx-xxxxx"
                    autocomplete="off"
                    autocapitalize="off"
                    spellcheck="false"
                    class="text-center font-mono tracking-widest"
                />
            {/key}
        </form>
    {:else}
        {#key attempt}
            <CodeInput bind:value={code} oncomplete={verify} />
        {/key}
    {/if}

    {#if error}
        <p role="alert" class="text-sm text-red-300 text-center">{error}</p>
    {/if}

    {#each alternatives as alternative (alternative)}
        <button
            type="button"
            class="text-sm text-primary transition-colors hover:text-primary/80"
            onclick={() => switchMethod(alternative)}
        >
            {{
                totp: "Use your authenticator app instead",
                passkey: "Use a passkey instead",
                backup_code: "Use a backup code instead",
            }[alternative]}
        </button>
    {/each}
</div>

<div class="mt-6 flex items-center justify-between gap-2">
    <Button type="button" variant="ghost" onclick={oncancel}>Back</Button>
    {#if method === "passkey"}
        <Button type="button" {loading} onclick={verifyPasskey}>
            Use passkey
        </Button>
    {:else}
        <Button
            type="button"
            {loading}
            disabled={code.trim().length < minimumLength}
            onclick={() => verify()}
        >
            Verify
        </Button>
    {/if}
</div>
