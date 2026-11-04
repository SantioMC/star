<script lang="ts">
    import { branding } from "@/branding";
    import * as star from "$lib/star";
    import Card from "@/components/card.svelte";
    import EmailPreview from "@/components/email-preview.svelte";
    import NumberField from "@/components/number-field.svelte";
    import Segmented from "@/components/segmented.svelte";
    import SettingsForm from "@/components/settings-form.svelte";
    import { getSettingsContext } from "../context.svelte";

    const settings = getSettingsContext();

    let codeLength = $state(String(settings.current!.email_code_length));
    let template = $state<star.EmailTemplate>("verification");

    const parsedLength = $derived(Number(codeLength));
    const dirty = $derived(
        parsedLength !== settings.current!.email_code_length,
    );

    const issue = $derived(
        Number.isInteger(parsedLength) ? null : "Code length must be a whole number",
    );

    const save = () => settings.update({ email_code_length: parsedLength });

    const discard = () => {
        codeLength = String(settings.current!.email_code_length);
    };
</script>

<SettingsForm
    title="Email Settings"
    description="Configure how {$branding.name} verifies email addresses and preview the emails it sends"
    {dirty}
    {issue}
    onsave={save}
    ondiscard={discard}
>
    <Card
        title="Email Verification"
        icon="mail"
        description="Length of the codes sent to verify email addresses. Longer codes are harder to guess but more tedious to type."
    >
        <div class="mt-6 sm:max-w-1/3">
            <NumberField
                bind:value={codeLength}
                id="emailCodeLength"
                label="Code Length"
                min={6}
                max={12}
            />
        </div>
    </Card>

    {#snippet preview()}
        <Card
            title="Email Templates"
            icon="mail-opened"
            description="Example emails with sample data, styled with the current branding."
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

            <EmailPreview {template} class="mt-4 h-[720px]" />
        </Card>
    {/snippet}
</SettingsForm>
