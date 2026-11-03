<script lang="ts">
    import PageHeader from "@/components/page-header.svelte";
    import ChangePasswordCard from "./components/change-password-card.svelte";
    import MultiFactorCard from "./components/multi-factor-card.svelte";
    import SessionsCard from "./components/sessions-card.svelte";
</script>

<PageHeader
    title="Account Security"
    description="Protect and secure your account"
/>

<ChangePasswordCard />
<MultiFactorCard />
<SessionsCard />
