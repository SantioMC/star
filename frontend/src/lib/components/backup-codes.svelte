<script lang="ts">
    import { downloadFile } from "@/utils";
    import { toast } from "svelte-sonner";
    import Button from "./button.svelte";
    import Icon from "./icon.svelte";
    import Notice from "./notice.svelte";

    const { codes }: { codes: string[] } = $props();

    const contents = $derived(codes.join("\n"));

    const copy = async () => {
        await navigator.clipboard.writeText(contents);
        toast.success("Backup codes copied to clipboard");
    };

    const download = () => {
        const blob = new Blob([contents + "\n"], { type: "text/plain" });
        downloadFile(blob, "star-backup-codes.txt");
    };
</script>

<div class="flex flex-col gap-4">
    <Notice icon="alert-triangle" title="Save these now.">
        Each code works once and they won't be shown again.
    </Notice>

    <ol
        class="grid grid-cols-2 gap-x-6 gap-y-2 rounded-lg border border-background-200 bg-background px-5 py-4 font-mono text-sm"
    >
        {#each codes as code, index (code)}
            <li class="flex items-center gap-3">
                <span class="w-4 text-right text-xs text-muted select-none">
                    {index + 1}
                </span>
                {code}
            </li>
        {/each}
    </ol>

    <div class="flex items-center gap-2">
        <Button type="button" variant="secondary" onclick={copy}>
            <Icon icon="copy" />
            Copy
        </Button>
        <Button type="button" variant="secondary" onclick={download}>
            <Icon icon="download" />
            Download
        </Button>
    </div>
</div>
