<script lang="ts">
	import { browser } from "$app/environment";
	import * as star from '$lib/star';
	import Icon from "@/components/icon.svelte";
	import Sidebar from "@/components/sidebar.svelte";
	import * as Card from '@/components/ui/card';
	import { hasPermission } from "@/index";
	import { relativeTime } from "@/utils";
	import { onMount } from "svelte";

	let metrics = $state<star.Metrics|null>(null)

	onMount(async () => {
		if (!browser) return
		if (!hasPermission('star:server:metrics')) return location.href = '/'

		metrics = await star.getMetrics()
	})

	const getUptime = () => {
		if (!metrics?.started_at) return "unknown"
		return relativeTime(metrics.started_at)
	}
</script>

<Sidebar>
	<div class="flex flex-col p-6 gap-6 max-w-6xl mx-auto">
		<div>
			<h1 class="text-2xl font-semibold">Overview</h1>
			<p class="text-sm text-muted-foreground mt-1">Server metrics and status at a glance</p>
		</div>

		<div class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
			<Card.Root class='border-blue-500/20'>
				<Card.CardHeader class="pb-2">
					<Card.Description class="flex items-center gap-1.5 text-blue-300/80">
						<Icon icon='users' class="text-sm" />
						Total Users
					</Card.Description>
				</Card.CardHeader>
				<Card.Content>
					<div class="flex items-baseline gap-2">
						<span class='text-4xl font-semibold text-blue-200'>{(metrics?.total_users ?? 0).toLocaleString()}</span>
						<span class='text-sm text-muted-foreground'>user{(metrics?.total_users ?? 0) == 1 ? '' : 's'}</span>
					</div>
				</Card.Content>
			</Card.Root>

			<Card.Root class='border-emerald-500/20'>
				<Card.CardHeader class="pb-2">
					<Card.Description class="flex items-center gap-1.5 text-emerald-300/80">
						<Icon icon='device-desktop' class="text-sm" />
						Active Sessions
					</Card.Description>
				</Card.CardHeader>
				<Card.Content>
					<div class="flex items-baseline gap-2">
						<span class='text-4xl font-semibold text-emerald-200'>{(metrics?.total_sessions ?? 0).toLocaleString()}</span>
						<span class='text-sm text-muted-foreground'>session{(metrics?.total_sessions ?? 0) == 1 ? '' : 's'}</span>
					</div>
				</Card.Content>
			</Card.Root>

			<Card.Root class='border-violet-500/20'>
				<Card.CardHeader class="pb-2">
					<Card.Description class="flex items-center gap-1.5 text-violet-300/80">
						<Icon icon='server' class="text-sm" />
						Server Info
					</Card.Description>
				</Card.CardHeader>
				<Card.Content>
					<div class="flex flex-col gap-1.5 text-sm">
						<div class="flex justify-between">
							<span class="text-muted-foreground">Version</span>
							<span class="text-violet-200 font-medium">{metrics?.version ?? "—"}</span>
						</div>
						<div class="flex justify-between">
							<span class="text-muted-foreground">Node</span>
							<span class="text-violet-200 font-medium font-mono text-xs">{metrics?.node_id ?? "—"}</span>
						</div>
						<div class="flex justify-between">
							<span class="text-muted-foreground">Uptime</span>
							<span class="text-violet-200 font-medium">{getUptime()}</span>
						</div>
					</div>
				</Card.Content>
			</Card.Root>
		</div>
	</div>
</Sidebar>
