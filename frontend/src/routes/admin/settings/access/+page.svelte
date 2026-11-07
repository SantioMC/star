<script lang="ts">
    import { branding } from "@/branding";
    import Card from "@/components/card.svelte";
    import CheckboxCard from "@/components/checkbox-card.svelte";
    import Notice from "@/components/notice.svelte";
    import NumberField from "@/components/number-field.svelte";
    import SettingsForm from "@/components/settings-form.svelte";
    import { getSettingsContext } from "../context.svelte";

    const settings = getSettingsContext();

    let allowRegistration = $state(settings.current!.allow_registration);
    let allowLogins = $state(settings.current!.allow_logins);
    let trackBirthdays = $state(settings.current!.track_birthdays);
    let minimumAge = $state(String(settings.current!.minimum_age));

    const effectiveMinimumAge = $derived(
        trackBirthdays ? Number(minimumAge) : 0,
    );

    const dirty = $derived(
        allowRegistration !== settings.current!.allow_registration ||
            allowLogins !== settings.current!.allow_logins ||
            trackBirthdays !== settings.current!.track_birthdays ||
            effectiveMinimumAge !== settings.current!.minimum_age,
    );

    const save = () =>
        settings.update({
            allow_registration: allowRegistration,
            allow_logins: allowLogins,
            track_birthdays: trackBirthdays,
            minimum_age: effectiveMinimumAge,
        });

    const discard = () => {
        allowRegistration = settings.current!.allow_registration;
        allowLogins = settings.current!.allow_logins;
        trackBirthdays = settings.current!.track_birthdays;
        minimumAge = String(settings.current!.minimum_age);
    };
</script>

<SettingsForm
    title="Access Settings"
    description="Control who can sign up and sign in to this {$branding.name} instance"
    {dirty}
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
                Lets anyone create an account. Turn this off to only allow
                accounts you create yourself.
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
            Lets users add a birthday to their account and share it with
            applications. While off, none are collected and any already saved
            stay stored but hidden.
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
                Users without a birthday are signed out and must enter one on
                their next sign in, and anyone under this age is blocked from
                signing in.
            </Notice>
        {/if}
    </Card>
</SettingsForm>
