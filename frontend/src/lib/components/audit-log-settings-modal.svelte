<script lang="ts">
    import * as star from "$lib/star";
    import { toast } from "svelte-sonner";
    import Button from "./button.svelte";
    import CheckboxCard from "./checkbox-card.svelte";
    import Modal from "./modal.svelte";
    import Notice from "./notice.svelte";
    import NumberField from "./number-field.svelte";
    import Select, { type SelectOption } from "./select.svelte";
    import { Label } from "./ui/label";

    let { settings = $bindable() }: { settings: star.AuditSettings } = $props();

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

    let modal = $state<Modal>();
    let saving = $state(false);
    let preset = $state("0");
    let customDays = $state("");
    let anonymize = $state(false);

    const days = $derived(
        preset === "custom" ? Number(customDays) : Number(preset),
    );

    const invalid = $derived(
        !Number.isInteger(days) || (days !== 0 && days < 7),
    );

    const shortened = $derived.by(() => {
        if (invalid || days === 0) return false;

        const current = settings.retention_days;
        return current === null || days < current;
    });

    const cutoff = $derived(
        new Date(Date.now() - days * 86_400_000).toLocaleDateString(undefined, {
            dateStyle: "long",
        }),
    );

    const dirty = $derived(
        days !== (settings.retention_days ?? 0) ||
            anonymize !== settings.anonymize,
    );

    export const open = () => {
        const current = String(settings.retention_days ?? 0);
        const known = presets.some((it) => it.value === current);

        preset = known ? current : "custom";
        customDays = known ? "" : current;
        anonymize = settings.anonymize;
        modal?.open();
    };

    const save = async (event: Event) => {
        event.preventDefault();
        if (saving || invalid || !dirty) return;

        saving = true;
        const result = await star.updateAuditSettings({
            retention_days: days,
            anonymize,
        });
        saving = false;

        if ("error" in result) {
            return toast.error(result.message);
        }

        settings = result;
        modal?.close();
        toast.success("Successfully updated the audit log settings");
    };
</script>

<Modal bind:this={modal} button>
    <form class="flex flex-col gap-6" onsubmit={save}>
        <div class="flex flex-col gap-1 pr-8">
            <h1 class="text-xl font-semibold">Retention & privacy</h1>
            <h2 class="text-sm text-muted">
                Choose how long audit logs are kept and what they store.
            </h2>
        </div>

        <div class="flex flex-col gap-5">
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
        </div>

        <div class="flex items-center gap-2 justify-end">
            <Button
                type="button"
                variant="secondary"
                onclick={() => modal?.close()}
            >
                Cancel
            </Button>
            <Button
                type="submit"
                variant={shortened ? "danger" : "primary"}
                loading={saving}
                disabled={invalid || !dirty}
            >
                {shortened ? "Save and delete older logs" : "Save"}
            </Button>
        </div>
    </form>
</Modal>
