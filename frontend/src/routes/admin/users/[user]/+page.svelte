<script lang="ts">
    import { branding } from "@/branding";
    import * as star from "$lib/star";
    import Badge from "@/components/badge.svelte";
    import BirthdayField from "@/components/birthday-field.svelte";
    import Button from "@/components/button.svelte";
    import Card from "@/components/card.svelte";
    import Input from "@/components/input.svelte";
    import Modal from "@/components/modal.svelte";
    import Notice from "@/components/notice.svelte";
    import RoleSelector from "@/components/role-selector.svelte";
    import { Label } from "@/components/ui/label";
    import { hasPermission } from "@/index";
    import { getResource } from "@/resource.svelte";
    import { formatTime, titleCase } from "@/utils";
    import { onMount } from "svelte";
    import { toast } from "svelte-sonner";

    const context = getResource<star.AdminUser>();
    const user = $derived(context.current);

    const canWrite = $derived(
        hasPermission("admin:user:write") && user?.manageable,
    );

    let username = $state("");
    let firstName = $state("");
    let lastName = $state("");
    let birthday = $state("");
    let birthdayTracked = $state(false);
    let saving = $state(false);

    let roles = $state<star.Role[]>([]);
    let savingRoles = $state(false);
    let unlocking = $state(false);

    let rolesModal = $state<Modal>();

    const resetForm = (data: star.AdminUser) => {
        username = data.username;
        firstName = data.first_name ?? "";
        lastName = data.last_name ?? "";
        birthday = data.birthday ?? "";
        roles = [...data.roles];
    };

    resetForm(context.current!);

    onMount(async () => {
        const requirements = await star.getRegisterRequirements();
        birthdayTracked = requirements?.birthday_tracked ?? false;
    });

    const updateUser = async (event: Event) => {
        event.preventDefault();
        if (!user || saving) return;

        saving = true;
        const result = await star.modifyUser(user.id, {
            username,
            first_name: firstName,
            last_name: lastName,
            birthday,
        });
        saving = false;

        if ("error" in result) {
            resetForm(user);
            return toast.error(result.message);
        }

        context.current = result;
        toast.success("Successfully updated user");
    };

    const openRolesModal = () => {
        if (user) roles = [...user.roles];
        rolesModal?.open();
    };

    const updateRoles = async (event: Event) => {
        event.preventDefault();
        if (!user || savingRoles) return;

        savingRoles = true;
        const result = await star.modifyUser(user.id, {
            roles: roles.map((it) => it.id),
        });
        savingRoles = false;

        if ("error" in result) {
            return toast.error(result.message);
        }

        context.current = result;
        roles = [...result.roles];
        rolesModal?.close();
        toast.success("Successfully updated the user's roles");
    };

    const unlock = async () => {
        if (!user || unlocking) return;

        unlocking = true;
        const result = await star.clearLockout(user.id);
        unlocking = false;

        if ("error" in result) {
            return toast.error(result.message);
        }

        context.current = result;
        toast.success(`Cleared ${result.username}'s failed sign in attempts`);
    };
</script>

{#if user}
    {#if user.lockout && (user.lockout.locked || user.lockout.attempts > 0)}
        {@const lockout = user.lockout}
        <Notice
            tone={lockout.locked ? "danger" : "warning"}
            icon={lockout.locked ? "lock" : "alert-triangle"}
            title={lockout.locked
                ? "Sign in locked."
                : "Failed sign in attempts."}
        >
            {user.username} failed to sign in {lockout.attempts}
            {lockout.attempts === 1 ? "time" : "times"}, most recently on
            {formatTime(new Date(lockout.last_attempt_at))}.
            {#if lockout.locked && lockout.locked_until}
                Sign in is blocked until
                {formatTime(new Date(lockout.locked_until))}.
            {/if}

            {#snippet actions()}
                {#if canWrite}
                    <Button
                        variant="secondary"
                        loading={unlocking}
                        onclick={unlock}
                    >
                        {lockout.locked ? "Unlock" : "Reset"}
                    </Button>
                {/if}
            {/snippet}
        </Notice>
    {/if}

    <Card
        title="Roles"
        icon="shield"
        description="Roles grant this user permissions across {$branding.name} and connected applications."
        class="flex flex-col gap-4"
    >
        {#snippet actions()}
            {#if canWrite}
                <Button variant="secondary" onclick={openRolesModal}>
                    {user!.roles.length > 0 ? "Modify roles" : "Add role"}
                </Button>
            {/if}
        {/snippet}

        <table>
            <tbody>
                {#each user.roles as role (role.id)}
                    <tr>
                        <td class="py-2.5 pl-1">
                            <div class="flex items-center gap-2">
                                <span>{titleCase(role.name)}</span>
                                {#if role.internal}
                                    <Badge tone="primary">built-in</Badge>
                                {/if}
                            </div>
                            {#if role.description}
                                <p class="text-sm text-muted">
                                    {role.description}
                                </p>
                            {/if}
                        </td>
                    </tr>
                {:else}
                    <tr class="hover:bg-transparent">
                        <td class="py-6 text-center text-sm text-muted">
                            This user has no assigned roles.
                        </td>
                    </tr>
                {/each}
            </tbody>
        </table>
    </Card>

    <Card title="Profile Information" icon="user">
        <form onsubmit={updateUser} class="mt-6">
            <div class="flex flex-col gap-5">
                <div class="grid gap-2">
                    <Label for="username">Username</Label>
                    <Input
                        bind:value={username}
                        id="username"
                        type="text"
                        disabled={!canWrite}
                        required
                    />
                </div>

                <div class="flex gap-2 sm:gap-6">
                    <div class="flex-1/2 gap-2 grid">
                        <Label for="givenName">Given Name</Label>
                        <Input
                            bind:value={firstName}
                            id="givenName"
                            type="text"
                            disabled={!canWrite}
                        />
                    </div>

                    <div class="flex-1/2 gap-2 grid">
                        <Label for="familyName">Family Name</Label>
                        <Input
                            bind:value={lastName}
                            id="familyName"
                            type="text"
                            disabled={!canWrite}
                        />
                    </div>
                </div>

                {#if birthdayTracked}
                    <BirthdayField bind:value={birthday} disabled={!canWrite} />
                {/if}
            </div>

            {#if canWrite}
                <div class="mt-4 flex items-center gap-2 justify-end">
                    <Button type="submit" loading={saving}>Update</Button>
                </div>
            {/if}
        </form>
    </Card>

    <Modal bind:this={rolesModal} button>
        <form class="flex flex-col gap-6" onsubmit={updateRoles}>
            <div class="flex flex-col gap-1 pr-8">
                <h1 class="text-xl font-semibold">Change roles</h1>
                <h2 class="text-sm text-muted">
                    Select the roles {user.username} should have.
                </h2>
            </div>

            <RoleSelector bind:roles />

            <div class="flex items-center gap-2 justify-end">
                <Button
                    type="button"
                    variant="secondary"
                    onclick={() => rolesModal?.close()}
                >
                    Cancel
                </Button>
                <Button type="submit" loading={savingRoles}>
                    Apply changes
                </Button>
            </div>
        </form>
    </Modal>
{/if}
