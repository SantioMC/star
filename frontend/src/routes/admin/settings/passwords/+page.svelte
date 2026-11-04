<script lang="ts">
    import { goto } from "$app/navigation";
    import * as star from "$lib/star";
    import Button from "@/components/button.svelte";
    import Card from "@/components/card.svelte";
    import ConfirmModal from "@/components/confirm-modal.svelte";
    import Icon from "@/components/icon.svelte";
    import NumberField from "@/components/number-field.svelte";
    import SettingsForm from "@/components/settings-form.svelte";
    import { clearSession, hasPermission } from "@/index";
    import type { PasswordSettings } from "@/star";
    import { toast } from "svelte-sonner";
    import { getSettingsContext } from "../context.svelte";

    type PolicyKey = keyof PasswordSettings;

    interface PolicyField {
        key: PolicyKey;
        label: string;
        min: number;
        max: number;
    }

    const lengthFields: PolicyField[] = [
        { key: "min_length", label: "Minimum Length", min: 8, max: 128 },
        { key: "max_length", label: "Maximum Length", min: 8, max: 512 },
    ];

    const characterFields: PolicyField[] = [
        { key: "min_symbols", label: "Minimum Symbols", min: 0, max: 32 },
        { key: "min_upper", label: "Minimum Uppercase", min: 0, max: 32 },
        { key: "min_lower", label: "Minimum Lowercase", min: 0, max: 32 },
    ];

    const policyKeys = [...lengthFields, ...characterFields].map((it) => it.key);

    const settings = getSettingsContext();

    const canRequireForAll = hasPermission("admin:password:require");

    let requireForAllModal = $state<ConfirmModal>();

    const toInputs = (policy: PasswordSettings) =>
        Object.fromEntries(
            policyKeys.map((key) => [key, String(policy[key])]),
        ) as Record<PolicyKey, string>;

    let inputs = $state(toInputs(settings.current!.password_settings));

    const parsed = $derived(
        Object.fromEntries(
            policyKeys.map((key) => [key, Number(inputs[key])]),
        ) as Record<PolicyKey, number>,
    );

    const dirty = $derived(
        policyKeys.some(
            (key) => parsed[key] !== settings.current!.password_settings[key],
        ),
    );

    const issue = $derived.by(() => {
        if (policyKeys.some((key) => !Number.isInteger(parsed[key]))) {
            return "Every field must be a whole number";
        }

        if (parsed.max_length < parsed.min_length) {
            return "The maximum length must be greater than the minimum length";
        }

        const required =
            parsed.min_symbols + parsed.min_upper + parsed.min_lower;

        if (required > parsed.min_length) {
            return "The required characters add up to more than the minimum length, no password could satisfy this policy";
        }

        return null;
    });

    const save = () => settings.update({ password_settings: parsed });

    const discard = () => {
        inputs = toInputs(settings.current!.password_settings);
    };

    const requireForAll = async () => {
        const result = await star.requirePasswordChangeForAll();
        if ("error" in result) {
            toast.error(result.message);
            return false;
        }

        clearSession();
        goto("/login");
    };
</script>

{#snippet fieldRow(fields: PolicyField[])}
    <div class="flex flex-col sm:flex-row gap-5 sm:gap-6">
        {#each fields as field (field.key)}
            <NumberField
                bind:value={inputs[field.key]}
                id={field.key}
                label={field.label}
                min={field.min}
                max={field.max}
            />
        {/each}
    </div>
{/snippet}

<SettingsForm
    title="Password Settings"
    description="Set the rules every new password must meet"
    {dirty}
    {issue}
    onsave={save}
    ondiscard={discard}
>
    <Card
        title="Password Policy"
        icon="password"
        description="Existing passwords are never revalidated, raising these only applies the next time a password is set."
    >
        <div class="mt-6 flex flex-col gap-5">
            {@render fieldRow(lengthFields)}
            {@render fieldRow(characterFields)}
        </div>

        {#if issue}
            <p class="mt-4 flex items-start gap-2 text-sm text-red-300">
                <Icon icon="alert-triangle" class="mt-0.5" />
                {issue}
            </p>
        {/if}
    </Card>

    {#if canRequireForAll}
        <Card
            title="Require Password Change"
            icon="lock"
            description="Use this if passwords may have been exposed. Everyone is signed out and every user with a password must choose a new one at their next sign in."
        >
            <div class="mt-4">
                <Button
                    type="button"
                    variant="danger"
                    onclick={() => requireForAllModal?.open()}
                >
                    Require for all users
                </Button>
            </div>
        </Card>
    {/if}
</SettingsForm>

<ConfirmModal
    bind:this={requireForAllModal}
    title="Require password change for all users"
    confirmLabel="Require password change"
    onconfirm={requireForAll}
>
    Every session and access token on this server is revoked immediately,
    including yours and those issued to connected applications. Every user with
    a password must choose a new one at their next sign in. This can not be
    undone.
</ConfirmModal>
