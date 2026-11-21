<script lang="ts">
    import * as star from "$lib/star";
    import Button from "@/components/button.svelte";
    import CheckboxCard from "@/components/checkbox-card.svelte";
    import Modal from "@/components/modal.svelte";
    import NewPasswordFields from "@/components/new-password-fields.svelte";
    import Segmented from "@/components/segmented.svelte";
    import type { SelectOption } from "@/components/select.svelte";
    import { validateUsername, type UsernameRules } from "@/username";
    import { toast } from "svelte-sonner";
    import UserProfileFields, {
        emptyProfile,
        toProfileRequest,
    } from "./user-profile-fields.svelte";
    import UsernameField from "./username-field.svelte";

    const {
        oncreated,
    }: {
        oncreated: (user: star.AdminUser) => void;
    } = $props();

    const typeOptions: SelectOption<star.UserType>[] = [
        { value: "regular", label: "Regular User" },
        { value: "service", label: "Service Account" },
    ];

    let modal = $state<Modal>();
    let type = $state<star.UserType>("regular");
    let profile = $state(emptyProfile());
    let password = $state("");
    let confirmPassword = $state("");
    let requirePasswordChange = $state(true);
    let saving = $state(false);

    let usernameRules = $state<UsernameRules | null>(null);
    let birthdayTracked = $state(false);

    const usernameIssue = $derived(
        usernameRules && profile.username
            ? validateUsername(profile.username, usernameRules)
            : null,
    );

    const complete = $derived.by(() => {
        if (!profile.username.trim() || usernameIssue) return false;
        if (type === "service") return true;

        return Boolean(profile.email && password && confirmPassword);
    });

    export const open = async () => {
        type = "regular";
        profile = emptyProfile();
        password = "";
        confirmPassword = "";
        requirePasswordChange = true;
        modal?.open();

        const requirements = await star.getRegisterRequirements();
        usernameRules = requirements?.username ?? null;
        birthdayTracked = requirements?.birthday_tracked ?? false;
    };

    const buildRequest = (): star.CreateUserRequest => {
        if (type === "service") {
            return { type, username: profile.username };
        }

        return {
            type,
            ...toProfileRequest(profile),
            password,
            require_password_change: requirePasswordChange,
        };
    };

    const create = async (event: Event) => {
        event.preventDefault();
        if (saving || usernameIssue) return;

        if (type === "regular" && password !== confirmPassword) {
            return toast.error("The passwords do not match");
        }

        saving = true;
        const result = await star.createUser(buildRequest());
        saving = false;

        if ("error" in result) {
            return toast.error(result.message);
        }

        modal?.close();
        toast.success(`Created ${result.username}`);
        oncreated(result);
    };
</script>

<Modal bind:this={modal} button>
    <form class="flex flex-col gap-6" onsubmit={create}>
        <div class="flex flex-col gap-1 pr-8">
            <h1 class="text-xl font-semibold">Create user</h1>
            <h2 class="text-sm text-muted">
                {#if type === "service"}
                    Create a service account for applications and automations,
                    it can not sign in with a password
                {:else}
                    Create a user with a password they can sign in with right
                    away
                {/if}
            </h2>
        </div>

        <div class="flex flex-col gap-5">
            <Segmented bind:value={type} options={typeOptions} label="Type" />

            {#if type === "service"}
                <UsernameField
                    bind:value={profile.username}
                    issue={usernameIssue}
                />
            {:else}
                <UserProfileFields
                    bind:profile
                    {usernameIssue}
                    {birthdayTracked}
                />
                <NewPasswordFields
                    bind:password
                    bind:confirmPassword
                    label="Password"
                />
                <CheckboxCard
                    bind:checked={requirePasswordChange}
                    title="Require password change"
                >
                    The user must pick a new password the first time they sign
                    in.
                </CheckboxCard>
            {/if}
        </div>

        <div class="flex items-center gap-2 justify-end">
            <Button
                type="button"
                variant="secondary"
                onclick={() => modal?.close()}
            >
                Cancel
            </Button>
            <Button type="submit" loading={saving} disabled={!complete}>
                Create user
            </Button>
        </div>
    </form>
</Modal>
