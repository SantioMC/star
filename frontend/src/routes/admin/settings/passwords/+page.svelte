<script lang="ts">
  import { goto } from "$app/navigation";
  import * as star from "$lib/star";
  import Button from "@/components/button.svelte";
  import Card from "@/components/card.svelte";
  import ConfirmModal from "@/components/confirm-modal.svelte";
  import Icon from "@/components/icon.svelte";
  import NumberField from "@/components/number-field.svelte";
  import Segmented from "@/components/segmented.svelte";
  import SettingsForm from "@/components/settings-form.svelte";
  import { clearSession, hasPermission } from "@/index";
  import { PASSWORD_ALGORITHMS } from "@/passwords";
  import type {
    PasswordAlgorithm,
    PasswordAlgorithmStatus,
    PasswordSettings,
  } from "@/star";
  import { onMount } from "svelte";
  import { toast } from "svelte-sonner";
  import { getSettingsContext } from "../context.svelte";

  type PolicyKey = Exclude<keyof PasswordSettings, "algorithm">;

  interface PolicyField {
    key: PolicyKey;
    label: string;
    min: number;
    max: number;
  }

  const lengthFields: PolicyField[] = [
    { key: "min_length", label: "Minimum Length", min: 8, max: 128 },
    { key: "max_length", label: "Maximum Length", min: 8, max: 512 },
  ];

  const characterFields: PolicyField[] = [
    { key: "min_symbols", label: "Minimum Symbols", min: 0, max: 32 },
    { key: "min_upper", label: "Minimum Uppercase", min: 0, max: 32 },
    { key: "min_lower", label: "Minimum Lowercase", min: 0, max: 32 },
  ];

  const policyKeys = [...lengthFields, ...characterFields].map((it) => it.key);

  const settings = getSettingsContext();

  const canRequireForAll = hasPermission("admin:password:require");

  let requireForAllModal = $state<ConfirmModal>();

  const algorithmOptions = Object.entries(PASSWORD_ALGORITHMS).map(
    ([value, { label }]) => ({ value: value as PasswordAlgorithm, label }),
  );

  const toInputs = (policy: PasswordSettings) =>
    Object.fromEntries(
      policyKeys.map((key) => [key, String(policy[key])]),
    ) as Record<PolicyKey, string>;

  let inputs = $state(toInputs(settings.current!.password_settings));
  let algorithm = $state<PasswordAlgorithm>(
    settings.current!.password_settings.algorithm,
  );

  let status = $state<PasswordAlgorithmStatus | null>(null);
  let statusFailed = $state(false);

  const total = $derived(
    status?.counts.reduce((sum, entry) => sum + entry.count, 0) ?? 0,
  );

  const migrated = $derived(
    status?.counts.find((entry) => entry.algorithm === status?.algorithm)
      ?.count ?? 0,
  );

  const percentOf = (count: number) =>
    total === 0 ? 100 : Math.floor((count / total) * 100);

  const loadStatus = async () => {
    status = await star.getPasswordAlgorithmStatus();
    statusFailed = status === null;
  };

  onMount(loadStatus);

  const parsed = $derived(
    Object.fromEntries(
      policyKeys.map((key) => [key, Number(inputs[key])]),
    ) as Record<PolicyKey, number>,
  );

  const dirty = $derived(
    algorithm !== settings.current!.password_settings.algorithm ||
      policyKeys.some(
        (key) => parsed[key] !== settings.current!.password_settings[key],
      ),
  );

  const issue = $derived.by(() => {
    if (policyKeys.some((key) => !Number.isInteger(parsed[key]))) {
      return "Every field must be a whole number";
    }

    if (parsed.max_length < parsed.min_length) {
      return "The maximum length must be greater than the minimum length";
    }

    const required = parsed.min_symbols + parsed.min_upper + parsed.min_lower;

    if (required > parsed.min_length) {
      return "The required characters add up to more than the minimum length, no password could satisfy this policy";
    }

    if (algorithm === "bcrypt" && parsed.max_length > 72) {
      return "The maximum configured password is too long, it must be less than or equal to 72 characters";
    }

    return null;
  });

  const save = async () => {
    const saved = await settings.update({
      password_settings: { ...parsed, algorithm },
    });

    if (saved) {
      await loadStatus();
    }
  };

  const discard = () => {
    inputs = toInputs(settings.current!.password_settings);
    algorithm = settings.current!.password_settings.algorithm;
  };

  const requireForAll = async () => {
    const result = await star.requirePasswordChangeForAll();
    if ("error" in result) {
      toast.error(result.message);
      return false;
    }

    clearSession();
    goto("/login");
  };
