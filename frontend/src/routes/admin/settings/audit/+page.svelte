<script lang="ts">
    import { goto } from "$app/navigation";
    import * as star from "$lib/star";
    import Card from "@/components/card.svelte";
    import CheckboxCard from "@/components/checkbox-card.svelte";
    import Icon from "@/components/icon.svelte";
    import Notice from "@/components/notice.svelte";
    import NumberField from "@/components/number-field.svelte";
    import Select, { type SelectOption } from "@/components/select.svelte";
    import SettingsForm from "@/components/settings-form.svelte";
    import { Label } from "@/components/ui/label";
    import { hasPermission } from "@/index";
    import { onMount } from "svelte";
    import { toast } from "svelte-sonner";

    const presets: SelectOption[] = [
        { value: "0", label: "Forever" },
        { value: "7", label: "7 days" },
        { value: "30", label: "30 days" },
        { value: "90", label: "90 days" },
        { value: "180", label: "180 days" },
        { value: "365", label: "1 year" },
        { value: "730", label: "2 years" },
        { value: "custom", label: "Custom" },
    ];

    let settings = $state<star.AuditSettings | null>(null);
    let preset = $state("0");
    let customDays = $state("");
    let anonymize = $state(false);

    const days = $derived(
        preset === "custom" ? Number(customDays) : Number(preset),
    );

    const issue = $derived(
        !Number.isInteger(days) || (days !== 0 && days < 7)
            ? "Logs must be kept for at least 7 days"
            : null,
    );

    const shortened = $derived.by(() => {
        if (issue || days === 0) return false;

        const current = settings?.retention_days ?? null;
        return current === null || days < current;
    });

    const cutoff = $derived(
        new Date(Date.now() - days * 86_400_000).toLocaleDateString(undefined, {
            dateStyle: "long",
        }),
    );

    const dirty = $derived(
        days !== (settings?.retention_days ?? 0) ||
            anonymize !== settings?.anonymize,
    );

    const reset = (current: star.AuditSettings) => {
        const retention = String(current.retention_days ?? 0);
        const known = presets.some((it) => it.value === retention);

        preset = known ? retention : "custom";
        customDays = known ? "" : retention;
        anonymize = current.anonymize;
    };

    const save = async () => {
        const result = await star.updateAuditSettings({
            retention_days: days,
            anonymize,
        });

        if ("error" in result) {
            toast.error(result.message);
            return;
        }

        settings = result;
        reset(result);
        toast.success("Successfully updated the audit log settings");
    };

    const discard = () => {
        if (settings) reset(settings);
    };

    onMount(async () => {
        if (!hasPermission("admin:audit:read")) return goto("/");

        const result = await star.getAuditSettings();
        if (!result) {
            toast.error("Failed to load the audit log settings");
            return;
        }

        settings = result;
        reset(result);
    });
</script>

{#if settings}
    <SettingsForm
        title="Audit Log Settings"
        description="Choose how long audit logs are kept and what they store"
        permission="admin:audit:write"
        saveLabel={shortened ? "Save and delete older logs" : undefined}
        saveVariant={shortened ? "danger" : undefined}
        {dirty}
        {issue}
        onsave={save}
        ondiscard={discard}
    >
        <Card
            title="Retention & Privacy"
            icon="history"
            description="Applies to every audit log recorded by this instance."
            class="flex flex-col gap-5"
        >
            <div class="grid gap-2">
                <Label for="retention">Keep logs for</Label>
                <Select id="retention" bind:value={preset} options={presets} />
            </div>

            {#if preset === "custom"}
                <NumberField
                    bind:value={customDays}
                    id="customDays"
                    label="Days to keep"
                    min={7}
                    max={36500}
                />
            {/if}

            {#if issue}
                <p class="flex items-start gap-2 text-sm text-red-300">
                    <Icon icon="alert-triangle" class="mt-0.5" />
                    {issue}
                </p>
            {/if}

            {#if shortened}
                <Notice
                    tone="danger"
                    icon="alert-triangle"
                    title="Older logs will be permanently deleted."
                >
                    Within the next hour, every audit log older than {days} days
                    (before {cutoff}) is deleted, and this keeps happening every
                    hour. Deleted logs can not be recovered.
                </Notice>
            {/if}

            <CheckboxCard bind:checked={anonymize} title="Anonymize new logs">
                New logs keep the user's ID, but not their name, email, IP
                address or browser. Existing logs are left as they are.
            </CheckboxCard>
        </Card>
    </SettingsForm>
{/if}
