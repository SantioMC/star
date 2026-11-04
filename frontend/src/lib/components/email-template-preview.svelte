<script lang="ts">
    import * as star from "$lib/star";
    import CollapsibleCard from "./collapsible-card.svelte";
    import EmailPreview from "./email-preview.svelte";
    import Segmented from "./segmented.svelte";

    const {
        branding,
        removeLogo = false,
    }: {
        branding: star.UpdateBrandingRequest;
        removeLogo?: boolean;
    } = $props();

    let template = $state<star.EmailTemplate>("verification");
</script>

<CollapsibleCard
    title="Email Templates"
    icon="mail-opened"
    description="Example emails with your unsaved changes applied. Newly uploaded logos appear once saved."
>
    <div class="mt-6">
        <Segmented
            bind:value={template}
            label="Email template"
            options={[
                { value: "verification", label: "Verification" },
                { value: "password-reset", label: "Password Reset" },
                { value: "password-changed", label: "Password Changed" },
            ]}
        />
    </div>

    <EmailPreview {template} {branding} {removeLogo} class="mt-4 h-[680px]" />
</CollapsibleCard>
