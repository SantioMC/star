<script lang="ts">
    import { page } from "$app/state";
    import Button from "@/components/button.svelte";
    import Icon from "@/components/icon.svelte";
</script>

<div
    class="flex flex-col gap-6 items-center justify-center text-center min-h-dvh px-4"
>
    <Icon icon="error-404" class="text-5xl text-primary/60" />
    <div class="flex flex-col gap-2">
        <h1 class="text-5xl font-semibold">Page not found</h1>
        <p class="text-sm text-muted max-w-md">
            Nothing lives at <span class="font-mono text-foreground/80 break-all"
                >{page.url.pathname}</span
            >
        </p>
    </div>

    <Button href="/" class="mt-8">Go home</Button>
</div>
