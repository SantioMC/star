<script lang="ts">
    import Card from "@/components/card.svelte";
    import NumberField from "@/components/number-field.svelte";
    import SettingsForm from "@/components/settings-form.svelte";
    import { getSettingsContext } from "../context.svelte";

    const settings = getSettingsContext();

    let codeLength = $state(String(settings.current!.email_code_length));

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
    description="Configure how Star verifies email addresses"
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
                min={8}
                max={12}
            />
        </div>
    </Card>
</SettingsForm>
