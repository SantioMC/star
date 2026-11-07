<script lang="ts">
  import * as star from "$lib/star";
  import BirthdayField from "@/components/birthday-field.svelte";
  import Button from "@/components/button.svelte";
  import Input from "@/components/input.svelte";
  import InviteExpiryField, {
    defaultInviteExpiry,
  } from "@/components/invite-expiry-field.svelte";
  import Modal from "@/components/modal.svelte";
  import { Label } from "@/components/ui/label";
  import { validateUsername, type UsernameRules } from "@/username";
  import { toast } from "svelte-sonner";

  const {
    oninvited,
  }: {
    oninvited: (link: star.InviteLink, username: string) => void;
  } = $props();

  let modal = $state<Modal>();
  let username = $state("");
  let email = $state("");
  let firstName = $state("");
  let lastName = $state("");
  let birthday = $state("");
  let expiry = $state(defaultInviteExpiry);
  let saving = $state(false);

  let usernameRules = $state<UsernameRules | null>(null);
  let birthdayTracked = $state(false);

  const usernameIssue = $derived(
    usernameRules && username
      ? validateUsername(username, usernameRules)
      : null,
  );

  export const open = async () => {
    username = "";
    email = "";
    firstName = "";
    lastName = "";
    birthday = "";
    expiry = defaultInviteExpiry;
    modal?.open();

    const requirements = await star.getRegisterRequirements();
    usernameRules = requirements?.username ?? null;
    birthdayTracked = requirements?.birthday_tracked ?? false;
  };

  const invite = async (event: Event) => {
    event.preventDefault();
    if (saving || usernameIssue) return;

    saving = true;
    const result = await star.createInvite({
      username,
      email,
      first_name: firstName ? firstName : undefined,
      last_name: lastName ? lastName : undefined,
      birthday: birthday ? birthday : undefined,
      expires_in: Number(expiry),
    });
    saving = false;

    if ("error" in result) {
      return toast.error(result.message);
    }

    modal?.close();
    oninvited(result, username);
  };
</script>

<Modal bind:this={modal} button>
  <form class="flex flex-col gap-6" onsubmit={invite}>
    <div class="flex flex-col gap-1 pr-8">
      <h1 class="text-xl font-semibold">Invite user</h1>
      <h2 class="text-sm text-muted">
        Create a user manually and share the invite link for them to setup their
        account
      </h2>
    </div>

    <div class="flex flex-col gap-5">
      <div class="grid gap-2">
        <Label for="inviteUsername" class="gap-0.5">
          Username
          <span class="text-red-400" aria-hidden="true">*</span>
        </Label>
        <Input
          bind:value={username}
          id="inviteUsername"
          type="text"
          autocomplete="off"
          aria-invalid={usernameIssue != null}
          aria-describedby={usernameIssue ? "invite-username-issue" : undefined}
          required
        />
        {#if usernameIssue}
          <p id="invite-username-issue" class="text-xs text-red-300">
            {usernameIssue}
          </p>
        {/if}
      </div>

      <div class="grid gap-2">
        <Label for="inviteEmail" class="gap-0.5">
          Email
          <span class="text-red-400" aria-hidden="true">*</span>
        </Label>
        <Input
          bind:value={email}
          id="inviteEmail"
          type="email"
          maxlength={255}
          autocomplete="off"
          required
        />
      </div>

      <div class="grid gap-5 sm:grid-cols-2">
        <div class="grid gap-2">
          <Label for="inviteFirstName">First Name</Label>
          <Input
            bind:value={firstName}
            id="inviteFirstName"
            type="text"
            maxlength={255}
            autocomplete="off"
          />
        </div>

        <div class="grid gap-2">
          <Label for="inviteLastName">Last Name</Label>
          <Input
            bind:value={lastName}
            id="inviteLastName"
            type="text"
            maxlength={255}
            autocomplete="off"
          />
        </div>
      </div>

      {#if birthdayTracked}
        <BirthdayField bind:value={birthday} />
      {/if}

      <InviteExpiryField bind:value={expiry} />
    </div>

    <div class="flex items-center gap-2 justify-end">
      <Button type="button" variant="secondary" onclick={() => modal?.close()}>
        Cancel
      </Button>
      <Button
        type="submit"
        loading={saving}
        disabled={!username.trim() || !email}
      >
        Create invite
      </Button>
    </div>
  </form>
</Modal>
