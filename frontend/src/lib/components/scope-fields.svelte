<script lang="ts">
    import Checkbox from "@/components/checkbox.svelte";
    import Input from "@/components/input.svelte";
    import { Label } from "@/components/ui/label";

    let {
        id,
        namespace,
        name = $bindable(),
        description = $bindable(),
        requestable = $bindable(),
    }: {
        id: string;
        namespace: string;
        name: string;
        description: string;
        requestable: boolean;
    } = $props();
</script>

<div class="flex flex-col gap-5 w-full">
    <div class="grid gap-2">
        <Label for="{id}_scope_name">Scope</Label>
        <Input
            id="{id}_scope_name"
            type="text"
            class="font-mono"
            placeholder="{namespace}:example"
            bind:value={name}
            required
        />
    </div>

    <div class="grid gap-2">
        <Label for="{id}_scope_description">Description</Label>
        <Input
            id="{id}_scope_description"
            type="text"
            placeholder="Allows doing something on your behalf"
            bind:value={description}
            required
        />
    </div>

    <label
        class="flex cursor-pointer items-start gap-3 rounded-lg border border-background-200 p-3 transition-colors hover:bg-hover has-checked:border-primary/40 has-checked:bg-hover"
    >
        <Checkbox
            id="{id}_scope_requestable"
            class="mt-0.5"
            bind:checked={requestable}
        />
        <div class="flex flex-col gap-0.5">
            <span class="text-sm select-none">
                Requestable by other applications
            </span>
            <span class="text-sm text-muted select-none">
                Lets users grant this scope when authenticating with other
                applications.
            </span>
        </div>
    </label>
</div>
