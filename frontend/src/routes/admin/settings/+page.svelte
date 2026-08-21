<script lang="ts">
	import { browser } from '$app/environment';
	import { goto } from '$app/navigation';
	import * as star from '$lib/star';
	import Icon from '@/components/icon.svelte';
	import Sidebar from '@/components/sidebar.svelte';
	import { Button } from '@/components/ui/button';
	import * as Card from '@/components/ui/card';
	import { Checkbox } from '@/components/ui/checkbox';
	import { Input } from '@/components/ui/input';
	import { Label } from '@/components/ui/label';
	import { Spinner } from '@/components/ui/spinner';
	import { hasPermission } from '@/index';
	import { onMount } from 'svelte';
	import { toast } from 'svelte-sonner';

	let settings = $state<star.ServerSettings | null>(null);
	let loading = $state<boolean>(false);

	let allowRegistration = $state<boolean>(true);
	let allowLogins = $state<boolean>(true);
	let emailCodeLength = $state<number>(6);
	let minLength = $state<number>(8);
	let maxLength = $state<number>(128);
	let minSymbols = $state<number>(1);
	let minUpper = $state<number>(1);
	let minLower = $state<number>(1);

	const canModify = $derived(hasPermission('admin:settings:write'));
	const requiredCharacters = $derived(minSymbols + minUpper + minLower);

	const policyIssue = $derived.by(() => {
		if (maxLength < minLength) {
			return 'The maximum length must be greater than the minimum length';
		}

		if (requiredCharacters > minLength) {
			return 'The required characters add up to more than the minimum length, no password could satisfy this policy';
		}

		return null;
	});

	const apply = (value: star.ServerSettings) => {
		settings = value;
		allowRegistration = value.allow_registration;
		allowLogins = value.allow_logins;
		emailCodeLength = value.email_code_length;
		minLength = value.password_settings.min_length;
		maxLength = value.password_settings.max_length;
		minSymbols = value.password_settings.min_symbols;
		minUpper = value.password_settings.min_upper;
		minLower = value.password_settings.min_lower;
	};

	onMount(async () => {
		if (!browser) return;
		if (!hasPermission('admin:settings:read')) return goto('/');

		const result = await star.getServerSettings();
		if (result == null) return toast.error('Failed to load the server settings');

		apply(result);
	});

	const save = async (e: Event) => {
		e.preventDefault();
		if (policyIssue) return toast.error(policyIssue);

		loading = true;
		const result = await star.updateServerSettings({
			allow_registration: allowRegistration,
			allow_logins: allowLogins,
			email_code_length: emailCodeLength,
			password_settings: {
				min_length: minLength,
				max_length: maxLength,
				min_symbols: minSymbols,
				min_upper: minUpper,
				min_lower: minLower
			}
		});
		loading = false;

		if ('error' in result) {
			toast.error(result.message);
			return;
		}

		apply(result);
		toast.success('Successfully updated the server settings');
	};
</script>

