<script lang="ts">
  import { afterNavigate } from "$app/navigation";
  import BrandMark from "@/components/brand-mark.svelte";
  import Icon from "@/components/icon.svelte";
  import type { Snippet } from "svelte";
  import type { ClassValue } from "svelte/elements";
  import NavGroup from "./nav-group.svelte";
  import NavItem from "./nav-item.svelte";
  import Category from "./category.svelte";
  import { hasPermission, user } from "@/index";

  let props: {
    children?: Snippet;
    hidden?: boolean;
    class?: ClassValue;
  } = $props();

  let isAdmin = $derived(
    $user?.permissions?.some((u) => u.startsWith("admin:")),
  );

  let mobileOpen = $state(false);

  afterNavigate(() => {
    mobileOpen = false;
  });
</script>

<svelte:window
  onkeydown={(event) => {
    if (event.key === "Escape") mobileOpen = false;
  }}
/>

{#if !props.hidden}
  <div
    class="md:hidden sticky top-0 z-30 flex h-12 items-center gap-2 border-b border-background-200 bg-background-100 px-2"
  >
    <button
      type="button"
      aria-label="Open navigation"
      aria-expanded={mobileOpen}
      class="rounded-lg p-2 transition-colors hover:bg-background-200"
      onclick={() => (mobileOpen = true)}
    >
      <Icon icon="menu-2" class="text-lg" />
    </button>
    <BrandMark class="min-w-0" />
  </div>

  {#if mobileOpen}
    <button
      type="button"
      aria-label="Close navigation"
      class="md:hidden fixed inset-0 z-40 bg-black/60 animate-in fade-in-0"
      onclick={() => (mobileOpen = false)}
    ></button>
  {/if}
{/if}

<div class="flex flex-row min-h-dvh">
  {#if !props.hidden}
    <nav
      data-open={mobileOpen}
      class="w-60 min-w-60 h-dvh sticky top-0 border-r border-background-200 flex flex-col py-4 px-3 gap-1 bg-background-100 overflow-y-auto max-md:fixed max-md:left-0 max-md:z-50 max-md:-translate-x-full max-md:transition-transform max-md:data-[open=true]:translate-x-0"
    >
      <BrandMark class="mb-4 px-1 text-foreground" />

      <div class="flex flex-col gap-4">
        <div class="flex flex-col gap-0.5">
          <Category>Your Account</Category>
          <NavItem href="/" icon="user">Profile</NavItem>
          <NavItem href="/security" icon="shield">Security</NavItem>
          <NavItem href="/applications" icon="apps">Applications</NavItem>

          {#if isAdmin}
            <Category>Directory</Category>
            <NavItem
              permission="admin:user:read"
              href="/admin/users"
              icon="users"
            >
              Users
            </NavItem>
            <NavItem
              permission="admin:role:read"
              href="/admin/roles"
              icon="user-shield"
            >
              Roles
            </NavItem>

            <Category>Platform</Category>
            <NavItem
              permission="admin:metrics:read"
              href="/admin/overview"
              icon="chart-line"
            >
              Overview
            </NavItem>
            <NavGroup
              label="Monitoring"
              icon="activity"
              items={[
                {
                  href: "/admin/alerts",
                  label: "Security Alerts",
                  permission: "admin:alert:read",
                },
                {
                  href: "/admin/audit",
                  label: "Audit Log",
                  permission: "admin:audit:read",
                },
              ]}
            />
            <NavGroup
              label="Integrations"
              icon="plug-connected"
              items={[
                {
                  href: "/admin/apps",
                  label: "Applications",
                  permission: "application:read",
                },
                {
                  href: "/admin/providers",
                  label: "Social Providers",
                  permission: "admin:provider:read",
                },
              ]}
            />
            <NavGroup
              label="Keys & Tokens"
              icon="key"
              items={[
                {
                  href: "/admin/keys",
                  label: "Signing Keys",
                  permission: "admin:key:read",
                },
                {
                  href: "/admin/revocations",
                  label: "Token Revocations",
                  permission: "admin:token:read",
                },
              ]}
            />
            {#if hasPermission("admin:settings:read")}
              <NavGroup
                label="Settings"
                icon="settings"
                items={[
                  { href: "/admin/settings/access", label: "Access" },
                  { href: "/admin/settings/email", label: "Email" },
                  { href: "/admin/settings/passwords", label: "Passwords" },
                  { href: "/admin/settings/branding", label: "Branding" },
                  {
                    href: "/admin/settings/audit",
                    label: "Audit Log",
                    permission: "admin:audit:read",
                  },
                ]}
              />
            {/if}
          {/if}
        </div>
      </div>

      <div class="mt-auto px-3 py-2">
        <NavItem href="/logout" class="rounded-xl p-3 hover:bg-red-400/5">
          <Icon icon="logout" class="text-base" />
          Sign out
        </NavItem>
      </div>
    </nav>
  {/if}

  <main class="flex-1 min-h-dvh overflow-y-auto {props.class ?? ''}">
    {@render props.children?.()}
  </main>
</div>
