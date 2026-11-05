<script lang="ts">
    import * as star from "$lib/star";
    import Badge from "@/components/badge.svelte";
    import Button from "@/components/button.svelte";
    import Icon from "@/components/icon.svelte";
    import Modal from "@/components/modal.svelte";
    import SearchInput from "@/components/search-input.svelte";
    import { cn } from "@/utils";
    import { toast } from "svelte-sonner";

    const { user }: { user: star.AdminUser } = $props();

    let modal = $state<Modal>();
    let applications = $state<star.SharedApplication[]>([]);
    let selected = $state<star.SharedApplication | null>(null);
    let query = $state("");
    let page = $state(1);
    let hasMore = $state(false);
    let loaded = $state(false);
    let loadingMore = $state(false);
    let starting = $state(false);

    let searchTask: ReturnType<typeof setTimeout> | null = null;
    let latestRequest = 0;

    const fetchPage = async (target: number) => {
        const request = ++latestRequest;
        const result = await star.getImpersonationApplications(
            user.id,
            target,
            query.trim(),
            25,
        );

        if (request !== latestRequest) return null;
        if ("error" in result) {
            toast.error(result.message);
            return [];
        }

        page = target;
        hasMore = result.length >= 25;

        return result;
    };

    const search = async () => {
        loaded = false;

        const result = await fetchPage(1);
        if (!result) return;

        applications = result;
        loaded = true;
    };

    const loadMore = async () => {
        if (!hasMore || loadingMore || !loaded) return;

        loadingMore = true;
        const result = await fetchPage(page + 1);
        loadingMore = false;

        if (result) applications = [...applications, ...result];
    };

    const handleSearch = () => {
        if (searchTask) clearTimeout(searchTask);
        searchTask = setTimeout(search, 250);
    };

    const handleScroll = (event: Event) => {
        const list = event.currentTarget as HTMLElement;
        const remaining =
            list.scrollHeight - list.scrollTop - list.clientHeight;

        if (remaining < 80) loadMore();
    };

    const host = (application: star.SharedApplication) => {
        if (!URL.canParse(application.url)) return application.url;
        return new URL(application.url).host;
    };

    export const open = () => {
        query = "";
        selected = null;
        applications = [];
        modal?.open();
        search();
    };

    const start = async (event: Event) => {
        event.preventDefault();
        if (!selected || starting) return;

        starting = true;

        const tab = window.open("about:blank", "_blank");
        const result = await star.impersonateUser(user.id, selected.id);

        starting = false;

        if ("error" in result) {
            tab?.close();
            return toast.error(result.message);
        }

        if (!tab) {
            window.location.href = result.url;
            return;
        }

        tab.opener = null;
        tab.location.href = result.url;

        modal?.close();
        toast.success(`Opening ${selected.name} as ${user.username}`);
    };
</script>

<Modal bind:this={modal} button>
    <form class="flex flex-col gap-5" onsubmit={start}>
        <div class="flex flex-col gap-1 pr-8">
            <h1 class="text-xl font-semibold">Impersonate {user.username}</h1>
            <h2 class="text-sm text-muted">
                Choose an application to sign in to as this user.
            </h2>
        </div>

        <SearchInput
            bind:value={query}
            oninput={handleSearch}
            placeholder="Search applications"
            aria-label="Search applications"
            class="sm:max-w-none"
        />

        <div
            class="h-72 overflow-y-auto overscroll-contain rounded-lg border border-background-200"
            onscroll={handleScroll}
        >
            {#if !loaded}
                <div class="flex h-full items-center justify-center">
                    <Icon
                        icon="loader-2"
                        class="animate-spin text-2xl text-muted"
                    />
                </div>
            {:else if applications.length === 0}
                <div
                    class="flex h-full flex-col items-center justify-center gap-1 px-6 text-center select-none"
                >
                    <Icon icon="apps" class="text-2xl text-muted" />
                    <p class="text-sm text-muted">
                        {query.trim()
                            ? `No applications match "${query.trim()}".`
                            : "You and this user do not share any applications."}
                    </p>
                </div>
            {:else}
                <div class="flex flex-col gap-1 p-1.5">
                    {#each applications as application (application.id)}
                        {@const active = selected?.id === application.id}
                        <button
                            type="button"
                            aria-pressed={active}
                            class={cn(
                                "flex w-full items-center gap-3 rounded-md border px-2.5 py-2 text-left transition-colors hover:bg-hover",
                                active
                                    ? "border-primary/40 bg-hover"
                                    : "border-transparent",
                            )}
                            onclick={() => (selected = application)}
                        >
                            <span
                                class="flex size-9 shrink-0 items-center justify-center rounded-md bg-background-200 text-sm font-medium uppercase select-none"
                                aria-hidden="true"
                            >
                                {application.name.charAt(0)}
                            </span>
                            <span class="flex min-w-0 flex-1 flex-col">
                                <span class="truncate text-sm">
                                    {application.name}
                                </span>
                                <span
                                    class="truncate font-mono text-xs text-muted"
                                >
                                    {host(application)}
                                </span>
                            </span>
                            {#if application.authorized}
                                <Badge
                                    tone="success"
                                    title="{user.username} has authorized this application"
                                >
                                    authorized
                                </Badge>
                            {/if}
                            {#if active}
                                <Icon icon="check" class="text-primary" />
                            {/if}
                        </button>
                    {/each}
                    {#if loadingMore}
                        <div class="flex justify-center py-2">
                            <Icon
                                icon="loader-2"
                                class="animate-spin text-muted"
                            />
                        </div>
                    {/if}
                </div>
            {/if}
        </div>

        <p class="text-xs text-muted">
            Only applications that you have access to will show up. The
            application opens in a new tab, and lasts at most an hour. This
            action will be recorded.
        </p>

        <div class="flex items-center gap-2 justify-end">
            <Button
                type="button"
                variant="secondary"
                onclick={() => modal?.close()}
            >
                Cancel
            </Button>
            <Button type="submit" loading={starting} disabled={!selected}>
                Impersonate
            </Button>
        </div>
    </form>
</Modal>
