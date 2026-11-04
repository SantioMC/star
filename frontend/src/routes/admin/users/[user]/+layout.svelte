<script lang="ts">
    import { goto } from "$app/navigation";
    import { resolve } from "$app/paths";
    import { page } from "$app/state";
    import * as star from "$lib/star";
    import Badge from "@/components/badge.svelte";
    import Button from "@/components/button.svelte";
    import ConfirmModal from "@/components/confirm-modal.svelte";
    import Icon from "@/components/icon.svelte";
    import Input from "@/components/input.svelte";
    import Menu, { type MenuItem } from "@/components/menu.svelte";
    import Modal from "@/components/modal.svelte";
    import Tabs from "@/components/tabs.svelte";
    import * as Avatar from "@/components/ui/avatar";
    import { Label } from "@/components/ui/label";
    import { hasPermission, user as currentUser } from "@/index";
    import { Resource, setResource } from "@/resource.svelte";
    import { downloadFile, simpleDate } from "@/utils";
    import { onMount } from "svelte";
    import { toast } from "svelte-sonner";

    let { children } = $props();

    const canWrite = hasPermission("admin:user:write");
    const canDelete = hasPermission("admin:user:delete");
    const canExport = hasPermission("admin:user:export");
    const canRevokeTokens = hasPermission("admin:token:revoke");
    const canReadAudit = hasPermission("admin:audit:read");

    const context = new Resource<star.AdminUser>();
    setResource(context);

    const user = $derived(context.current);

    let newPassword = $state("");
    let confirmPassword = $state("");
    let savingPassword = $state(false);

    let deleteUserModal = $state<ConfirmModal>();
    let signOutModal = $state<ConfirmModal>();
    let revokeTokensModal = $state<ConfirmModal>();
    let requirePasswordChangeModal = $state<ConfirmModal>();
    let passwordModal = $state<Modal>();

    const tabs = $derived.by(() => {
        if (!user) return [];

        const overview = resolve("/admin/users/[user]", { user: user.id });
        const sessions = resolve("/admin/users/[user]/sessions", {
            user: user.id,
        });
        const audit = resolve("/admin/users/[user]/audit", { user: user.id });

        const items = [
            { label: "Overview", href: overview, icon: "user" },
            { label: "Sessions", href: sessions, icon: "device-desktop" },
        ];

        if (canReadAudit) {
            items.push({ label: "Audit Log", href: audit, icon: "history" });
        }

        return items;
    });

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

        const canRequirePasswordChange =
            canWrite &&
            user?.has_password &&
            !user.password_change_required &&
            user.id !== $currentUser?.id;

        if (canRequirePasswordChange) {
            items.push({
                label: "Require password change",
                icon: "lock",
                onclick: () => requirePasswordChangeModal?.open(),
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

    onMount(async () => {
        if (!hasPermission("admin:user:read")) return goto("/");

        const data = await star.getUser(page.params.user!);
        if (!data) return goto("/admin/users");

        context.current = data;
    });

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

    const requirePasswordChange = async () => {
        if (!user) return;

        const result = await star.requirePasswordChange(user.id);
        if ("error" in result) {
            toast.error(result.message);
            return false;
        }

        context.current = result;
        toast.success(`${result.username} must change their password`);
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
                {#if user.password_change_required}
                    <Badge tone="warning">password change required</Badge>
                {/if}
            </div>
        </div>

        <Menu items={menuItems} label="User actions" />
    </div>

    <Tabs items={tabs} label="User sections" />

    {@render children()}

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
        bind:this={requirePasswordChangeModal}
        title="Require password change"
        confirmLabel="Require change"
        onconfirm={requirePasswordChange}
    >
        <span class="text-foreground">{user.username}</span> will be signed out
        everywhere and must choose a new password the next time they sign in.
    </ConfirmModal>

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
