<script lang="ts">
  import * as star from "$lib/star";
  import Button from "@/components/button.svelte";
  import InviteExpiryField, {
    defaultInviteExpiry,
  } from "@/components/invite-expiry-field.svelte";
  import Modal from "@/components/modal.svelte";
  import { validateUsername, type UsernameRules } from "@/username";
  import { toast } from "svelte-sonner";
  import UserProfileFields, {
    emptyProfile,
    toProfileRequest,
  } from "./user-profile-fields.svelte";

  const {
    oninvited,
  }: {
    oninvited: (link: star.InviteLink, username: string) => void;
  } = $props();

  let modal = $state<Modal>();
  let profile = $state(emptyProfile());
  let expiry = $state(defaultInviteExpiry);
  let saving = $state(false);

  let usernameRules = $state<UsernameRules | null>(null);
  let birthdayTracked = $state(false);

  const usernameIssue = $derived(
    usernameRules && profile.username
      ? validateUsername(profile.username, usernameRules)
      : null,
  );

  export const open = async () => {
    profile = emptyProfile();
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
      ...toProfileRequest(profile),
      expires_in: Number(expiry),
    });
    saving = false;

    if ("error" in result) {
      return toast.error(result.message);
    }

    modal?.close();
    oninvited(result, profile.username);
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
      <UserProfileFields bind:profile {usernameIssue} {birthdayTracked} />
      <InviteExpiryField bind:value={expiry} />
    </div>

    <div class="flex items-center gap-2 justify-end">
      <Button type="button" variant="secondary" onclick={() => modal?.close()}>
        Cancel
      </Button>
      <Button
        type="submit"
        loading={saving}
        disabled={!profile.username.trim() || !profile.email}
      >
        Create invite
      </Button>
    </div>
  </form>
</Modal>
