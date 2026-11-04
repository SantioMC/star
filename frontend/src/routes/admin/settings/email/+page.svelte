<script lang="ts">
    import { branding } from "@/branding";
    import Card from "@/components/card.svelte";
    import CheckboxCard from "@/components/checkbox-card.svelte";
    import CollapsibleCard from "@/components/collapsible-card.svelte";
    import DisposableDomainList from "@/components/disposable-domain-list.svelte";
    import DomainList from "@/components/domain-list.svelte";
    import Notice from "@/components/notice.svelte";
    import NumberField from "@/components/number-field.svelte";
    import Segmented from "@/components/segmented.svelte";
    import SettingsForm from "@/components/settings-form.svelte";
    import { getSettingsContext } from "../context.svelte";

    const settings = getSettingsContext();

    let codeLength = $state(String(settings.current!.email_code_length));
    let verification = $state(settings.current!.email_verification);
    let domainMode = $state(settings.current!.email_domain_mode);
    let domains = $state(settings.current!.email_domains);
    let blockDisposable = $state(settings.current!.block_disposable_emails);

    const parsedLength = $derived(Number(codeLength));
    const emptyWhitelist = $derived(
        domainMode === "whitelist" && domains.length === 0,
    );

    const dirty = $derived(
        parsedLength !== settings.current!.email_code_length ||
            verification !== settings.current!.email_verification ||
            domainMode !== settings.current!.email_domain_mode ||
            domains.join() !== settings.current!.email_domains.join() ||
            blockDisposable !== settings.current!.block_disposable_emails,
    );

    const issue = $derived.by(() => {
        if (!Number.isInteger(parsedLength)) {
            return "Code length must be a whole number";
        }

        return emptyWhitelist ? "A whitelist needs at least one domain" : null;
    });

    const discard = () => {
        codeLength = String(settings.current!.email_code_length);
        verification = settings.current!.email_verification;
        domainMode = settings.current!.email_domain_mode;
        domains = settings.current!.email_domains;
        blockDisposable = settings.current!.block_disposable_emails;
    };

    const save = async () => {
        const saved = await settings.update({
            email_code_length: parsedLength,
            email_verification: verification,
            email_domain_mode: domainMode,
            email_domains: domains,
            block_disposable_emails: blockDisposable,
        });

        if (saved) discard();
    };
</script>

<SettingsForm
    title="Email Settings"
    description="Configure how {$branding.name} verifies email addresses and which domains may be used"
    {dirty}
    {issue}
    onsave={save}
    ondiscard={discard}
>
    <Card
        title="Email Verification"
        icon="mail"
        description="Confirm that email addresses belong to the people adding them. Longer codes are harder to guess but more tedious to type."
        class="flex flex-col gap-4"
    >
        <CheckboxCard bind:checked={verification} title="Require email verification">
            Sends a code to confirm that new addresses belong to the person
            adding them. Turn this off to accept addresses immediately.
        </CheckboxCard>

        {#if !verification}
            <Notice icon="alert-triangle" title="Email ownership is not checked." tone="danger">
                Anyone can claim an address they don't own, blocking its real
                owner from using it here. These addresses are reported to
                applications as unverified, and password resets still go to the
                address itself.
            </Notice>
        {/if}

        <div class="sm:max-w-1/3">
            <NumberField
                bind:value={codeLength}
                id="emailCodeLength"
                label="Code Length"
                min={6}
                max={12}
            />
        </div>
    </Card>

    <Card
        title="Allowed Domains"
        icon="world"
        description="Restrict which email domains can be used to register or be added to an account. Existing addresses are not affected. Subdomains are included."
        class="flex flex-col gap-4"
    >
        <Segmented
            bind:value={domainMode}
            label="Domain restriction"
            options={[
                { value: "disabled", label: "Off" },
                { value: "whitelist", label: "Whitelist" },
                { value: "blacklist", label: "Blacklist" },
            ]}
        />

        {#if domainMode !== "disabled"}
            <DomainList
                bind:domains
                id="emailDomains"
                label={domainMode === "whitelist" ? "Allowed domains" : "Blocked domains"}
            />
        {/if}

        {#if emptyWhitelist}
            <Notice icon="alert-triangle" title="This whitelist is empty.">
                Add at least one domain, otherwise nobody could add an email
                address.
            </Notice>
        {/if}

        <CheckboxCard bind:checked={blockDisposable} title="Block known disposable domains">
            Rejects addresses from publicly known temporary email services.
            {#if domainMode === "whitelist"}
                Has no effect while a whitelist is active, only whitelisted
                domains are accepted.
            {/if}
        </CheckboxCard>
    </Card>

    {#snippet preview()}
        <CollapsibleCard
            title="Known Disposable Domains"
            icon="list-search"
            description="The publicly maintained list used to block temporary email services. Subdomains are included."
        >
            <DisposableDomainList />
        </CollapsibleCard>
    {/snippet}
</SettingsForm>