<Sidebar class="md:p-8 max-w-6xl mx-auto">
	<h1 class="text-2xl font-semibold">Server Settings</h1>
	<h2 class="text-lg text-primary/60">
		Configure how this star instance signs users up and in
	</h2>

	{#if settings}
		<form onsubmit={save} class="mt-16 flex flex-col gap-8">
			<Card.Root class="min-w-1/3 w-full md:w-auto">
				<Card.CardHeader>
					<Card.Title class="flex items-center gap-1">
						<Icon icon="door-enter" class="text-2xl" /> Access
					</Card.Title>
					<Card.Description class="md:max-w-2/3">
						Controls who is allowed to reach this server. These apply to every
						application authenticating through star, including this dashboard.
					</Card.Description>
				</Card.CardHeader>
				<Card.Content class="flex flex-col gap-6">
					<Label for="allow_registration" class="p-3 border rounded-lg">
						<Checkbox
							id="allow_registration"
							bind:checked={allowRegistration}
							disabled={!canModify}
						/>
						<div class="flex flex-col items-start gap-1">
							<span>Allow registration</span>
							<span class="text-primary/60 leading-5">
								Lets anyone create an account. Turn this off to only allow
								accounts you create yourself.
							</span>
						</div>
					</Label>

					<Label for="allow_logins" class="p-3 border rounded-lg">
						<Checkbox
							id="allow_logins"
							bind:checked={allowLogins}
							disabled={!canModify}
						/>
						<div class="flex flex-col items-start gap-1">
							<span>Allow logins</span>
							<span class="text-primary/60 leading-5">
								Lets users sign in. Only accounts holding the
								<code>user:login:bypass</code>
								permission may sign in while this is off, existing sessions stay
								valid either way.
							</span>
						</div>
					</Label>

					{#if !allowLogins}
						<span class="text-sm text-orange-400 leading-5">
							<Icon icon="alert-triangle" /> Logins are disabled. Make sure your
							own account has the
							<code>user:login:bypass</code>
							permission before signing out.
						</span>
					{/if}
				</Card.Content>
			</Card.Root>

			<Card.Root class="min-w-1/3 w-full md:w-auto">
				<Card.CardHeader>
					<Card.Title class="flex items-center gap-1">
						<Icon icon="mail" class="text-2xl" /> Email Verification
					</Card.Title>
					<Card.Description class="md:max-w-2/3">
						The length of the codes sent out to verify email addresses. Longer
						codes are harder to guess but more annoying to type.
					</Card.Description>
				</Card.CardHeader>
				<Card.Content class="flex flex-col gap-6">
					<div class="grid gap-2 md:max-w-1/3">
						<Label for="email_code_length">Code Length</Label>
						<Input
							id="email_code_length"
							type="number"
							min="4"
							max="12"
							bind:value={emailCodeLength}
							disabled={!canModify}
						/>
					</div>
				</Card.Content>
			</Card.Root>

			<Card.Root class="min-w-1/3 w-full md:w-auto">
				<Card.CardHeader>
					<Card.Title class="flex items-center gap-1">
						<Icon icon="password" class="text-2xl" /> Password Policy
					</Card.Title>
					<Card.Description class="md:max-w-2/3">
						The rules new passwords are checked against. Existing passwords are
						never revalidated, so raising these only affects the next time a
						password is set.
					</Card.Description>
				</Card.CardHeader>
				<Card.Content class="flex flex-col gap-6">
					<div class="flex gap-2 sm:gap-8">
						<div class="flex-1/2 gap-2 grid">
							<Label for="min_length">Minimum Length</Label>
							<Input
								id="min_length"
								type="number"
								min="8"
								max="128"
								bind:value={minLength}
								disabled={!canModify}
							/>
						</div>

						<div class="flex-1/2 gap-2 grid">
							<Label for="max_length">Maximum Length</Label>
							<Input
								id="max_length"
								type="number"
								min="8"
								max="512"
								bind:value={maxLength}
								disabled={!canModify}
							/>
						</div>
					</div>

					<div class="flex flex-col sm:flex-row gap-2 sm:gap-8">
						<div class="flex-1/3 gap-2 grid">
							<Label for="min_symbols">Minimum Symbols</Label>
							<Input
								id="min_symbols"
								type="number"
								min="0"
								max="32"
								bind:value={minSymbols}
								disabled={!canModify}
							/>
						</div>

						<div class="flex-1/3 gap-2 grid">
							<Label for="min_upper">Minimum Uppercase</Label>
							<Input
								id="min_upper"
								type="number"
								min="0"
								max="32"
								bind:value={minUpper}
								disabled={!canModify}
							/>
						</div>

						<div class="flex-1/3 gap-2 grid">
							<Label for="min_lower">Minimum Lowercase</Label>
							<Input
								id="min_lower"
								type="number"
								min="0"
								max="32"
								bind:value={minLower}
								disabled={!canModify}
							/>
						</div>
					</div>

					{#if policyIssue}
						<span class="text-sm text-red-400 leading-5">
							<Icon icon="alert-triangle" />
							{policyIssue}
						</span>
					{/if}
				</Card.Content>
				{#if canModify}
					<Card.Footer>
						<Button
							type="submit"
							class="ml-auto"
							disabled={loading || policyIssue != null}
						>
							{#if loading}
								<Spinner />
							{/if}
							Save Settings
						</Button>
					</Card.Footer>
				{/if}
			</Card.Root>
		</form>
	{/if}
</Sidebar>
