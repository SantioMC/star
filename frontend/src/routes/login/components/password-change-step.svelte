<script lang="ts">
    import Button from "@/components/button.svelte";
    import NewPasswordFields from "@/components/new-password-fields.svelte";
    import * as star from "@/star";

    const {
        flow,
        onstage,
        oncancel,
    }: {
        flow: string;
        onstage: (response: star.LoginResponse) => void;
        oncancel: () => void;
    } = $props();

    let password = $state("");
    let confirmPassword = $state("");
    let loading = $state(false);
    let error = $state("");

    const submit = async (event: Event) => {
        event.preventDefault();
        if (loading) return;

        error = "";

        if (password !== confirmPassword) {
            error = "Your passwords do not match!";
            return;
        }

        loading = true;
        const response = await star.loginPassword(flow, password);
        loading = false;

        if ("error" in response) {
            error = response.message;
            return;
        }

        onstage(response);
    };
</script>

<form onsubmit={submit}>
    <NewPasswordFields bind:password bind:confirmPassword />

    {#if error}
        <p role="alert" class="mt-3 text-sm text-red-300">{error}</p>
    {/if}

    <div class="mt-6 flex items-center justify-between gap-2">
        <Button type="button" variant="ghost" onclick={oncancel}>Back</Button>
        <Button type="submit" {loading}>Change password</Button>
    </div>
</form>
