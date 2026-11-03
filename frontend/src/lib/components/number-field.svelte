<script lang="ts">
    import { Label } from "@/components/ui/label";
    import Input from "./input.svelte";

    let {
        value = $bindable(""),
        id,
        label,
        min,
        max,
    }: {
        value?: string;
        id: string;
        label: string;
        min: number;
        max: number;
    } = $props();
</script>

<div class="grid flex-1 gap-2">
    <Label for={id}>{label}</Label>
    <Input
        bind:value
        {id}
        {min}
        {max}
        type="number"
        inputmode="numeric"
        class="input-clean"
        required
    />
</div>
