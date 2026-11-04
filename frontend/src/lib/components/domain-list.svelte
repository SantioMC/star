<script lang="ts">
    import Button from "./button.svelte";
    import Input from "./input.svelte";
    import PermissionChip from "./permission-chip.svelte";
    import { Label } from "./ui/label";

    let {
        domains = $bindable(),
        id,
        label,
    }: {
        domains: string[];
        id: string;
        label: string;
    } = $props();

    let draft = $state("");

    const add = () => {
        const entered = draft
            .toLowerCase()
            .split(/[\s,]+/)
            .map((domain) => domain.slice(domain.lastIndexOf("@") + 1))
            .filter((domain) => domain.length > 0);

        domains = [...new Set([...domains, ...entered])];
        draft = "";
    };

    const remove = (domain: string) => {
        domains = domains.filter((entry) => entry !== domain);
    };

    const onkeydown = (event: KeyboardEvent) => {
        if (event.key !== "Enter") return;

        event.preventDefault();
        add();
    };
</script>

<div class="grid gap-2">
    <Label for={id}>{label}</Label>
    <div class="flex gap-2">
        <Input
            {id}
            bind:value={draft}
            placeholder="example.com"
            autocomplete="off"
            spellcheck="false"
            {onkeydown}
            onblur={add}
        />
        <Button type="button" variant="secondary" onclick={add}>Add</Button>
    </div>

    {#if domains.length > 0}
        <div class="flex flex-wrap gap-1.5">
            {#each domains as domain (domain)}
                <PermissionChip
                    name={domain}
                    label={domain}
                    description="Remove {domain}"
                    removable
                    onremove={() => remove(domain)}
                />
            {/each}
        </div>
    {/if}
</div>
