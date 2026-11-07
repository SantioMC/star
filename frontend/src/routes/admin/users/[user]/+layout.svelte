<script lang="ts">
    import { goto } from "$app/navigation";
    import { resolve } from "$app/paths";
    import { page } from "$app/state";
    import * as star from "$lib/star";
    import Badge, { type BadgeTone } from "@/components/badge.svelte";
    import Button from "@/components/button.svelte";
    import ConfirmModal from "@/components/confirm-modal.svelte";
    import Icon from "@/components/icon.svelte";
    import Input from "@/components/input.svelte";
    import Menu, { type MenuItem } from "@/components/menu.svelte";
    import Modal from "@/components/modal.svelte";
    import Notice from "@/components/notice.svelte";
    import Tabs from "@/components/tabs.svelte";
    import * as Avatar from "@/components/ui/avatar";
    import InviteLinkModal from "@/components/invite-link-modal.svelte";
    import { Label } from "@/components/ui/label";
    import { hasPermission, user as currentUser } from "@/index";
    import { Resource, setResource } from "@/resource.svelte";
    import { downloadFile, formatTime, simpleDate } from "@/utils";
    import { onMount } from "svelte";
    import { toast } from "svelte-sonner";
    import ImpersonateModal from "./components/impersonate-modal.svelte";
    import RegenerateInviteModal from "./components/regenerate-invite-modal.svelte";
    import SuspendModal from "./components/suspend-modal.svelte";
    import UnsuspendModal from "./components/unsuspend-modal.svelte";

    let { children } = $props();

    const canWrite = hasPermission("admin:user:write");
    const canDelete = hasPermission("admin:user:delete");
    const canExport = hasPermission("admin:user:export");
    const canSuspend = hasPermission("admin:user:suspend");
    const canRevokeTokens = hasPermission("admin:token:revoke");
    const canReadAudit = hasPermission("admin:audit:read");
    const canImpersonate = hasPermission("admin:user:impersonate");
    const canInvite = hasPermission("admin:user:invite");

    const context = new Resource<star.AdminUser>();
    setResource(context);

    const user = $derived(context.current);
    const manageable = $derived(user?.manageable);

    let newPassword = $state("");
    let confirmPassword = $state("");
    let savingPassword = $state(false);

    let deleteUserModal = $state<ConfirmModal>();
    let signOutModal = $state<ConfirmModal>();
    let revokeTokensModal = $state<ConfirmModal>();
    let requirePasswordChangeModal = $state<ConfirmModal>();
    let convertModal = $state<ConfirmModal>();
    let passwordModal = $state<Modal>();
    let suspendModal = $state<SuspendModal>();
    let unsuspendModal = $state<UnsuspendModal>();
    let impersonateModal = $state<ImpersonateModal>();
    let regenerateInviteModal = $state<RegenerateInviteModal>();
    let inviteLinkModal = $state<InviteLinkModal>();

    const tabs = $derived.by(() => {
        if (!user) return [];

        const overview = resolve("/admin/users/[user]", { user: user.id });
        const sessions = resolve("/admin/users/[user]/sessions", {
            user: user.id,
        });
        const suspensions = resolve("/admin/users/[user]/suspensions", {
            user: user.id,
        });
        const audit = resolve("/admin/users/[user]/audit", { user: user.id });

        const items = [
            { label: "Overview", href: overview, icon: "user" },
            { label: "Sessions", href: sessions, icon: "device-desktop" },
            { label: "Suspensions", href: suspensions, icon: "user-off" },
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

    const badges = $derived.by(() => {
        const items: { label: string; tone: BadgeTone }[] = [];
        if (!user) return items;

        if (user.type === "service") {
            items.push({ label: "service", tone: "info" });
        }

        if (user.invite) {
            items.push({ label: "invited", tone: "info" });
        }

        if (user.email && !user.email_verified) {
            items.push({ label: "unverified", tone: "warning" });
        }

        if (user.lockout?.locked) {
            items.push({ label: "locked", tone: "danger" });
        }

        if (user.suspension) {
            items.push({ label: "suspended", tone: "danger" });
        }

        if (user.password_change_required) {
            items.push({ label: "password change required", tone: "warning" });
        }

        return items;
    });

    const menuItems = $derived.by(() => {
        const items: MenuItem[] = [];

        const isService = user?.type === "service";

        if (canWrite && manageable && !user?.invite && !isService) {
            items.push({
                label: "Set password",
                icon: "key",
                group: "account",
                onclick: openPasswordModal,
            });
        }

        const canRequirePasswordChange =
            canWrite &&
            manageable &&
            user?.has_password &&
            !user.password_change_required &&
            user.id !== $currentUser?.id;

        if (canRequirePasswordChange) {
            items.push({
                label: "Require password change",
                icon: "lock",
                group: "account",
                onclick: () => requirePasswordChangeModal?.open(),
            });
        }

        const canConvert =
            canWrite &&
            manageable &&
            !user?.invite &&
            user?.id !== $currentUser?.id;

        if (canConvert) {
            items.push({
                label: isService
                    ? "Convert to regular user"
                    : "Convert to service account",
                icon: isService ? "user" : "robot",
                group: "account",
                onclick: () => convertModal?.open(),
            });
        }

        if (canExport) {
            items.push({
                label: "Export data",
                icon: "download",
                group: "account",
                onclick: exportData,
            });
        }

        const canImpersonateUser =
            canImpersonate &&
            manageable &&
            !user?.suspension &&
            !user?.invite &&
            !isService &&
            user?.id !== $currentUser?.id;

        if (canImpersonateUser) {
            items.push({
                label: "Impersonate",
                icon: "spy",
                group: "access",
                onclick: () => impersonateModal?.open(),
            });
        }

        if (canDelete && manageable) {
            items.push({
                label: "Sign out all sessions",
                icon: "logout",
                group: "access",
                onclick: () => signOutModal?.open(),
            });
        }

        if (canRevokeTokens && manageable) {
            items.push({
                label: "Revoke access tokens",
                icon: "ban",
                group: "access",
                onclick: () => revokeTokensModal?.open(),
            });
        }

        if (canSuspend && manageable && user?.suspension) {
            items.push({
                label: "Unsuspend user",
                icon: "user-check",
                group: "moderation",
                onclick: () => unsuspendModal?.open(),
            });
        } else if (canSuspend && manageable && user?.id !== $currentUser?.id) {
            items.push({
                label: "Suspend user",
                icon: "user-off",
                danger: true,
                group: "moderation",
                onclick: () => suspendModal?.open(),
            });
        }

        if (canDelete && manageable) {
            items.push({
                label: "Delete user",
                icon: "trash",
                danger: true,
                group: "moderation",
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

    const convertUser = async () => {
        if (!user) return;

        const type = user.type === "service" ? "regular" : "service";

        const result = await star.modifyUser(user.id, { type });
        if ("error" in result) {
            toast.error(result.message);
            return false;
        }

        context.current = result;
        toast.success(`Converted ${result.username}`);
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
                </span>
                <span>
                    {user.invite ? "Invited" : "Joined"}
                    {simpleDate(new Date(user.created_at))}
                </span>
            </div>
            {#if badges.length > 0}
                <div class="flex flex-wrap items-center gap-1.5 mt-1.5">
                    {#each badges as badge (badge.label)}
                        <Badge tone={badge.tone}>{badge.label}</Badge>
                    {/each}
                </div>
            {/if}
        </div>

        <Menu items={menuItems} label="User actions" />
    </div>

    {#if user.suspension}
        {@const suspension = user.suspension}
        {@const suspender = suspension.suspended_by_name ?? "a deleted user"}
        <Notice
            tone="danger"
            icon="user-off"
            title={suspension.expires_at
                ? `Suspended by ${suspender} until ${formatTime(new Date(suspension.expires_at))}.`
                : `Permanently suspended by ${suspender}.`}
        >
            {suspension.reason}

            {#snippet actions()}
                {#if canSuspend && manageable}
                    <Button
                        variant="secondary"
                        onclick={() => unsuspendModal?.open()}
                    >
                        Unsuspend
                    </Button>
                {/if}
            {/snippet}
        </Notice>
    {/if}

    {#if user.invite}
        {@const inviter = user.invite.invited_by_name ?? "a deleted user"}
        <Notice
            tone="info"
            icon="mail-forward"
            title={`Invited by ${inviter}, expires ${formatTime(new Date(user.invite.expires_at))}.`}
        >
            This user can not sign in until they accept the invite, and is
            deleted if it expires.

            {#snippet actions()}
                {#if canInvite && manageable}
                    <Button
                        variant="secondary"
                        onclick={() => regenerateInviteModal?.open()}
                    >
                        Regenerate link
                    </Button>
                {/if}
            {/snippet}
        </Notice>
    {/if}

    <Tabs items={tabs} label="User sections" />

    {@render children()}

    <ImpersonateModal bind:this={impersonateModal} {user} />

    <RegenerateInviteModal
        bind:this={regenerateInviteModal}
        {user}
        onregenerated={async (link) => {
            inviteLinkModal?.show(link, user.username);
            context.current = (await star.getUser(user.id)) ?? user;
        }}
    />

    <InviteLinkModal bind:this={inviteLinkModal} />

    <SuspendModal
        bind:this={suspendModal}
        {user}
        onsuspended={(updated) => (context.current = updated)}
    />

    <UnsuspendModal
        bind:this={unsuspendModal}
        {user}
        onunsuspended={(updated) => (context.current = updated)}
    />

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
        bind:this={convertModal}
        title={user.type === "service"
            ? "Convert to regular user"
            : "Convert to service account"}
        confirmLabel="Convert"
        onconfirm={convertUser}
    >
        {#if user.type === "service"}
            <span class="text-foreground">{user.username}</span> will be able to
            sign in with any credentials on this account.
            {#if !user.has_password}
                They have no password yet, so set one after converting.
            {/if}
        {:else}
            <span class="text-foreground">{user.username}</span> will be signed
            out everywhere, and will not be able to sign in again. Their
            credentials are kept in case they are converted back.
        {/if}
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
