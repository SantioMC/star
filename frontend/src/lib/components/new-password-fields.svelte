<script lang="ts">
    import Input from "@/components/input.svelte";
    import { Label } from "@/components/ui/label";

    let {
        password = $bindable(""),
        confirmPassword = $bindable(""),
    }: {
        password?: string;
        confirmPassword?: string;
    } = $props();
</script>

<div class="flex flex-col gap-6">
    <div class="grid gap-2">
        <Label for="password" class="select-none">New Password</Label>
        <Input
            bind:value={password}
            id="password"
            type="password"
            autocomplete="new-password"
            required
        />
    </div>
    <div class="grid gap-2">
        <Label for="confirm-password" class="select-none">
            Confirm Password
        </Label>
        <Input
            bind:value={confirmPassword}
            id="confirm-password"
            type="password"
            autocomplete="new-password"
            required
        />
    </div>
</div>
