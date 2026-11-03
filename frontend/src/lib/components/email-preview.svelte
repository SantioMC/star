<script lang="ts">
    import * as star from "$lib/star";
    import { cn } from "@/utils";
    import type { ClassValue } from "svelte/elements";
    import Icon from "./icon.svelte";

    const {
        template,
        branding,
        removeLogo = false,
        class: className,
    }: {
        template: star.EmailTemplate;
        branding?: star.UpdateBrandingRequest;
        removeLogo?: boolean;
        class?: ClassValue;
    } = $props();

    let url = $state<string | null>(null);
    let error = $state("");

    $effect(() => {
        const request: star.EmailPreviewRequest = {
            template,
            branding: $state.snapshot(branding),
            remove_logo: removeLogo,
        };

        let cancelled = false;
        const timer = setTimeout(async () => {
            const result = await star.createEmailPreview(request);
            if (cancelled) return;

            if ("error" in result) {
                error = result.message;
                return;
            }

            error = "";
            url = star.getEmailPreviewUrl(result.ticket);
        }, 300);

        return () => {
            cancelled = true;
            clearTimeout(timer);
        };
    });
</script>

<div
    class={cn(
        "relative overflow-hidden rounded-lg border border-background-200 bg-[#1a1a1a]",
        className,
    )}
>
    {#if url}
        <iframe
            title="Email preview"
            src={url}
            sandbox=""
            referrerpolicy="no-referrer"
            class="size-full"
        ></iframe>
    {:else}
        <div class="flex size-full items-center justify-center">
            {#if error}
                <p role="alert" class="text-sm text-red-300">{error}</p>
            {:else}
                <Icon icon="loader-2" class="animate-spin text-3xl text-muted" />
            {/if}
        </div>
    {/if}
</div>
