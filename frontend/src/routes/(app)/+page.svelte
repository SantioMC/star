<script lang="ts">
    import * as Avatar from "@/components/ui/avatar";
    import { Label } from "@/components/ui/label";
    import { hasPermission, user } from "@/index";
    import * as star from "@/star";
    import { downloadFile, simpleDate } from "@/utils";
    import { onMount } from "svelte";
    import { toast } from "svelte-sonner";
    import Card from "@/components/card.svelte";
    import Input from "@/components/input.svelte";
    import Button from "@/components/button.svelte";
    import Icon from "@/components/icon.svelte";
    import Modal from "@/components/modal.svelte";
    import Menu, { type MenuItem } from "@/components/menu.svelte";
    import { goto } from "$app/navigation";
    import { syncPasskeys } from "@/passkey";
    import { resolve } from "$app/paths";

    let username = $state<string>("");
    let firstName = $state<string>("");
    let lastName = $state<string>("");

    let addEmailModal = $state<Modal>();
    let newEmail = $state<string>("");
    let addingEmail = $state<boolean>(false);

    let deleteEmailModal = $state<Modal>();
    let pendingEmail = $state<star.Email | null>(null);
    let deletingEmail = $state<boolean>(false);

    let pendingAction = $state<string | null>(null);

    const emails = $derived(
        [...($user?.emails ?? [])].sort((a, b) => {
            if (a.primary !== b.primary) return a.primary ? -1 : 1;
            return Date.parse(a.created_at) - Date.parse(b.created_at);
        }),
    );

    onMount(async () => {
        username = $user!.username;
        firstName = $user!.first_name;
        lastName = $user!.last_name;
    });

    const setEmails = (emails: star.Email[]) => {
        user.update((current) => {
            if (!current) return current;
            return { ...current, emails };
        });
    };

    const updateDetails = async (e: Event) => {
        e.preventDefault();

        const response = await star.updateUser({
            username,
            first_name: firstName,
            last_name: lastName,
        });

        if ("error" in response) {
            return toast.error(response.message);
        }

        toast.success("Successfully updated your details!");
        syncPasskeys();

        user.update((user) => {
            if (!user) return user;

            user.username = username;
            user.first_name = firstName;
            user.last_name = lastName;
            return user;
        });
    };

    const addEmail = async (e: Event) => {
        e.preventDefault();

        const email = newEmail.trim();
        if (!email || addingEmail) return;

        addingEmail = true;
        const response = await star.addEmail(email);
        addingEmail = false;

        if ("error" in response) {
            return toast.error(response.message);
        }

        setEmails([...($user?.emails ?? []), response]);
        newEmail = "";
        addEmailModal?.close();

        toast.success(
            `A verification email has been sent to ${response.email}`,
        );
    };

    const confirmDelete = (email: star.Email) => {
        pendingEmail = email;
        deleteEmailModal?.open();
    };

    const deleteEmail = async () => {
        const email = pendingEmail;
        if (!email || deletingEmail) return;

        deletingEmail = true;
        const response = await star.deleteEmail(email.id);
        deletingEmail = false;

        if ("error" in response) {
            return toast.error(response.message);
        }

        setEmails(($user?.emails ?? []).filter((it) => it.id != email.id));
        deleteEmailModal?.close();

        toast.success(`Removed ${email.email} from your account`);
    };

    const makePrimary = async (email: star.Email) => {
        if (pendingAction) return;

        pendingAction = email.id;
        const response = await star.setPrimaryEmail(email.id);
        pendingAction = null;

        if ("error" in response) {
            return toast.error(response.message);
        }

        user.update((current) => {
            if (!current) return current;

            return {
                ...current,
                email: email.email,
                email_verified: email.verified,
                emails: current.emails.map((it) => ({
                    ...it,
                    primary: it.id == email.id,
                })),
            };
        });

        toast.success(`${email.email} is now your primary email`);
    };

    const resendVerification = async (email: star.Email) => {
        if (pendingAction) return;

        pendingAction = email.id;
        const response = await star.resendVerification(email.email);
        pendingAction = null;

        if ("error" in response) {
            return toast.error(response.message);
        }

        toast.success(`A verification email has been sent to ${email.email}`);
    };

    const exportData = async () => {
        const result = await star.exportSelf();
        if ("error" in result) {
            return toast.error(result.message);
        }

        downloadFile(result.blob, result.filename);
        toast.success("Your data export has been downloaded");
    };

    const canExport = hasPermission("user:data:export");
    const canViewAdmin = hasPermission("admin:user:read");

    const adminPage = $derived(
        resolve("/admin/users/[user]", { user: $user?.id ?? "" }),
    );

    const menuItems: MenuItem[] = $derived(
        canExport
            ? [
                  ...(canViewAdmin
                      ? [
                            {
                                label: "View admin page",
                                icon: "external-link",
                                onclick: () => goto(adminPage),
                            },
                        ]
                      : []),
                  {
                      label: "Export my data",
                      icon: "download",
                      onclick: exportData,
                  },
              ]
            : [],
    );

    const emailColor = (email: star.Email): string => {
        if (!email.verified) return "#bb9955";
        if (email.primary) return "#99aadd";
        return "#99dd99";
    };
</script>

