<script lang="ts">
    import Button from "@/components/button.svelte";
    import Captcha from "@/components/captcha.svelte";

    const {
        onsolve,
        oncancel,
    }: {
        onsolve: (token: string) => void;
        oncancel: () => void;
    } = $props();

    let error = $state("");
</script>

<div class="flex flex-col items-center gap-3">
    <Captcha onsuccess={onsolve} onerror={(message) => (error = message)} />

    {#if error}
        <p role="alert" class="text-sm text-red-300 text-center">{error}</p>
    {/if}
</div>

<div class="mt-6 flex justify-start">
    <Button type="button" variant="ghost" onclick={oncancel}>Back</Button>
</div>
