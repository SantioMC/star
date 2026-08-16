<script lang="ts">
	import * as InputOTP from '@/components/ui/input-otp';
	import { REGEXP_ONLY_DIGITS } from 'bits-ui';

	let {
		value = $bindable(''),
		disabled = false
	}: {
		value?: string;
		disabled?: boolean;
	} = $props();
</script>

<InputOTP.Root
	maxlength={6}
	pattern={REGEXP_ONLY_DIGITS}
	bind:value
	{disabled}
	autocomplete="one-time-code"
>
	{#snippet children({ cells })}
		<InputOTP.Group>
			{#each cells.slice(0, 3) as cell (cell)}
				<InputOTP.Slot {cell} class="size-10 text-base" />
			{/each}
		</InputOTP.Group>
		<InputOTP.Separator />
		<InputOTP.Group>
			{#each cells.slice(3, 6) as cell (cell)}
				<InputOTP.Slot {cell} class="size-10 text-base" />
			{/each}
		</InputOTP.Group>
	{/snippet}
</InputOTP.Root>
