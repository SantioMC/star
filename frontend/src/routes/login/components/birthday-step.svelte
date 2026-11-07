<script lang="ts">
    import BirthdayField from "@/components/birthday-field.svelte";
    import Button from "@/components/button.svelte";
    import * as star from "@/star";

    const {
        flow,
        onstage,
        onrejected,
        oncancel,
    }: {
        flow: string;
        onstage: (response: star.LoginResponse) => void;
        onrejected: (message: string) => void;
        oncancel: () => void;
    } = $props();

    let birthday = $state("");
    let loading = $state(false);
    let error = $state("");

    const submit = async (event: Event) => {
        event.preventDefault();
        if (loading) return;

        error = "";
        loading = true;
        const response = await star.loginBirthday(flow, birthday);
        loading = false;

        if ("error" in response) {
            if (response.status === 403) return onrejected(response.message);

            error = response.message;
            return;
        }

        onstage(response);
    };
</script>

<form onsubmit={submit}>
    <BirthdayField bind:value={birthday} required />

    {#if error}
        <p role="alert" class="mt-3 text-sm text-red-300">{error}</p>
    {/if}

    <div class="mt-6 flex items-center justify-between gap-2">
        <Button type="button" variant="ghost" onclick={oncancel}>Back</Button>
        <Button type="submit" {loading}>Continue</Button>
    </div>
</form>
