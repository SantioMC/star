<script lang="ts">
    import * as star from "$lib/star";
    import Button from "@/components/button.svelte";
    import Icon from "@/components/icon.svelte";
    import Input from "@/components/input.svelte";
    import { Label } from "@/components/ui/label";
    import { titleCase } from "@/utils";

    let {
        username = $bindable(""),
        password = $bindable(""),
        lockUsername = false,
        allowRegistration = true,
        submitLabel = "Login",
        loading = false,
        providers = [],
        onpasskey,
    }: {
        username?: string;
        password?: string;
        lockUsername?: boolean;
        allowRegistration?: boolean;
        submitLabel?: string;
        loading?: boolean;
        providers?: star.LoginProvider[];
        onpasskey?: () => void;
    } = $props();
</script>

<div class="flex flex-col gap-6">
    <div class="grid gap-2">
        <Label for="username">Username or Email</Label>
        <Input
            bind:value={username}
            id="username"
            type="text"
            autocomplete={onpasskey ? "username webauthn" : "username"}
            readonly={lockUsername}
            class={lockUsername ? "text-muted" : ""}
            required
        />
    </div>
    <div class="grid gap-2">
        <div class="flex items-center justify-between">
            <Label for="password">Password</Label>
            {#if !lockUsername}
                <a
                    href="/forgot-password"
                    class="text-sm text-primary hover:text-primary/80 transition-colors"
                >
                    Forgot password?
                </a>
            {/if}
        </div>
        <Input
            bind:value={password}
            id="password"
            type="password"
            autocomplete="current-password"
            required
        />
    </div>
</div>

<div class="mt-6 flex flex-col gap-2">
    <Button type="submit" class="w-full" {loading}>{submitLabel}</Button>

    {#if onpasskey}
        <Button
            type="button"
            variant="secondary"
            class="w-full"
            disabled={loading}
            onclick={onpasskey}
        >
            <Icon icon="fingerprint" />
            Sign in with a passkey
        </Button>
    {/if}

    {#each providers as provider (provider.id)}
        <Button
            type="button"
            variant="secondary"
            class="w-full"
            disabled={loading}
            onclick={() =>
                location.assign(star.getProviderLoginUrl(provider.id))}
        >
            <Icon icon={provider.icon || "login"} />
            Continue with {titleCase(provider.name)}
        </Button>
    {/each}

    {#if allowRegistration}
        <div class="flex items-center gap-1 text-sm text-muted select-none">
            New user?
            <a
                href="/register"
                class="text-primary hover:text-primary/80 transition-colors"
            >
                Create an account
            </a>
        </div>
    {/if}
</div>
