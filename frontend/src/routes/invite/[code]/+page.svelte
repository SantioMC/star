<script lang="ts">
    import { goto } from "$app/navigation";
    import { page } from "$app/state";
    import * as star from "$lib/star";
    import AuthShell from "@/components/auth-shell.svelte";
    import BirthdayField from "@/components/birthday-field.svelte";
    import Button from "@/components/button.svelte";
    import Icon from "@/components/icon.svelte";
    import Input from "@/components/input.svelte";
    import NewPasswordFields from "@/components/new-password-fields.svelte";
    import { Label } from "@/components/ui/label";
    import { onMount } from "svelte";

    const code = page.params.code!;

    let invite = $state<star.InviteDetails | null>(null);
    let invalid = $state<string | null>(null);

    let firstName = $state("");
    let lastName = $state("");
    let birthday = $state("");
    let password = $state("");
    let confirmPassword = $state("");
    let loading = $state(false);
    let error = $state("");

    onMount(async () => {
        const response = await star.lookupInvite(code);

        if ("error" in response) {
            invalid = response.message;
            return;
        }

        firstName = response.first_name ?? "";
        lastName = response.last_name ?? "";
        birthday = response.birthday ?? "";
        invite = response;
    });

    const submit = async (event: Event) => {
        event.preventDefault();
        if (loading) return;

        error = "";

        if (password !== confirmPassword) {
            error = "Your passwords do not match!";
            return;
        }

        loading = true;
        const response = await star.acceptInvite({
            code,
            password,
            first_name: firstName ? firstName : undefined,
            last_name: lastName ? lastName : undefined,
            birthday: birthday ? birthday : undefined,
        });
        loading = false;

        if ("error" in response) {
            if (response.status === 410) {
                invalid = response.message;
                return;
            }

            error = response.message;
            return;
        }

        localStorage.setItem(
            "star-login-note",
            "Your account is ready, you may now login",
        );

        goto("/login");
    };
</script>

<svelte:head>
    <meta name="referrer" content="no-referrer" />
</svelte:head>

<AuthShell>
    {#if invalid}
        <div class="flex flex-col">
            <h1 class="text-lg">Unable to use this invite</h1>
            <p class="text-red-300 text-sm mt-1">{invalid}</p>
        </div>
        <div class="mt-6">
            <Button
                class="w-full"
                variant="secondary"
                onclick={() => goto("/login")}
            >
                Go to Login
            </Button>
        </div>
    {:else if !invite}
        <div class="flex flex-col">
            <h1 class="text-lg">Checking your invite</h1>
            <p class="text-muted text-sm">
                One moment while we verify your invite link...
            </p>
        </div>
        <div class="flex items-center justify-center h-40 w-full">
            <Icon icon="loader-2" class="animate-spin text-3xl text-muted" />
        </div>
    {:else}
        <div class="flex flex-col mb-6">
            <h1 class="text-lg">Finish setting up your account</h1>
            <p class="text-muted text-sm">
                You can change these details after signing in
            </p>
        </div>

        {#if error}
            <div
                role="alert"
                class="mb-6 flex items-start gap-2 rounded-lg border border-red-400/30 bg-red-400/5 px-3 py-2 text-sm text-red-300"
            >
                <Icon icon="alert-triangle" class="mt-0.5 shrink-0" />
                <span>{error}</span>
            </div>
        {/if}

        <form onsubmit={submit}>
            <div class="flex flex-col gap-6">
                <div class="grid gap-2">
                    <Label for="username" class="select-none">Username</Label>
                    <Input value={invite.username} id="username" disabled />
                </div>
                <div class="grid gap-2">
                    <Label for="email" class="select-none">Email</Label>
                    <Input value={invite.email} id="email" disabled />
                </div>
                <div class="grid gap-6 sm:grid-cols-2">
                    <div class="grid gap-2">
                        <Label for="first-name" class="select-none">
                            First Name
                        </Label>
                        <Input
                            bind:value={firstName}
                            id="first-name"
                            type="text"
                            maxlength={255}
                            autocomplete="given-name"
                            disabled={invite.first_name !== null}
                        />
                    </div>
                    <div class="grid gap-2">
                        <Label for="last-name" class="select-none">
                            Last Name
                        </Label>
                        <Input
                            bind:value={lastName}
                            id="last-name"
                            type="text"
                            maxlength={255}
                            autocomplete="family-name"
                            disabled={invite.last_name !== null}
                        />
                    </div>
                </div>
                {#if invite.birthday_tracked}
                    <BirthdayField
                        bind:value={birthday}
                        required={invite.birthday_required}
                        disabled={invite.birthday !== null}
                    />
                {/if}
                <NewPasswordFields bind:password bind:confirmPassword />
            </div>
            <div class="mt-6">
                <Button type="submit" class="w-full" {loading}>
                    Create Account
                </Button>
            </div>
        </form>
    {/if}
</AuthShell>
