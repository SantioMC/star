<script lang="ts" module>
    import type { SelectOption } from "@/components/select.svelte";

    const options: SelectOption[] = [
        { value: "1", label: "In 1 hour" },
        { value: "12", label: "In 12 hours" },
        { value: "24", label: "In a day" },
        { value: "72", label: "In 3 days" },
        { value: "168", label: "In 7 days" },
    ];

    export const defaultInviteExpiry = "72";
</script>

<script lang="ts">
    import Select from "@/components/select.svelte";
    import { Label } from "@/components/ui/label";

    let {
        value = $bindable(defaultInviteExpiry),
    }: {
        value?: string;
    } = $props();
</script>

<div class="grid gap-2">
    <Label for="inviteExpiry" class="select-none">Expires</Label>
    <Select id="inviteExpiry" bind:value {options} />
    <p class="text-xs text-muted select-none">
        The invited user is deleted when it expires.
    </p>
</div>
