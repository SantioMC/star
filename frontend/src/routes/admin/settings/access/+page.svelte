<script lang="ts">
  import { branding } from "@/branding";
  import Card from "@/components/card.svelte";
  import CheckboxCard from "@/components/checkbox-card.svelte";
  import Notice from "@/components/notice.svelte";
  import Input from "@/components/input.svelte";
  import NumberField from "@/components/number-field.svelte";
  import Segmented from "@/components/segmented.svelte";
  import SettingsForm from "@/components/settings-form.svelte";
  import { Label } from "@/components/ui/label";
  import type { UpdateServerSettingsRequest } from "@/star";
  import { getSettingsContext } from "../context.svelte";
  import CaptchaTest from "./components/captcha-test.svelte";

  const settings = getSettingsContext();

  let allowRegistration = $state(settings.current!.allow_registration);
  let allowLogins = $state(settings.current!.allow_logins);
  let trackBirthdays = $state(settings.current!.track_birthdays);
  let minimumAge = $state(String(settings.current!.minimum_age));
  let captchaProvider = $state(settings.current!.captcha_provider);
  let turnstileSiteKey = $state(settings.current!.turnstile_site_key);
  let turnstileSecretKey = $state("");
  let verifiedCaptcha = $state<string | null>(null);

  const effectiveMinimumAge = $derived(trackBirthdays ? Number(minimumAge) : 0);

  const dirty = $derived(
    allowRegistration !== settings.current!.allow_registration ||
      allowLogins !== settings.current!.allow_logins ||
      trackBirthdays !== settings.current!.track_birthdays ||
      effectiveMinimumAge !== settings.current!.minimum_age ||
      captchaProvider !== settings.current!.captcha_provider ||
      turnstileSiteKey.trim() !== settings.current!.turnstile_site_key ||
      turnstileSecretKey.trim() !== "",
  );

  const captchaChanges = $derived.by(() => {
    const changes: UpdateServerSettingsRequest = {
      captcha_provider: captchaProvider,
      turnstile_site_key: turnstileSiteKey.trim(),
    };

    if (turnstileSecretKey.trim()) {
      changes.turnstile_secret_key = turnstileSecretKey.trim();
    }

    return changes;
  });

  const captchaChanged = $derived(
    captchaProvider !== settings.current!.captcha_provider ||
      turnstileSiteKey.trim() !== settings.current!.turnstile_site_key ||
      turnstileSecretKey.trim() !== "",
  );

  const captchaUntested = $derived(
    captchaProvider !== "disabled" &&
      captchaChanged &&
      verifiedCaptcha !== JSON.stringify(captchaChanges),
  );

  const issue = $derived.by(() => {
    if (captchaProvider !== "turnstile") return null;

    const hasSecret =
      turnstileSecretKey.trim() !== "" ||
      settings.current!.has_turnstile_secret_key;

    if (!turnstileSiteKey.trim() || !hasSecret) {
      return "Turnstile requires both a site key and a secret key";
    }

    return captchaUntested ? "Test the captcha before saving" : null;
  });

  const save = async () => {
    const saved = await settings.update({
      allow_registration: allowRegistration,
      allow_logins: allowLogins,
      track_birthdays: trackBirthdays,
      minimum_age: effectiveMinimumAge,
      ...captchaChanges,
    });

    if (saved) {
      turnstileSecretKey = "";
      verifiedCaptcha = null;
    }
  };

  const discard = () => {
    allowRegistration = settings.current!.allow_registration;
    allowLogins = settings.current!.allow_logins;
    trackBirthdays = settings.current!.track_birthdays;
    minimumAge = String(settings.current!.minimum_age);
    captchaProvider = settings.current!.captcha_provider;
    turnstileSiteKey = settings.current!.turnstile_site_key;
    turnstileSecretKey = "";
    verifiedCaptcha = null;
  };
</script>

<SettingsForm
  title="Access Settings"
  description="Control who can sign up and sign in to this {$branding.name} instance"
  {dirty}
  {issue}
  onsave={save}
  ondiscard={discard}
>
  <Card
    title="Registration & Logins"
    icon="door-enter"
    description="Applies to every application authenticating through {$branding.name}, including this dashboard."
    class="flex flex-col gap-4"
  >
    <div class="flex flex-col gap-2">
      <CheckboxCard bind:checked={allowRegistration} title="Allow registration">
        Lets anyone create an account. Turn this off to only allow accounts you
        create yourself.
      </CheckboxCard>

      <CheckboxCard bind:checked={allowLogins} title="Allow logins">
        Lets users sign in. While off, only accounts holding the
        <span class="font-mono text-foreground/80">user:login:bypass</span>
        permission may sign in. Existing sessions stay valid either way.
      </CheckboxCard>
    </div>

    {#if !allowLogins}
      <Notice icon="alert-triangle" title="Logins are disabled.">
        Make sure your own account has the
        <span class="font-mono">user:login:bypass</span>
        permission before signing out.
      </Notice>
    {/if}
  </Card>

  <Card
    title="Birthdays"
    icon="calendar"
    description="Choose whether {$branding.name} collects birthdays"
    class="flex flex-col gap-4"
  >
    <CheckboxCard bind:checked={trackBirthdays} title="Track birthdays">
      Lets users add a birthday to their account and share it with applications.
      While off, none are collected and any already saved stay stored but
      hidden.
    </CheckboxCard>

    {#if trackBirthdays}
      <div class="sm:max-w-1/3">
        <NumberField
          bind:value={minimumAge}
          id="minimumAge"
          label="Minimum Age"
          min={0}
          max={120}
        />
      </div>
    {/if}

    {#if effectiveMinimumAge > 0}
      <Notice icon="alert-triangle" title="Birthdays become required.">
        Users without a birthday are signed out and must enter one on their next
        sign in, and anyone under this age is blocked from signing in.
      </Notice>
    {/if}
  </Card>

  <Card
    title="Bot Protection"
    icon="shield-check"
    description="Require a captcha on sign in, registration and password resets."
    class="flex flex-col gap-4"
  >
    <Segmented
      bind:value={captchaProvider}
      label="Captcha provider"
      options={[
        { value: "disabled", label: "Off" },
        { value: "turnstile", label: "Cloudflare Turnstile" },
      ]}
    />

    {#if captchaProvider === "turnstile"}
      <div class="grid gap-4 sm:grid-cols-2">
        <div class="grid gap-2">
          <Label for="turnstileSiteKey" class="select-none">Site Key</Label>
          <Input
            bind:value={turnstileSiteKey}
            id="turnstileSiteKey"
            class="font-mono"
            autocomplete="off"
          />
        </div>
        <div class="grid gap-2">
          <Label for="turnstileSecretKey" class="select-none">Secret Key</Label>
          <Input
            bind:value={turnstileSecretKey}
            id="turnstileSecretKey"
            type="password"
            autocomplete="new-password"
            placeholder={settings.current!.has_turnstile_secret_key
              ? "Saved, leave blank to keep"
              : ""}
          />
        </div>
      </div>

      <CaptchaTest
        changes={captchaChanges}
        sitekey={turnstileSiteKey.trim()}
        bind:verified={verifiedCaptcha}
      />

      {#if captchaUntested}
        <Notice icon="alert-triangle" title="Test before saving.">
          Wrong keys block every sign in, including yours. Changes to the
          captcha can only be saved once a test with these keys passes.
        </Notice>
      {/if}
    {/if}
  </Card>
</SettingsForm>
