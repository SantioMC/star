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
  import Input from "@/components/input.svelte";
  import { Label } from "@/components/ui/label";
  import Button from "@/components/button.svelte";
  import * as star from "@/star";
  import { toast } from "svelte-sonner";
  import { getSettingsContext } from "../context.svelte";

  const settings = getSettingsContext();

  let codeLength = $state(String(settings.current!.email_code_length));
  let verification = $state(settings.current!.email_verification);
  let domainMode = $state(settings.current!.email_domain_mode);
  let domains = $state(settings.current!.email_domains);
  let blockDisposable = $state(settings.current!.block_disposable_emails);
  let provider = $state(settings.current!.email_provider);
  let sender = $state(settings.current!.email_sender);
  let smtpHost = $state(settings.current!.smtp_host);
  let smtpPort = $state(String(settings.current!.smtp_port));
  let smtpUsername = $state(settings.current!.smtp_username);
  let smtpPassword = $state("");
  let resendApiKey = $state("");

  const parsedLength = $derived(Number(codeLength));
  const parsedPort = $derived(Number(smtpPort));
  const emptyWhitelist = $derived(
    domainMode === "whitelist" && domains.length === 0,
  );

  const dirty = $derived(
    parsedLength !== settings.current!.email_code_length ||
      verification !== settings.current!.email_verification ||
      domainMode !== settings.current!.email_domain_mode ||
      domains.join() !== settings.current!.email_domains.join() ||
      blockDisposable !== settings.current!.block_disposable_emails ||
      provider !== settings.current!.email_provider ||
      sender.trim() !== settings.current!.email_sender ||
      smtpHost.trim() !== settings.current!.smtp_host ||
      parsedPort !== settings.current!.smtp_port ||
      smtpUsername.trim() !== settings.current!.smtp_username ||
      smtpPassword !== "" ||
      resendApiKey.trim() !== "",
  );

  const deliveryIssue = $derived.by(() => {
    if (provider === "disabled") return null;
    if (!sender.trim()) return "Sending emails requires a sender address";

    if (provider === "smtp") {
      if (!smtpHost.trim()) return "SMTP requires a host";
      if (
        !Number.isInteger(parsedPort) ||
        parsedPort < 1 ||
        parsedPort > 65535
      ) {
        return "The SMTP port must be between 1 and 65535";
      }
    }

    if (
      provider === "resend" &&
      !resendApiKey.trim() &&
      !settings.current!.has_resend_api_key
    ) {
      return "Resend requires an API key";
    }

    return null;
  });

  const issue = $derived.by(() => {
    if (!Number.isInteger(parsedLength)) {
      return "Code length must be a whole number";
    }

    if (emptyWhitelist) return "A whitelist needs at least one domain";
    return deliveryIssue;
  });

  const deliveryChanges = $derived.by(() => {
    const changes: star.UpdateServerSettingsRequest = {
      email_provider: provider,
      email_sender: sender.trim(),
      smtp_host: smtpHost.trim(),
      smtp_port: parsedPort,
      smtp_username: smtpUsername.trim(),
    };

    if (smtpPassword) changes.smtp_password = smtpPassword;
    if (resendApiKey.trim()) changes.resend_api_key = resendApiKey.trim();

    return changes;
  });

  let sendingTest = $state(false);

  const sendTest = async () => {
    sendingTest = true;
    const result = await star.testEmailSettings(deliveryChanges);
    sendingTest = false;

    if ("error" in result) {
      toast.error(result.message);
      return;
    }

    toast.success(`Sent a test email to ${result.email}`);
  };

  const secretPlaceholder = (saved: boolean) =>
    saved ? "Saved, leave blank to keep" : "";

  const discard = () => {
    codeLength = String(settings.current!.email_code_length);
    verification = settings.current!.email_verification;
    domainMode = settings.current!.email_domain_mode;
    domains = settings.current!.email_domains;
    blockDisposable = settings.current!.block_disposable_emails;
    provider = settings.current!.email_provider;
    sender = settings.current!.email_sender;
    smtpHost = settings.current!.smtp_host;
    smtpPort = String(settings.current!.smtp_port);
    smtpUsername = settings.current!.smtp_username;
    smtpPassword = "";
    resendApiKey = "";
  };

  const save = async () => {
    const saved = await settings.update({
      email_code_length: parsedLength,
      email_verification: verification,
      email_domain_mode: domainMode,
      email_domains: domains,
      block_disposable_emails: blockDisposable,
      ...deliveryChanges,
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
    title="Delivery"
    icon="send"
    description="Choose how {$branding.name} sends verification codes, password resets, invites and security notices."
    class="flex flex-col gap-4"
  >
    <Segmented
      bind:value={provider}
      label="Email provider"
      options={[
        { value: "disabled", label: "Off" },
        { value: "smtp", label: "SMTP" },
        { value: "resend", label: "Resend" },
      ]}
    />

    {#if provider === "disabled"}
      <Notice icon="alert-triangle" title="Emails are not sent.">
        Messages are written to the server log instead. Users can't receive
        verification codes, password resets or invites until a provider is
        configured.
      </Notice>
    {:else}
      <div class="grid gap-2">
        <Label for="emailSender" class="select-none">Sender</Label>
        <Input
          bind:value={sender}
          id="emailSender"
          placeholder="{$branding.name} <noreply@example.com>"
          autocomplete="off"
        />
      </div>
    {/if}

    {#if provider === "smtp"}
      <div class="grid gap-4 sm:grid-cols-[1fr_8rem]">
        <div class="grid gap-2">
          <Label for="smtpHost" class="select-none">Host</Label>
          <Input
            bind:value={smtpHost}
            id="smtpHost"
            placeholder="smtp.example.com"
            autocomplete="off"
          />
        </div>
        <NumberField
          bind:value={smtpPort}
          id="smtpPort"
          label="Port"
          min={1}
          max={65535}
        />
      </div>
      <div class="grid gap-4 sm:grid-cols-2">
        <div class="grid gap-2">
          <Label for="smtpUsername" class="select-none">Username</Label>
          <Input
            bind:value={smtpUsername}
            id="smtpUsername"
            autocomplete="off"
          />
        </div>
        <div class="grid gap-2">
          <Label for="smtpPassword" class="select-none">Password</Label>
          <Input
            bind:value={smtpPassword}
            id="smtpPassword"
            type="password"
            autocomplete="new-password"
            placeholder={secretPlaceholder(settings.current!.has_smtp_password)}
          />
        </div>
      </div>
    {:else if provider === "resend"}
      <div class="grid gap-2">
        <Label for="resendApiKey" class="select-none">API Key</Label>
        <Input
          bind:value={resendApiKey}
          id="resendApiKey"
          type="password"
          autocomplete="new-password"
          placeholder={secretPlaceholder(settings.current!.has_resend_api_key)}
        />
      </div>
    {/if}

    {#if provider !== "disabled"}
      <div class="flex flex-wrap items-center gap-3">
        <Button
          type="button"
          variant="secondary"
          loading={sendingTest}
          disabled={sendingTest || deliveryIssue !== null}
          onclick={sendTest}
        >
          Send test email
        </Button>
        <span class="text-sm text-muted select-none">
          Sends to your primary address using the settings above, even before
          they are saved
        </span>
      </div>
    {/if}
  </Card>

  <Card
    title="Email Verification"
    icon="mail"
    description="Confirm that email addresses belong to the people adding them. Longer codes are harder to guess but more tedious to type."
    class="flex flex-col gap-4"
  >
    <CheckboxCard
      bind:checked={verification}
      title="Require email verification"
    >
      Sends a code to confirm that new addresses belong to the person adding
      them. Turn this off to accept addresses immediately.
    </CheckboxCard>

    {#if !verification}
      <Notice
        icon="alert-triangle"
        title="Email ownership is not checked."
        tone="danger"
      >
        Anyone can claim an address they don't own, blocking its real owner from
        using it here. These addresses are reported to applications as
        unverified, and password resets still go to the address itself.
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
        label={domainMode === "whitelist"
          ? "Allowed domains"
          : "Blocked domains"}
      />
    {/if}

    {#if emptyWhitelist}
      <Notice icon="alert-triangle" title="This whitelist is empty.">
        Add at least one domain, otherwise nobody could add an email address.
      </Notice>
    {/if}

    <CheckboxCard
      bind:checked={blockDisposable}
      title="Block known disposable domains"
    >
      Rejects addresses from publicly known temporary email services.
      {#if domainMode === "whitelist"}
        Has no effect while a whitelist is active, only whitelisted domains are
        accepted.
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
