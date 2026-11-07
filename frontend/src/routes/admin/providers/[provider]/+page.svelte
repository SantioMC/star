<script lang="ts">
  import { branding } from "@/branding";
  import * as star from "$lib/star";
  import Badge from "@/components/badge.svelte";
  import Button from "@/components/button.svelte";
  import Card from "@/components/card.svelte";
  import Checkbox from "@/components/checkbox.svelte";
  import Copy from "@/components/copy.svelte";
  import Icon from "@/components/icon.svelte";
  import Input from "@/components/input.svelte";
  import Modal from "@/components/modal.svelte";
  import { Label } from "@/components/ui/label";
  import { hasPermission } from "@/index";
  import { getResource } from "@/resource.svelte";
  import { toast } from "svelte-sonner";

  const canWrite = hasPermission("admin:provider:write");

  const resource = getResource<star.SocialProviderConfig>();
  const provider = $derived(resource.current);

  let icon = $state("");
  let order = $state("");
  let trustEmail = $state(false);
  let enabled = $state(true);
  let saving = $state(false);

  let clientId = $state("");
  let clientSecret = $state("");
  let savingCredentials = $state(false);

  let credentialsModal = $state<Modal>();

  const resetForm = (data: star.SocialProviderConfig) => {
    icon = data.icon;
    order = String(data.order);
    trustEmail = data.trust_email;
    enabled = data.enabled;
  };

  resetForm(resource.current!);

  const updateDisplay = async (event: Event) => {
    event.preventDefault();
    if (!provider || saving) return;

    saving = true;
    const result = await star.updateSocialProvider(provider.id, {
      icon: icon.trim() || undefined,
      order: Number(order),
      trust_email: trustEmail,
      enabled,
    });
    saving = false;

    if ("error" in result) {
      resetForm(provider);
      return toast.error(result.message);
    }

    resource.current = result;
    toast.success("Successfully updated provider");
  };

  const openCredentialsModal = () => {
    clientId = "";
    clientSecret = "";
    credentialsModal?.open();
  };

  const updateCredentials = async (event: Event) => {
    event.preventDefault();
    if (!provider || savingCredentials) return;

    const changes: star.UpdateSocialProviderRequest = {};
    if (clientId.trim()) changes.client_id = clientId.trim();
    if (clientSecret) changes.client_secret = clientSecret;

    if (Object.keys(changes).length === 0) {
      credentialsModal?.close();
      return;
    }

    savingCredentials = true;
    const result = await star.updateSocialProvider(provider.id, changes);
    savingCredentials = false;

    if ("error" in result) {
      return toast.error(result.message);
    }

    resource.current = result;
    credentialsModal?.close();
    toast.success("Successfully updated credentials");
  };
</script>

