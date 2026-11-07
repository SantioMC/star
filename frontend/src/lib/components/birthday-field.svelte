<script lang="ts">
    import { getLocalTimeZone, today } from "@internationalized/date";
    import DatePicker from "./date-picker.svelte";

    let {
        value = $bindable(""),
        required = false,
        disabled = false,
    }: {
        value?: string;
        required?: boolean;
        disabled?: boolean;
    } = $props();

    const now = today(getLocalTimeZone());
</script>

<DatePicker
    bind:value
    {required}
    {disabled}
    label="Birthday"
    name="birthday"
    min={now.subtract({ years: 150 })}
    max={now}
/>