</script>

{#snippet fieldRow(fields: PolicyField[])}
  <div class="flex flex-col sm:flex-row gap-5 sm:gap-6">
    {#each fields as field (field.key)}
      <NumberField
        bind:value={inputs[field.key]}
        id={field.key}
        label={field.label}
        min={field.min}
        max={field.max}
      />
    {/each}
  </div>
{/snippet}

<SettingsForm
  title="Password Settings"
  description="Set the rules every new password must meet"
  {dirty}
  {issue}
  onsave={save}
  ondiscard={discard}
>
  <Card
    title="Password Policy"
    icon="password"
    description="Existing passwords are never revalidated, raising these only applies the next time a password is set."
  >
    <div class="mt-6 flex flex-col gap-5">
      {@render fieldRow(lengthFields)}
      {@render fieldRow(characterFields)}
    </div>

    {#if issue}
      <p class="mt-4 flex items-start gap-2 text-sm text-red-300">
        <Icon icon="alert-triangle" class="mt-0.5" />
        {issue}
      </p>
    {/if}
  </Card>

  <Card
    title="Password Hashing"
    icon="key"
    description="New passwords are hashed with this algorithm. Existing passwords are rehashed the next time their owner signs in."
  >
    <div class="mt-6 flex flex-col gap-3">
      <Segmented
        bind:value={algorithm}
        options={algorithmOptions}
        label="Hashing algorithm"
      />
      <p class="text-sm text-muted">
        {PASSWORD_ALGORITHMS[algorithm].description}
      </p>
    </div>

    <div class="mt-6 flex flex-col gap-3">
      {#if status}
        <div class="flex flex-wrap items-baseline justify-between gap-2">
          <p class="text-sm">
            {#if migrated >= total}
              All passwords for every user use {PASSWORD_ALGORITHMS[
                status.algorithm
              ].label}
            {:else}
              {migrated.toLocaleString()} of {total.toLocaleString()}
              passwords use {PASSWORD_ALGORITHMS[status.algorithm].label}
            {/if}
          </p>
          <span class="font-mono text-sm text-muted select-none">
            {percentOf(migrated)}%
          </span>
        </div>

        <div
          role="progressbar"
          aria-label="Password migration progress"
          aria-valuemin={0}
          aria-valuemax={100}
          aria-valuenow={percentOf(migrated)}
          class="h-1.5 w-full overflow-hidden rounded-full bg-background-200"
        >
          <div
            class="h-full rounded-full bg-primary transition-[width]"
            style:width="{percentOf(migrated)}%"
          ></div>
        </div>

        {#if status.counts.length > 0}
          <dl
            class="grid gap-px overflow-hidden rounded-lg border border-background-200 bg-background-200 sm:grid-cols-2"
          >
            {#each status.algorithms as entry (entry)}
              {@const count =
                status.counts.find((it) => it.algorithm === entry)?.count ?? 0}
              <div class="flex flex-col gap-1 bg-background-100 p-4">
                <dt
                  class="flex items-center gap-2 text-sm text-muted select-none"
                >
                  {PASSWORD_ALGORITHMS[entry].label}
                  {#if entry === status.algorithm}
                    <Icon icon="check" class="text-primary" />
                  {/if}
                </dt>
                <dd class="text-lg font-semibold">
                  {count.toLocaleString()}
                  <span class="text-sm font-normal text-muted">
                    {count === 1 ? "password" : "passwords"}
                  </span>
                </dd>
              </div>
            {/each}
          </dl>
        {/if}
      {:else if statusFailed}
        <p class="flex items-center gap-2 text-sm text-red-300">
          <Icon icon="alert-triangle" />
          Failed to load the migration status
        </p>
      {:else}
        <p class="text-sm text-muted">Loading migration status</p>
      {/if}
    </div>
  </Card>

  {#if canRequireForAll}
    <Card
      title="Require Password Change"
      icon="lock"
      description="Use this if passwords may have been exposed. Everyone is signed out and every user with a password must choose a new one at their next sign in."
    >
      <div class="mt-4">
        <Button
          type="button"
          variant="danger"
          onclick={() => requireForAllModal?.open()}
        >
          Require for all users
        </Button>
      </div>
    </Card>
  {/if}
</SettingsForm>

<ConfirmModal
  bind:this={requireForAllModal}
  title="Require password change for all users"
  confirmLabel="Require password change"
  onconfirm={requireForAll}
>
  Every session and access token on this server is revoked immediately,
  including yours and those issued to connected applications. Every user with a
  password must choose a new one at their next sign in. This can not be undone.
</ConfirmModal>
