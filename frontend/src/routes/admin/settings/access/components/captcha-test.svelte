<script lang="ts">
    import Badge from "@/components/badge.svelte";
    import Button from "@/components/button.svelte";
    import Captcha from "@/components/captcha.svelte";
    import Notice from "@/components/notice.svelte";
    import * as star from "@/star";

    let {
        changes,
        sitekey,
        verified = $bindable(null),
    }: {
        changes: star.UpdateServerSettingsRequest;
        sitekey: string;
        verified?: string | null;
    } = $props();

    const fingerprint = $derived(JSON.stringify(changes));
    const passed = $derived(verified === fingerprint);

    let solving = $state(false);
    let checking = $state(false);
    let error = $state<string | null>(null);

    const start = () => {
        error = null;
        solving = true;
    };

    const verify = async (token: string) => {
        const tested = fingerprint;

        solving = false;
        checking = true;
        const result = await star.testCaptchaSettings(changes, token);
        checking = false;

        if ("error" in result) {
            error = result.message;
            return;
        }

        verified = tested;
    };

    const fail = (message: string) => {
        solving = false;
        error = message;
    };

    $effect(() => {
        void fingerprint;
        solving = false;
        error = null;
    });
</script>

<div class="flex flex-col gap-3">
    <div class="flex flex-wrap items-center gap-3">
        <Button
            type="button"
            variant="secondary"
            loading={checking}
            disabled={solving || checking || !sitekey}
            onclick={start}
        >
            Test captcha
        </Button>

        {#if passed}
            <Badge tone="success">passed</Badge>
        {:else}
            <span class="text-sm text-muted select-none">
                Solves a captcha with these keys and verifies it with
                Cloudflare
            </span>
        {/if}
    </div>

    {#if solving}
        {#key fingerprint}
            <Captcha {sitekey} onsuccess={verify} onerror={fail} />
        {/key}
    {/if}

    {#if error}
        <Notice icon="alert-triangle" title="The captcha test failed." tone="danger">
            {error}
        </Notice>
    {/if}
</div>
