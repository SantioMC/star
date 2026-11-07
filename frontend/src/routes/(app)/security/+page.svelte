<script lang="ts">
    import PageHeader from "@/components/page-header.svelte";
    import ChangePasswordCard from "./components/change-password-card.svelte";
    import MultiFactorCard from "./components/multi-factor-card.svelte";
    import SessionsCard from "./components/sessions-card.svelte";

    let passwordChanges = $state(0);
</script>

<PageHeader
    title="Account Security"
    description="Protect and secure your account"
/>

<ChangePasswordCard onchanged={() => passwordChanges++} />
<MultiFactorCard />
{#key passwordChanges}
    <SessionsCard />
{/key}
