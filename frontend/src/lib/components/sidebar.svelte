<script lang="ts">
	import { page } from "$app/state";
	import Icon from "@/components/icon.svelte";
	import { hasPermission } from "@/index";
	import type { StarPermission } from "@/star";
	import type { Snippet } from "svelte";
	import type { ClassValue } from "svelte/elements";

  const Category = [
    {
      name: 'Account',
      id: 'account',
    },
    {
      name: 'Server',
      id: 'server',
    },
    {
      name: 'Administration',
      id: 'admin',
    },
  ] as const satisfies readonly {
    name: string
    id: string
  }[]

  type Category = typeof Category[number]['id']

  const items: {
    name: string
    icon?: string
    href: string
    category: Category
    permission?: StarPermission | (string & {})
  }[] = [
    {
      name: 'Profile',
      icon: 'user',
      href: '/',
      category: 'account'
    },
    {
      name: 'Overview',
      icon: 'chart-bar',
      href: '/overview',
      category: 'server',
      permission: 'star:server:metrics'
    },
    {
      name: 'Users',
      icon: 'users',
      href: '/admin/users',
      category: 'admin',
      permission: 'admin:user:read'
    },
    {
      name: 'Roles',
      icon: 'shield',
      href: '/admin/roles',
      category: 'admin',
      permission: 'admin:role:read'
    },
    {
      name: 'Applications',
      icon: 'apps',
      href: '/admin/apps',
      category: 'admin',
      permission: 'application:read'
    },
    {
      name: 'Providers',
      icon: 'plug-connected',
      href: '/admin/providers',
      category: 'admin',
      permission: 'admin:provider:read'
    },
    {
      name: 'Settings',
      icon: 'settings',
      href: '/admin/settings',
      category: 'admin',
      permission: 'admin:settings:read'
    }
  ]

  const visibleItems = (category: Category) =>
    items.filter(i => i.category == category && (!i.permission || hasPermission(i.permission)))

  const isActive = (href: string): boolean => {
    if (href.includes('#')) {
      return page.url.pathname === href.split('#')[0];
    }

    if (href === '/') return page.url.pathname === '/';
    return page.url.pathname.startsWith(href);
  }

  let props: {
    children?: Snippet,
    hidden?: boolean,
    class?: ClassValue,
  } = $props()
</script>

<div class="flex flex-row min-h-dvh">
  {#if !props.hidden}
    <nav class='w-60 min-w-60 h-dvh sticky top-0 border-r border-border/50 flex flex-col py-4 px-3 gap-1 bg-sidebar overflow-y-auto'>
      <div class="flex items-center gap-2 px-3 mb-4">
        <div class="size-7 rounded-md bg-primary/10 flex items-center justify-center">
          <Icon icon="star-filled" class="text-sm text-primary/80" />
        </div>
        <span class="font-semibold text-sm text-primary/90 tracking-tight">Star</span>
      </div>

      {#each Category as category (category.id)}
        {#if visibleItems(category.id).length > 0}
          <span class='text-muted-foreground/60 text-[11px] font-medium uppercase tracking-wider px-3 pt-5 pb-1'>
            {category.name}
          </span>

          {#each visibleItems(category.id) as item (item.name)}
            <a
              href={item.href}
              class="flex items-center gap-2.5 px-3 py-1.5 rounded-md text-sm transition-colors
                {isActive(item.href)
                  ? 'bg-accent text-accent-foreground font-medium'
                  : 'text-muted-foreground hover:bg-accent/50 hover:text-accent-foreground'}"
            >
              {#if item.icon}
                <Icon icon={item.icon} class="text-base opacity-70" />
              {/if}
              {item.name}
            </a>
          {/each}
        {/if}
      {/each}

      <div class="mt-auto px-3 py-2">
        <a
          href="/logout"
          class="flex items-center gap-2.5 px-3 py-1.5 rounded-md text-sm text-muted-foreground/60 hover:text-destructive hover:bg-destructive/10 transition-colors"
        >
          <Icon icon="logout" class="text-base" />
          Sign out
        </a>
      </div>
    </nav>
  {/if}

  <main class='flex-1 min-h-dvh overflow-y-auto {props.class ?? ''}'>
    {@render props.children?.()}
  </main>
</div>
