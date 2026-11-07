<script lang="ts">
    import { branding } from "@/branding";
    import * as star from "$lib/star";
    import { formatTime } from "@/utils";
    import { toast } from "svelte-sonner";
    import Button from "./button.svelte";
    import Copy from "./copy.svelte";
    import Modal from "./modal.svelte";
    import Notice from "./notice.svelte";

    let modal = $state<Modal>();
    let link = $state<star.InviteLink | null>(null);
    let username = $state("");
    let sending = $state(false);

    export const show = (value: star.InviteLink, name: string) => {
        link = value;
        username = name;
        modal?.open();
    };

    const close = () => {
        link = null;
        modal?.close();
    };

    const sendEmail = async () => {
        if (!link || sending) return;

        sending = true;
        const result = await star.emailInvite(link.user_id, link.code);
        sending = false;

        if ("error" in result) {
            return toast.error(result.message);
        }

        toast.success(`Emailed the invite to ${username}`);
    };
</script>

<Modal bind:this={modal} button>
    {#if link}
        <div class="flex flex-col gap-6">
            <div class="flex flex-col gap-1 pr-8">
                <h1 class="text-xl font-semibold">Share the invite link</h1>
                <h2 class="text-sm text-muted">
                    Anyone with this link can finish setting up
                    <span class="text-foreground">{username}</span>. It expires
                    {formatTime(new Date(link.expires_at))}.
                </h2>
            </div>

            <Notice icon="alert-triangle" title="Copy it now.">
                {$branding.name} only stores a hash of this link, so it can not be
                shown again. Regenerate it from the user's page if it is lost.
            </Notice>

            <div
                class="flex min-w-0 items-center gap-1.5 rounded-lg border border-background-200 px-3 py-2"
            >
                <code class="min-w-0 flex-1 truncate font-mono text-sm">
                    {link.url}
                </code>
                <Copy content={link.url} />
            </div>

            <div class="flex items-center gap-2 justify-end">
                <Button
                    variant="secondary"
                    onclick={sendEmail}
                    loading={sending}
                    disabled={!link.email_enabled}
                >
                    Email invite
                </Button>
                <Button onclick={close}>Done</Button>
            </div>
        </div>
    {/if}
</Modal>