{#if provider}
  <Card
    title="Credentials"
    icon="key"
    description="The OAuth client issued to {$branding.name} by this provider."
    class="flex flex-col gap-4"
  >
    {#snippet actions()}
      {#if canWrite}
        <Button variant="secondary" onclick={openCredentialsModal}>
          Update
        </Button>
      {/if}
    {/snippet}

    <dl class="grid gap-4 sm:grid-cols-2">
      <div class="min-w-0">
        <dt class="text-xs text-muted select-none">Client ID</dt>
        <dd class="mt-1 flex items-center gap-2 font-mono text-sm break-all">
          {provider.client_id}
          <Copy content={provider.client_id} class="text-muted" />
        </dd>
      </div>
      <div class="min-w-0">
        <dt class="text-xs text-muted select-none">Client Secret</dt>
        <dd class="mt-1">
          {#if provider.has_client_secret}
            <Badge tone="success">configured</Badge>
          {:else}
            <Badge tone="warning">missing</Badge>
          {/if}
        </dd>
      </div>
      {#if provider.scopes?.length}
        <div class="min-w-0 sm:col-span-2">
          <dt class="text-xs text-muted select-none">Scopes</dt>
          <dd class="mt-1.5 flex flex-wrap gap-1">
            {#each provider.scopes as scope (scope)}
              <Badge>{scope}</Badge>
            {/each}
          </dd>
        </div>
      {/if}
    </dl>
  </Card>

  <Card
    title="Provider Settings"
    icon="settings"
    description="How this provider appears on the login page and how its accounts are trusted."
  >
    <form onsubmit={updateDisplay} class="mt-6">
      <div class="flex flex-col gap-5">
        <div class="grid gap-2">
          <Label for="icon">Icon</Label>
          <div class="flex items-center gap-3">
            <div
              class="flex size-9 shrink-0 items-center justify-center rounded-md border border-background-200"
            >
              <Icon icon={icon || "plug-connected"} class="text-lg" />
            </div>
            <Input
              bind:value={icon}
              id="icon"
              class="font-mono"
              placeholder="brand-discord"
              disabled={!canWrite}
            />
          </div>
          <p class="text-xs text-muted select-none">
            Specify the name of a <a
              href="https://tabler.io/icons"
              target="_blank"
              class="text-blue-400/70 hover:text-blue-400"
            >
              Tabler icon
            </a>
          </p>
        </div>

        <div class="grid gap-2">
          <Label for="order">Display order</Label>
          <Input
            bind:value={order}
            id="order"
            type="number"
            min="0"
            class="font-mono"
            disabled={!canWrite}
          />
          <p class="text-xs text-muted select-none">
            Lower numbers appear first
          </p>
        </div>

        <label
          class="flex items-start gap-3 rounded-lg border border-background-200 p-3 transition-colors has-enabled:cursor-pointer has-enabled:hover:bg-hover"
        >
          <Checkbox
            class="mt-0.5"
            bind:checked={enabled}
            disabled={!canWrite}
          />
          <div class="flex flex-col gap-0.5">
            <span class="text-sm select-none">Enabled</span>
            <span class="text-sm text-muted select-none">
              Lets users sign in and connect their accounts with this provider.
              While off, users who can only sign in through it are locked out of
              their accounts.
            </span>
          </div>
        </label>

        <label
          class="flex items-start gap-3 rounded-lg border border-background-200 p-3 transition-colors has-checked:border-red-400/40 has-checked:bg-red-400/5 has-enabled:cursor-pointer has-enabled:hover:bg-hover"
        >
          <Checkbox
            class="mt-0.5"
            bind:checked={trustEmail}
            disabled={!canWrite}
          />
          <div class="flex flex-col gap-0.5">
            <span class="text-sm select-none"> Trust verified emails </span>
            <span class="text-sm text-muted select-none">
              New accounts created through this provider keep the email it
              reports as verified. Applications may link accounts by verified
              email, so only enable this for providers that confirm email
              ownership themselves.
            </span>
          </div>
        </label>
      </div>

      {#if canWrite}
        <div class="mt-4 flex items-center gap-2 justify-end">
          <Button type="submit" loading={saving}>Update</Button>
        </div>
      {/if}
    </form>
  </Card>

  <Modal bind:this={credentialsModal} button>
    <form class="flex flex-col gap-6" onsubmit={updateCredentials}>
      <div class="flex flex-col gap-1 pr-8">
        <h1 class="text-xl font-semibold">Update credentials</h1>
        <h2 class="text-sm text-muted">
          Leave a field empty to keep its current value.
        </h2>
      </div>

      <div class="flex flex-col gap-5">
        <div class="grid gap-2">
          <Label for="newClientId">Client ID</Label>
          <Input
            bind:value={clientId}
            id="newClientId"
            class="font-mono"
            autocomplete="off"
            placeholder={provider.client_id}
          />
        </div>
        <div class="grid gap-2">
          <Label for="newClientSecret">Client secret</Label>
          <Input
            bind:value={clientSecret}
            id="newClientSecret"
            type="password"
            autocomplete="off"
          />
        </div>
      </div>

      <div class="flex items-center gap-2 justify-end">
        <Button
          type="button"
          variant="secondary"
          onclick={() => credentialsModal?.close()}
        >
          Cancel
        </Button>
        <Button type="submit" loading={savingCredentials}>
          Update credentials
        </Button>
      </div>
    </form>
  </Modal>
{/if}
