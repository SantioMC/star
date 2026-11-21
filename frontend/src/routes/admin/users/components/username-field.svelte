<script lang="ts">
    import Input from "@/components/input.svelte";
    import { Label } from "@/components/ui/label";

    let {
        value = $bindable(""),
        issue,
    }: {
        value?: string;
        issue: string | null;
    } = $props();

    const uid = $props.id();
</script>

<div class="grid gap-2">
    <Label for="{uid}-username" class="gap-0.5 select-none">
        Username
        <span class="text-red-400" aria-hidden="true">*</span>
    </Label>
    <Input
        bind:value
        id="{uid}-username"
        type="text"
        autocomplete="off"
        aria-invalid={issue != null}
        aria-describedby={issue ? `${uid}-username-issue` : undefined}
        required
    />
    {#if issue}
        <p id="{uid}-username-issue" class="text-xs text-red-300">
            {issue}
        </p>
    {/if}
</div>
