<script lang="ts">
    import { branding } from "@/branding";
    import { cn } from "@/utils";
    import type { ClassValue } from "svelte/elements";

    const { class: className }: { class?: ClassValue } = $props();

    const links = $derived(
        [
            {
                label: "Support",
                href: $branding.support_email
                    ? `mailto:${$branding.support_email}`
                    : null,
            },
            { label: "Privacy Policy", href: $branding.privacy_url },
            { label: "Terms of Service", href: $branding.terms_url },
        ].filter((link) => link.href !== null),
    );
</script>

{#if links.length > 0}
    <nav
        aria-label="Legal and support"
        class={cn("flex flex-wrap items-center gap-x-4 gap-y-1 text-sm", className)}
    >
        {#each links as link (link.label)}
            <a
                href={link.href}
                target="_blank"
                rel="noopener noreferrer"
                class="transition-colors hover:text-foreground"
            >
                {link.label}
            </a>
        {/each}
    </nav>
{/if}
