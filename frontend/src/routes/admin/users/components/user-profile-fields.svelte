<script lang="ts" module>
    import type * as star from "$lib/star";

    export interface UserProfile {
        username: string;
        email: string;
        firstName: string;
        lastName: string;
        birthday: string;
    }

    export const emptyProfile = (): UserProfile => ({
        username: "",
        email: "",
        firstName: "",
        lastName: "",
        birthday: "",
    });

    export const toProfileRequest = (
        profile: UserProfile,
    ): star.UserProfileRequest => ({
        username: profile.username,
        email: profile.email,
        first_name: profile.firstName ? profile.firstName : undefined,
        last_name: profile.lastName ? profile.lastName : undefined,
        birthday: profile.birthday ? profile.birthday : undefined,
    });
</script>

<script lang="ts">
    import BirthdayField from "@/components/birthday-field.svelte";
    import Input from "@/components/input.svelte";
    import { Label } from "@/components/ui/label";
    import UsernameField from "./username-field.svelte";

    let {
        profile = $bindable(),
        usernameIssue,
        birthdayTracked,
    }: {
        profile: UserProfile;
        usernameIssue: string | null;
        birthdayTracked: boolean;
    } = $props();

    const uid = $props.id();
</script>

<UsernameField bind:value={profile.username} issue={usernameIssue} />

<div class="grid gap-2">
    <Label for="{uid}-email" class="gap-0.5 select-none">
        Email
        <span class="text-red-400" aria-hidden="true">*</span>
    </Label>
    <Input
        bind:value={profile.email}
        id="{uid}-email"
        type="email"
        maxlength={255}
        autocomplete="off"
        required
    />
</div>

<div class="grid gap-5 sm:grid-cols-2">
    <div class="grid gap-2">
        <Label for="{uid}-first-name" class="select-none">First Name</Label>
        <Input
            bind:value={profile.firstName}
            id="{uid}-first-name"
            type="text"
            maxlength={255}
            autocomplete="off"
        />
    </div>

    <div class="grid gap-2">
        <Label for="{uid}-last-name" class="select-none">Last Name</Label>
        <Input
            bind:value={profile.lastName}
            id="{uid}-last-name"
            type="text"
            maxlength={255}
            autocomplete="off"
        />
    </div>
</div>

{#if birthdayTracked}
    <BirthdayField bind:value={profile.birthday} />
{/if}
