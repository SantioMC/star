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
	<div class="flex m-4 gap-6 items-center xl:items-stretch justify-evenly flex-col xl:flex-row">
		<Card.Root class='max-w-lg min-w-72 w-full text-blue-200'>
			<Card.CardHeader>
				<Card.Title><Icon icon='users' /> Total Users</Card.Title>
				<Card.Description>The total amount of users who have an account</Card.Description>
			</Card.CardHeader>
			<Card.Content class="flex flex-row items-end gap-2 overflow-hidden flex-wrap">
				<span class='text-5xl'>{(metrics?.total_users ?? 0).toLocaleString()}</span>
				<span class='text-lg text-white'>user{(metrics?.total_users ?? 0) == 1 ? '' : 's'}</span>
			</Card.Content>
		</Card.Root>
		<Card.Root class='max-w-lg min-w-72 w-full text-emerald-200'>
			<Card.CardHeader>
				<Card.Title><Icon icon='device-desktop' /> Total Sessions</Card.Title>
				<Card.Description>The total amount of currently logged in sessions/devices</Card.Description>
			</Card.CardHeader>
			<Card.Content class="flex flex-row items-end gap-2 overflow-hidden flex-wrap">
				<span class='text-5xl'>{(metrics?.total_sessions ?? 0).toLocaleString()}</span>
				<span class='text-lg text-white'>session{(metrics?.total_sessions ?? 0) == 1 ? '' : 's'}</span>
			</Card.Content>
		</Card.Root>
		<Card.Root class='max-w-lg min-w-72 w-full text-fuchsia-200'>
			<Card.CardHeader>
				<Card.Title><Icon icon='app-window' /> Star Metadata</Card.Title>
				<Card.Description>Metadata relating to the Star server currently requesting data from</Card.Description>
			</Card.CardHeader>
			<Card.Content class="flex flex-col">
				<span class='text-white text-sm'>Current version: <span class='text-lg text-fuchsia-200'>{metrics?.version ?? "unknown"}</span></span>
				<span class='text-white text-sm'>Connected Node: <span class='text-lg text-fuchsia-200'>{metrics?.node_id ?? "unknown"}</span></span>
				<span class='text-white text-sm'>Node online for: <span class='text-lg text-fuchsia-200'>{getUptime()}</span></span>
			</Card.Content>
		</Card.Root>
	</div>
</Sidebar>