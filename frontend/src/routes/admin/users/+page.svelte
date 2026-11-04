<script lang="ts">
    import { branding } from "@/branding";
    import { goto } from "$app/navigation";
    import { resolve } from "$app/paths";
    import Badge from "@/components/badge.svelte";
    import Card from "@/components/card.svelte";
    import Pagination from "@/components/pagination.svelte";
    import PageHeader from "@/components/page-header.svelte";
    import SearchInput from "@/components/search-input.svelte";
    import * as Avatar from "@/components/ui/avatar";
    import { hasPermission } from "@/index";
    import * as star from "@/star";
    import { simpleDate, titleCase } from "@/utils";
    import { onMount } from "svelte";

    const PAGE_SIZE = 25;

    let users = $state<star.AdminUser[]>([]);
    let loaded = $state(false);
    let query = $state("");
    let page = $state(1);
    let searchTask: ReturnType<typeof setTimeout> | null = null;

    const hasNextPage = $derived(users.length >= PAGE_SIZE);

    const load = async () => {
        loaded = false;
        users = await star.getAllUsers(page, query.trim());
        loaded = true;
    };

    onMount(() => {
        if (!hasPermission("admin:user:read")) return goto("/");
        load();
    });

    const handleSearch = () => {
        if (searchTask) clearTimeout(searchTask);

        searchTask = setTimeout(() => {
            page = 1;
            load();
        }, 250);
    };

    const changePage = (delta: number) => {
        page = Math.max(1, page + delta);
        load();
    };

    const displayName = (user: star.AdminUser): string => {
        const fullName = [user.first_name, user.last_name]
            .filter(Boolean)
            .join(" ");

        return fullName || user.username;
    };
</script>

<PageHeader
    title="User Directory"
    description="View and manage every user registered with {$branding.name}"
/>

<SearchInput
    bind:value={query}
    oninput={handleSearch}
    placeholder="Search by username or email"
/>

<Card class="p-2 sm:p-3">
    <table class="w-full text-sm">
        <thead>
            <tr class="hover:bg-transparent">
                <th class="px-3 text-xs font-medium text-muted select-none">
                    User
                </th>
                <th
                    class="px-3 text-xs font-medium text-muted select-none max-sm:hidden"
                >
                    Roles
                </th>
                <th
                    class="px-3 text-xs font-medium text-muted select-none text-right max-sm:hidden"
                >
                    Joined
                </th>
            </tr>
        </thead>
        <tbody>
            {#each users as user (user.id)}
                <tr
                    class="cursor-pointer hover:bg-hover"
                    onclick={() =>
                        goto(resolve("/admin/users/[user]", { user: user.id }))}
                >
                    <td class="px-3 py-2.5">
                        <div class="flex items-center gap-3">
                            <Avatar.Root class="size-8">
                                <Avatar.Image
                                    src={user.profile_picture}
                                    alt="{user.username}'s profile picture"
                                />
                                <Avatar.Fallback>
                                    {user.username.charAt(0).toUpperCase()}
                                </Avatar.Fallback>
                            </Avatar.Root>
                            <div class="min-w-0">
                                <a
                                    href={resolve("/admin/users/[user]", {
                                        user: user.id,
                                    })}
                                    class="block truncate font-medium"
                                    onclick={(event) => event.stopPropagation()}
                                >
                                    {displayName(user)}
                                </a>
                                <p class="truncate text-xs text-muted">
                                    <span class="font-mono">
                                        {user.username}
                                    </span>
                                    {#if user.email}
                                        &middot; {user.email}
                                    {/if}
                                </p>
                            </div>
                            {#if !user.email_verified}
                                <Badge tone="warning" class="ml-1">
                                    unverified
                                </Badge>
                            {/if}
                            {#if user.lockout?.locked}
                                <Badge tone="danger" class="ml-1">locked</Badge>
                            {/if}
                        </div>
                    </td>
                    <td class="px-3 max-sm:hidden">
                        <div class="flex flex-wrap gap-1">
                            {#each user.roles as role (role.id)}
                                <Badge>{titleCase(role.name)}</Badge>
                            {:else}
                                <span class="text-xs text-muted">None</span>
                            {/each}
                        </div>
                    </td>
                    <td
                        class="px-3 text-right text-muted whitespace-nowrap max-sm:hidden"
                    >
                        {simpleDate(new Date(user.created_at))}
                    </td>
                </tr>
            {:else}
                <tr class="hover:bg-transparent">
                    <td colspan="3" class="py-10 text-center text-muted">
                        {#if !loaded}
                            Loading users...
                        {:else if query}
                            No users match "{query}".
                        {:else}
                            No users found.
                        {/if}
                    </td>
                </tr>
            {/each}
        </tbody>
    </table>
</Card>

<Pagination
    {page}
    hasNext={hasNextPage}
    {loaded}
    onprevious={() => changePage(-1)}
    onnext={() => changePage(1)}
/>
