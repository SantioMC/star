<script lang="ts">
    import { goto } from "$app/navigation";
    import { page } from "$app/state";
    import * as star from "$lib/star";
    import Badge from "@/components/badge.svelte";
    import Button from "@/components/button.svelte";
    import Card from "@/components/card.svelte";
    import ConfirmModal from "@/components/confirm-modal.svelte";
    import Icon from "@/components/icon.svelte";
    import Input from "@/components/input.svelte";
    import Menu, { type MenuItem } from "@/components/menu.svelte";
    import Modal from "@/components/modal.svelte";
    import Notice from "@/components/notice.svelte";
    import RoleSelector from "@/components/role-selector.svelte";
    import * as Avatar from "@/components/ui/avatar";
    import { Label } from "@/components/ui/label";
    import { hasPermission } from "@/index";
    import { downloadFile, formatTime, simpleDate, titleCase } from "@/utils";
    import { onMount } from "svelte";
    import { toast } from "svelte-sonner";

    const canWrite = hasPermission("admin:user:write");
    const canDelete = hasPermission("admin:user:delete");
    const canExport = hasPermission("admin:user:export");
    const canRevokeTokens = hasPermission("admin:token:revoke");

    let user = $state<star.AdminUser | null>(null);

    let username = $state("");
    let firstName = $state("");
    let lastName = $state("");
    let saving = $state(false);

    let roles = $state<star.Role[]>([]);
    let savingRoles = $state(false);

    let newPassword = $state("");
    let confirmPassword = $state("");
    let savingPassword = $state(false);
    let unlocking = $state(false);

    let deleteUserModal = $state<ConfirmModal>();
    let signOutModal = $state<ConfirmModal>();
    let revokeTokensModal = $state<ConfirmModal>();
    let rolesModal = $state<Modal>();
    let passwordModal = $state<Modal>();

    const displayName = $derived.by(() => {
        if (!user) return "";

        const fullName = [user.first_name, user.last_name]
            .filter(Boolean)
            .join(" ");

        return fullName || user.username;
    });

    const menuItems = $derived.by(() => {
        const items: MenuItem[] = [];

        if (canWrite) {
            items.push({
                label: "Set password",
                icon: "key",
                onclick: openPasswordModal,
            });
        }

        if (canExport) {
            items.push({
                label: "Export data",
                icon: "download",
                onclick: exportData,
            });
        }

        if (canRevokeTokens) {
            items.push({
                label: "Revoke access tokens",
                icon: "ban",
                onclick: () => revokeTokensModal?.open(),
            });
        }

        if (canDelete) {
            items.push({
                label: "Sign out all sessions",
                icon: "logout",
                onclick: () => signOutModal?.open(),
            });
            items.push({
                label: "Delete user",
                icon: "trash",
                danger: true,
                onclick: () => deleteUserModal?.open(),
            });
        }

        return items;
    });

    const resetForm = (data: star.AdminUser) => {
        username = data.username;
        firstName = data.first_name ?? "";
        lastName = data.last_name ?? "";
        roles = [...data.roles];
    };

    onMount(async () => {
        if (!hasPermission("admin:user:read")) return goto("/");

        const data = await star.getUser(page.params.user!);
        if (!data) return goto("/admin/users");

        user = data;
        resetForm(data);
    });

    const updateUser = async (event: Event) => {
        event.preventDefault();
        if (!user || saving) return;

        saving = true;
        const result = await star.modifyUser(user.id, {
            username,
            first_name: firstName,
            last_name: lastName,
        });
        saving = false;

        if ("error" in result) {
            resetForm(user);
            return toast.error(result.message);
        }

        user = result;
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

        user = result;
        roles = [...result.roles];
        rolesModal?.close();
        toast.success("Successfully updated the user's roles");
    };

    const openPasswordModal = () => {
        newPassword = "";
        confirmPassword = "";
        passwordModal?.open();
    };

    const setPassword = async (event: Event) => {
        event.preventDefault();
        if (!user || savingPassword) return;

        if (newPassword !== confirmPassword) {
            return toast.error("The passwords do not match");
        }

        savingPassword = true;
        const result = await star.adminSetPassword(user.id, newPassword);
        savingPassword = false;

        if ("error" in result) {
            return toast.error(result.message);
        }

        passwordModal?.close();
        toast.success("Successfully changed the user's password");
    };

    const unlock = async () => {
        if (!user || unlocking) return;

        unlocking = true;
        const result = await star.clearLockout(user.id);
        unlocking = false;

        if ("error" in result) {
            return toast.error(result.message);
        }

        user = result;
        toast.success(`Cleared ${result.username}'s failed sign in attempts`);
    };

    const exportData = async () => {
        if (!user) return;

        const result = await star.exportUser(user.id);
        if ("error" in result) {
            return toast.error(result.message);
        }

        downloadFile(result.blob, result.filename);
        toast.success(`Exported ${user.username}'s data`);
    };

    const invalidateSessions = async () => {
        if (!user) return;

        const result = await star.invalidateUserSessions(user.id);
        if ("error" in result) {
            toast.error(result.message);
            return false;
        }

        toast.success("Successfully signed out all user sessions");
    };

    const revokeTokens = async () => {
        if (!user) return;

        const result = await star.revokeTokens({ kind: "user", user: user.id });
        if ("error" in result) {
            toast.error(result.message);
            return false;
        }

        toast.success(`Revoked ${user.username}'s access tokens`);
    };

    const deleteUser = async () => {
        if (!user) return;

        const result = await star.deleteUser(user.id);
        if ("error" in result) {
            toast.error(result.message);
            return false;
        }

        toast.success(`Deleted ${user.username}`);
        goto("/admin/users");
    };
