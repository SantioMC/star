<script lang="ts">
	import Icon from "@/components/icon.svelte";
	import { hasPermission } from "@/index";
	import type { StarPermission } from "@/star";
	import type { Snippet } from "svelte";
	import type { ClassValue } from "svelte/elements";

  const Category = [
    {
      name: 'My Account',
      id: 'personal',
    },
    {
      name: 'Management',
      id: 'management',
      permission: 'star:admin',
    },
  ] as const satisfies readonly {
    name: string
    id: string
    permission?: StarPermission | string
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
      name: 'Account Details',
      icon: 'user',
      href: '/',
      category: 'personal'
    },
    // {
    //   name: 'Activity Log',
    //   icon: 'logs',
    //   href: '/',
    //   category: 'personal',
    //   permission: 'star:user:activity'
    // },
    {
      name: 'Overview',
      icon: 'menu-3',
      href: '/overview',
      category: 'management',
      permission: 'star:server:metrics'
    },
    {
      name: 'User Management',
      icon: 'users',
      href: '/admin/users',
      category: 'management',
      permission: 'star:admin:read:users'
    },
    {
      name: 'Roles',
      icon: 'hammer',
      href: '/admin/roles',
      category: 'management',
      permission: 'star:admin:read:roles'
    },
    {
      name: 'Applications',
      icon: 'puzzle',
      href: '/admin/apps',
      category: 'management',
      permission: 'star:admin:read:applications'
    }
  ]

  const props: { 
    children?: Snippet,
    hidden?: boolean,
    class?: ClassValue,
  } = $props()
</script>

<div class="flex flex-row gap-1 min-h-dvh">
  {#if !props.hidden}
    <div class='bg-secondary w-56 min-w-56 shadow flex flex-col px-2 transition-all'>
      {#each Category as category (category.id)}
        {#if !('permission' in category) || hasPermission(category.permission)}
          <span class='text-secondary-foreground/50 text-sm px-2 pt-4'>{category.name}</span>

          {#each items.filter(i => i.category == category.id) as item (item.name)}
            {#if !item.permission || hasPermission(item.permission)}
              <a href={item.href} class='transition-colors hover:bg-neutral-200/10 py-1 px-2 mx-1 rounded-lg'>
                {#if item.icon}
                  <Icon icon={item.icon} />
                {/if}

                {item.name}
              </a>
            {/if}
          {/each}
        {/if}
      {/each}
    </div>
  {/if}

  <div class='w-full h-full {props.class ?? ''}'>
    {@render props.children?.()}
  </div>
</div>