<script lang="ts">
    import Card from "@/components/card.svelte";
    import CheckboxCard from "@/components/checkbox-card.svelte";
    import Notice from "@/components/notice.svelte";
    import SettingsForm from "@/components/settings-form.svelte";
    import { getSettingsContext } from "../context.svelte";

    const settings = getSettingsContext();

    let allowRegistration = $state(settings.current!.allow_registration);
    let allowLogins = $state(settings.current!.allow_logins);

    const dirty = $derived(
        allowRegistration !== settings.current!.allow_registration ||
            allowLogins !== settings.current!.allow_logins,
    );

    const save = () =>
        settings.update({
            allow_registration: allowRegistration,
            allow_logins: allowLogins,
        });

    const discard = () => {
        allowRegistration = settings.current!.allow_registration;
        allowLogins = settings.current!.allow_logins;
    };
</script>

<SettingsForm
    title="Access Settings"
    description="Control who can sign up and sign in to this Star instance"
    {dirty}
    onsave={save}
    ondiscard={discard}
>
    <Card
        title="Registration & Logins"
        icon="door-enter"
        description="Applies to every application authenticating through Star, including this dashboard."
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
</SettingsForm>
