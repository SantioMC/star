<script lang="ts">
    import { branding } from "@/branding";
    import type { Snippet } from "svelte";
    import BrandLinks from "./brand-links.svelte";
    import BrandMark from "./brand-mark.svelte";
    import Card from "./card.svelte";

    const {
        children,
        banner,
    }: {
        children: Snippet;
        banner?: Snippet;
    } = $props();
</script>

{#snippet content()}
    {#if banner}
        <div class="mb-4">{@render banner()}</div>
    {/if}
    {@render children()}
{/snippet}

{#if $branding.layout === "split"}
    <div class="min-h-full w-full lg:grid lg:grid-cols-[minmax(0,5fr)_minmax(0,6fr)]">
        <aside
            class="flex flex-col justify-between gap-6 bg-primary px-6 py-5 text-primary-foreground lg:sticky lg:top-0 lg:h-dvh lg:p-12"
        >
            <BrandMark name={false} class="max-lg:hidden" />
            <BrandMark class="lg:hidden" />

            <div class="flex flex-col gap-6 max-lg:hidden">
                <p
                    class="font-serif text-5xl leading-tight font-semibold tracking-tight break-words select-none"
                >
                    {$branding.name}
                </p>
                <BrandLinks class="text-primary-foreground/80 [&_a:hover]:text-primary-foreground" />
            </div>
        </aside>

        <div class="flex flex-col items-center justify-center px-4 py-12 lg:min-h-dvh">
            <div class="w-full sm:w-sm">
                {@render content()}
            </div>
            <BrandLinks class="mt-10 justify-center text-muted lg:hidden" />
        </div>
    </div>
{:else if $branding.layout === "minimal"}
    <div class="flex min-h-full w-full flex-col px-6 py-6 sm:px-10">
        <BrandMark />

        <div class="flex flex-1 items-center justify-center py-12">
            <div class="w-full sm:w-sm">
                {@render content()}
            </div>
        </div>

        <div
            class="flex flex-col-reverse gap-2 border-t border-background-200 pt-6 text-sm text-muted sm:flex-row sm:items-center sm:justify-between"
        >
            <span class="select-none">{$branding.name}</span>
            <BrandLinks />
        </div>
    </div>
{:else}
    <div class="flex min-h-full w-full flex-col items-center justify-center gap-8 px-4 py-12">
        <BrandMark class="text-lg" />

        <div class="w-full sm:w-sm">
            {#if banner}
                <div class="mb-4">{@render banner()}</div>
            {/if}
            <Card class="w-full">
                {@render children()}
            </Card>
        </div>

        <BrandLinks class="justify-center text-muted" />
    </div>
{/if}
