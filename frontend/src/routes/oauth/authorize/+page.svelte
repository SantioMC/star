<script lang="ts">
	import { page } from '$app/state';
	import { Button } from '@/components/ui/button';
	import * as Card from '@/components/ui/card';

	let error = page.url.searchParams.get('error');

	const submit = async () => {
		if (!!error) {
			history.back();
			return;
		}
	};
</script>

<div class="h-full w-full flex items-center justify-center">
	<form onsubmit={submit} class="w-full sm:w-fit">
		<Card.Root class="-my-4 w-full sm:w-sm">
			{#if error}
				<Card.Header>
					<Card.Title>An error has occurred</Card.Title>
					<Card.Description>
						Failed to authorize your request, please try again or contact the
						site administrator if the problem persists.
					</Card.Description>
				</Card.Header>
				<Card.Content
					class="w-96 flex items-center flex-col gap-2 text-red-300"
				>
					{error}
				</Card.Content>
				<Card.Footer class="flex-col gap-2">
					<Button type="submit" class="w-full">Go back</Button>
				</Card.Footer>
			{/if}
		</Card.Root>
	</form>
</div>
