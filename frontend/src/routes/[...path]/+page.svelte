<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import Icon from '@/components/icon.svelte';
	import Sidebar from '@/components/sidebar.svelte';
	import { Button } from '@/components/ui/button';
</script>

<Sidebar class="md:p-8">
	<div
		class="flex flex-col gap-4 items-center justify-center text-center min-h-dvh"
	>
		<Icon icon="error-404" class="text-5xl text-primary/60" />
		<h1 class="text-2xl font-semibold">Page not found</h1>
		<h2 class="text-lg text-primary/60 max-w-md">
			Nothing lives at <code>{page.url.pathname}</code>
		</h2>

		<Button class="mt-2" onclick={() => goto('/')}>Go home</Button>
	</div>
</Sidebar>