</script>

{#if user}
    <div class="flex flex-row items-center gap-4">
        <Avatar.Root class="size-14">
            <Avatar.Image
                src={user.profile_picture}
                alt="{user.username}'s profile picture"
            />
            <Avatar.Fallback class="text-lg">
                {user.username.charAt(0).toUpperCase()}
            </Avatar.Fallback>
        </Avatar.Root>

        <div class="min-w-0 flex-1">
            <h1 class="leading-8 text-2xl font-semibold truncate">
                {displayName}
            </h1>
            <div
                class="flex flex-wrap items-center gap-x-3 gap-y-1 text-sm text-muted mt-0.5"
            >
                <span class="font-mono text-xs">{user.id}</span>
                <span class="flex items-center gap-1">
                    <Icon icon="mail" />
                    {user.email ?? "No email"}
                    {#if user.email && !user.email_verified}
                        <Badge tone="warning" class="ml-1">unverified</Badge>
                    {/if}
                </span>
                <span>Joined {simpleDate(new Date(user.created_at))}</span>
                {#if user.lockout?.locked}
                    <Badge tone="danger">locked</Badge>
                {/if}
            </div>
        </div>

        <Menu items={menuItems} label="User actions" />
    </div>

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
        description="Roles grant this user permissions across Star and connected applications."
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

    <Modal bind:this={passwordModal} button>
        <form class="flex flex-col gap-6" onsubmit={setPassword}>
            <div class="flex flex-col gap-1 pr-8">
                <h1 class="text-xl font-semibold">Set password</h1>
                <h2 class="text-sm text-muted">
                    All of {user.username}'s sessions will be signed out and
                    they will need to log in again with the new password.
                </h2>
            </div>

            <div class="flex flex-col gap-5">
                <div class="grid gap-2">
                    <Label for="newPassword">New Password</Label>
                    <Input
                        bind:value={newPassword}
                        id="newPassword"
                        type="password"
                        autocomplete="new-password"
                        required
                    />
                </div>

                <div class="grid gap-2">
                    <Label for="confirmPassword">Confirm Password</Label>
                    <Input
                        bind:value={confirmPassword}
                        id="confirmPassword"
                        type="password"
                        autocomplete="new-password"
                        required
                    />
                </div>
            </div>

            <div class="flex items-center gap-2 justify-end">
                <Button
                    type="button"
                    variant="secondary"
                    onclick={() => passwordModal?.close()}
                >
                    Cancel
                </Button>
                <Button
                    type="submit"
                    loading={savingPassword}
                    disabled={!newPassword || !confirmPassword}
                >
                    Set password
                </Button>
            </div>
        </form>
    </Modal>

    <ConfirmModal
        bind:this={signOutModal}
        title="Sign out all sessions"
        confirmLabel="Sign out sessions"
        onconfirm={invalidateSessions}
    >
        Every session for
        <span class="text-foreground">{user.username}</span> will be invalidated
        and the access tokens issued to them stop validating within a few seconds.
    </ConfirmModal>

    <ConfirmModal
        bind:this={revokeTokensModal}
        title="Revoke access tokens"
        confirmLabel="Revoke tokens"
        onconfirm={revokeTokens}
    >
        Every access token issued to
        <span class="text-foreground">{user.username}</span> stops validating
        within a few seconds. Their sessions stay signed in and receive new
        tokens with their current roles on the next refresh.
    </ConfirmModal>

    <ConfirmModal
        bind:this={deleteUserModal}
        title="Delete user"
        confirmLabel="Delete user"
        confirmText={user.username}
        onconfirm={deleteUser}
    >
        All of <span class="text-foreground">{user.username}</span>'s sessions,
        emails and linked accounts will be permanently deleted. This action can
        not be reversed.
    </ConfirmModal>
{/if}
