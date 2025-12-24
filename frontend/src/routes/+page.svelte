<script lang="ts">
	import { browser } from "$app/environment";
	import { goto } from "$app/navigation";
	import Icon from '$lib/components/icon.svelte';
	import Sidebar from "@/components/sidebar.svelte";
	import * as Avatar from "@/components/ui/avatar";
	import { Button } from "@/components/ui/button";
	import * as Card from "@/components/ui/card";
	import { Input } from "@/components/ui/input";
	import * as Item from "@/components/ui/item";
	import { Label } from "@/components/ui/label";
	import { hasPermission } from "@/index";
	import * as star from "@/star";
	import { getIcon, relativeTime } from "@/utils";
	import { onMount } from "svelte";

	let data = $state<star.User|null>(null)
	$inspect(data)

	onMount(async () => {
		if (!browser) return
		data = await star.getSelf()

		if (data == null) {
			goto('/login')
		}
	})
</script>

<Sidebar hidden={data == null || !hasPermission('star:server') && !hasPermission('star:admin')}>
	{#if data != null}
		<div class="flex flex-col md:w-3xl mx-auto py-24 gap-8">
			<div class="flex items-center">
				<div>
					<h1 class="leading-8 text-2xl font-semibold">Manage your account</h1>
					<h2 class="leading-4 text-primary/70">Manage your account settings and preferences</h2>
				</div>

				{#if hasPermission('star:admin:read:users')}
					<Button onclick={() => goto(`/admin/users/${data!.id}`)} variant='secondary' class='ml-auto'>View Admin Page</Button>
				{/if}
			</div>
			
			<Card.Root class='min-w-1/3 w-full md:w-auto'>
				<Card.CardHeader>
					<Card.Title><Icon icon='user' /> Profile Information</Card.Title>
				</Card.CardHeader>
				<Card.Content class="flex flex-col gap-8">
					<div class="flex gap-2 items-center">
						<Avatar.Root>
							<Avatar.Image src={data.profile_picture} alt="{data.username}'s Profile Picture" />
						</Avatar.Root>
						<div class="flex flex-col gap-1 items-baseline">
							<span class="leading-4 text-lg">
								{#if data.first_name && data.last_name}
									{data.first_name} {data.last_name}
								{:else}
									{data.username}
								{/if}
							</span>
							<span class="leading-4 text-sm text-primary/70">
								Not you? <a class="text-blue-200/70" href="/logout">Log out</a>
							</span>
						</div>
					</div>

					<form class="flex flex-col gap-6">
						<div class="grid gap-2">
							<Label for="username">Username</Label>
							<Input id="username" type="text" required value={data.username} />
						</div>

						<div class="grid gap-2">
							<Label for="email">Email</Label>
							<Input id="email" type="email" value={data.email} disabled />
						</div>

						<div class="flex gap-2 sm:gap-8">
							<div class="flex-1/2 gap-2 grid">
								<Label for="givenName">Given Name</Label>
								<Input id="givenName" type="text" value={data.first_name} />
							</div>

							<div class="flex-1/2 gap-2 grid">
								<Label for="familyName">Family Name</Label>
								<Input id="familyName" type="text" value={data.last_name} />
							</div>
						</div>
					</form>
				</Card.Content>
				<Card.Footer>
					<Button class='ml-auto'>Update</Button>
				</Card.Footer>
			</Card.Root>

			<Card.Root class='min-w-1/3 w-full md:w-auto'>
				<Card.CardHeader>
					<Card.Title><Icon icon='lock' /> Change Password</Card.Title>
				</Card.CardHeader>
				<Card.Content class="flex flex-col gap-8">
					<form class="flex flex-col gap-6">
						<div class="grid gap-2">
							<Label for="currentPassword">Current Password</Label>
							<Input id="currentPassword" type="password" required placeholder={"••••••••••••••••"} />
						</div>

						<div class="grid gap-2">
							<Label for="password">New Password</Label>
							<Input id="password" type="password" />
						</div>

						<div class="grid gap-2">
							<Label for="confirmPassword">Confirm Password</Label>
							<Input id="confirmPassword" type="password" />
						</div>
					</form>
				</Card.Content>
				<Card.Footer>
					<Button class='ml-auto'>Change Password</Button>
				</Card.Footer>
			</Card.Root>

			<Card.Root class='min-w-1/3 w-full md:w-auto'>
				<Card.CardHeader>
					<Card.Title><Icon icon='device-desktop' /> Sessions</Card.Title>
					<Card.Description class='max-w-xl'>The different devices and sessions currently logged into your account, invalidating a session might take up to an hour to take effect.</Card.Description>
				</Card.CardHeader>
				<Card.Content class="flex flex-col gap-2">
					{#each data.sessions.filter(i => !i.expired).sort((a, b) => b.created_at - a.created_at) as session}
						<Item.Root variant='outline'>
							<Item.Content class='flex flex-row items-center gap-4'>
								<Icon class='text-3xl' icon={getIcon(session)} />
								<div>
									<p class="text-lg">{session.device} {session.operating_system}</p>
									<Item.Description class='text-xs'>IP Address: {session.ip_address} • Created {relativeTime(session.created_at)} ago</Item.Description>
								</div>
								<div class='ml-auto'>
									<Button variant='destructive'>Log out</Button>
								</div>
							</Item.Content>
						</Item.Root>
					{/each}
				</Card.Content>
				<Card.Footer>
					<Button class='ml-auto' variant='destructive'>Sign out all sessions</Button>
				</Card.Footer>
			</Card.Root>

			<Card.Root class='min-w-1/3 w-full md:w-auto' >
				<Card.CardHeader>
					<Card.Title class="text-red-300"><Icon icon='trash' /> Delete Account</Card.Title>
				</Card.CardHeader>
				<Card.Content class="flex flex-col gap-8 text-red-200/60">
					<span>
						Delete your account and all information about you, this action can not be reversed! We might preserve some required information needed to prevent abuse on our systems.
					</span>
				</Card.Content>
				<Card.Footer>
					<Button class='ml-auto' variant='destructive'>Delete Account</Button>
				</Card.Footer>
			</Card.Root>
		</div>
	{/if}
</Sidebar>