{#if $user}
    <div class="flex items-center justify-between">
        <div>
            <h1 class="leading-8 text-2xl font-semibold">
                Manage your account
            </h1>
            <h2 class="leading-6 text-muted text-sm mt-1">
                Manage your account settings and preferences
            </h2>
        </div>

        <div class="flex items-center gap-2 flex-row">
            {#if canViewAdmin && !canExport}
                <Button href={adminPage} variant="ghost">
                    View Admin Page
                    <Icon icon="external-link" />
                </Button>
            {/if}

            <Menu items={menuItems} label="Account actions" />
        </div>
    </div>

    <Card class="min-w-1/3 w-full md:w-auto">
        <form onsubmit={updateDetails}>
            <div class="flex flex-col gap-8">
                <div class="flex gap-3 items-center text-xl">
                    <Avatar.Root class="size-12">
                        <Avatar.Image
                            src={$user.picture}
                            alt="{username}'s Profile Picture"
                        />
                        <Avatar.Fallback class="text-lg">
                            {username.charAt(0).toUpperCase()}
                        </Avatar.Fallback>
                    </Avatar.Root>
                    <div class="flex flex-col">
                        <span class="font-medium">
                            {#if firstName && lastName}
                                {firstName} {lastName}
                            {:else}
                                {username}
                            {/if}
                        </span>
                        <div
                            class="flex flex-row items-center gap-1 text-sm text-muted"
                        >
                            <span>
                                Joined {simpleDate(new Date($user.created_at))}
                            </span>
                        </div>
                    </div>
                </div>

                <div class="flex flex-col gap-5">
                    <div class="grid gap-2">
                        <Label for="username">Username</Label>
                        <Input
                            bind:value={username}
                            id="username"
                            type="text"
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
                            />
                        </div>

                        <div class="flex-1/2 gap-2 grid">
                            <Label for="familyName">Family Name</Label>
                            <Input
                                bind:value={lastName}
                                id="familyName"
                                type="text"
                            />
                        </div>
                    </div>
                </div>
            </div>
            <div class="mt-4 flex items-center gap-2 justify-end">
                <Button type="submit" onclick={updateDetails}>Update</Button>
            </div>
        </form>
    </Card>

    <Card class="flex flex-col gap-2">
        <div class="flex flex-row items-center justify-between">
            <h2 class="text-sm font-medium">Email addresses</h2>
            <Button variant="secondary" onclick={() => addEmailModal?.open()}>
                Add email
            </Button>
        </div>
        <table>
            <tbody>
                {#each emails as email (email.id)}
                    <tr style:--color={emailColor(email)}>
                        <td class="flex flex-row items-center gap-2">
                            <div
                                class="size-2 bg-(--color) rounded-full translate-0.5"
                            ></div>
                            <span>{email.email}</span>
                        </td>
                        <td class="pr-2 font-mono text-xs">
                            <div class="flex justify-end">
                                {#if email.primary}
                                    <span class="text-green-300/75 px-3">
                                        primary
                                    </span>
                                {:else if email.verified}
                                    <Button
                                        variant="ghost"
                                        class="text-xs text-orange-300"
                                        disabled={pendingAction != null}
                                        onclick={() => makePrimary(email)}
                                    >
                                        make primary
                                    </Button>
                                {:else if !email.verified}
                                    <Button
                                        variant="ghost"
                                        class="text-xs text-yellow-300"
                                        disabled={pendingAction != null}
                                        onclick={() =>
                                            resendVerification(email)}
                                    >
                                        resend verification
                                    </Button>
                                {/if}
                            </div>
                        </td>
                        <td class="w-8">
                            <Button
                                disabled={email.primary}
                                variant="icon"
                                aria-label="Remove {email.email}"
                                onclick={() => confirmDelete(email)}
                            >
                                <Icon icon="trash" class="text-red-400" />
                            </Button>
                        </td>
                    </tr>
                {/each}
            </tbody>
        </table>
    </Card>

    <Modal bind:this={addEmailModal} button>
        <form class="flex flex-col gap-6" onsubmit={addEmail}>
            <div class="flex flex-col gap-1">
                <h1 class="text-xl font-semibold">Add an email address</h1>
                <h2 class="text-sm text-muted">
                    We'll send a verification link to this address before it can
                    be used on your account.
                </h2>
            </div>

            <div class="grid gap-2">
                <Label for="newEmail">Email address</Label>
                <Input
                    bind:value={newEmail}
                    id="newEmail"
                    type="email"
                    placeholder="you@example.com"
                    required
                />
            </div>

            <div class="flex items-center gap-2 justify-end">
                <Button
                    type="button"
                    variant="secondary"
                    onclick={() => addEmailModal?.close()}
                >
                    Cancel
                </Button>
                <Button
                    type="submit"
                    disabled={addingEmail || !newEmail.trim()}
                >
                    {#if addingEmail}
                        <Icon icon="loader-2" class="animate-spin" />
                    {/if}

                    Add email
                </Button>
            </div>
        </form>
    </Modal>

    <Modal bind:this={deleteEmailModal} button>
        <div class="flex flex-col gap-6">
            <div class="flex flex-col gap-1">
                <h1 class="text-xl font-semibold">Remove email address</h1>
                <h2 class="text-sm text-muted">
                    <span class="text-foreground">{pendingEmail?.email}</span>
                    will be removed from your account. This action can not be reversed.
                </h2>
            </div>

            <div class="flex items-center gap-2 justify-end">
                <Button
                    variant="secondary"
                    onclick={() => deleteEmailModal?.close()}
                >
                    Cancel
                </Button>
                <Button
                    variant="danger"
                    loading={deletingEmail}
                    onclick={deleteEmail}
                >
                    Remove email
                </Button>
            </div>
        </div>
    </Modal>
{/if}
