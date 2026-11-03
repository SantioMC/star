<script lang="ts">
    import { cn } from "@/utils";
    import type { Snippet } from "svelte";
    import type {
        ClassValue,
        HTMLButtonAttributes,
        HTMLAnchorAttributes,
    } from "svelte/elements";

    import type { ResolvedPathname } from "$app/types";
    import Icon from "./icon.svelte";

    type BaseOptions = {
        class?: ClassValue;
        value?: string;
        children?: Snippet;
        variant?: keyof typeof variants;
        disabled?: boolean;
        loading?: boolean;
    };

    type ButtonOptions = {
        href?: undefined;
    } & HTMLButtonAttributes &
        BaseOptions;

    type AnchorOptions = {
        href: ResolvedPathname;
    } & HTMLAnchorAttributes &
        BaseOptions;

    let {
        value = $bindable(),
        class: className,
        children,
        variant,
        href,
        disabled,
        loading,
        ...rest
    }: ButtonOptions | AnchorOptions = $props();

    const variants = {
        primary: "bg-primary text-primary-foreground not-disabled:hover:bg-primary/80",
        secondary:
            "bg-transparent border border-background-200 text-foreground not-disabled:hover:bg-background-200",
        danger: "bg-transparent border border-red-300/80 text-foreground not-disabled:hover:bg-red-400/30 text-red-300",
        ghost: "bg-transparent text-foreground hover:bg-background-200",
        icon: "p-2 bg-transparent not-disabled:hover:bg-background-200",
    } as const;
</script>

{#if href}
    <a
        class={cn(
            "px-3 py-1.5 rounded-lg text-sm transition-all not-disabled:active:scale-95 flex items-center gap-2 not-disabled:cursor-pointer disabled:brightness-50",
            variants[variant ?? "primary"],
            className,
        )}
        {...rest as HTMLAnchorAttributes}
        {href}
    >
        {#if loading}
            <Icon icon="loader-2" class="animate-spin" />
        {/if}
        {@render children?.()}
    </a>
{:else}
    <button
        class={cn(
            "px-3 py-1.5 rounded-lg text-sm transition-all not-disabled:active:scale-95 flex items-center gap-2 not-disabled:cursor-pointer disabled:brightness-50 justify-center",
            variants[variant ?? "primary"],
            className,
        )}
        disabled={disabled || loading}
        {...rest as HTMLButtonAttributes}
    >
        {#if loading}
            <Icon icon="loader-2" class="animate-spin" />
        {/if}
        {@render children?.()}
    </button>
{/if}
