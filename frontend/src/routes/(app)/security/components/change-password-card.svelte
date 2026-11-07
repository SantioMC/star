<script lang="ts">
    import Button from "@/components/button.svelte";
    import Card from "@/components/card.svelte";
    import Icon from "@/components/icon.svelte";
    import Input from "@/components/input.svelte";
    import { toast } from "svelte-sonner";
    import * as star from "@/star";

    const { onchanged }: { onchanged: () => void } = $props();

    let currentPassword = $state("");
    let newPassword = $state("");
    let confirmPassword = $state("");

    let error = $state("");
    let loading = $state(false);

    const changePassword = async (e: Event) => {
        e.preventDefault();
        error = "";

        if (newPassword !== confirmPassword) {
            error =
                "Your new password does not match, please make sure there is not a spelling mistake";
            return;
        } else if (!currentPassword || !newPassword) {
            error = "Please provide your current and new password";
            return;
        }

        loading = true;
        const result = await star.changePassword(currentPassword, newPassword);
        loading = false;

        if ("error" in result) {
            error = result.message;
            return;
        }

        currentPassword = "";
        newPassword = "";
        confirmPassword = "";

        toast.success(
            "Your password has been updated, your other sessions were signed out",
        );

        onchanged();
    };
</script>

<Card class="flex flex-col gap-4">
    <h2 class="flex flex-row items-center gap-2 text-sm font-semibold">
        <Icon icon="lock" /> Change Password
    </h2>
    <form onsubmit={changePassword}>
        <div class="flex flex-col gap-4 sm:px-4">
            <div class="flex flex-col sm:flex-row sm:items-center gap-2 sm:gap-4">
                <label class="text-sm sm:w-48 shrink-0 select-none" for="currentPassword">
                    Current Password
                </label>
                <Input
                    bind:value={currentPassword}
                    id="currentPassword"
                    type="password"
                    required
                    placeholder={"•".repeat(12)}
                />
            </div>

            <div class="flex flex-col sm:flex-row sm:items-center gap-2 sm:gap-4">
                <label class="text-sm sm:w-48 shrink-0 select-none" for="password">New Password</label>
                <Input bind:value={newPassword} id="password" type="password" />
            </div>

            <div class="flex flex-col sm:flex-row sm:items-center gap-2 sm:gap-4">
                <label class="text-sm sm:w-48 shrink-0 select-none" for="confirmPassword">
                    Confirm Password
                </label>
                <Input
                    bind:value={confirmPassword}
                    id="confirmPassword"
                    type="password"
                />
            </div>
        </div>
        <div class="mt-4 flex flex-row gap-4 items-center">
            <p class="ml-auto text-red-400/70 text-right">{error}</p>
            <Button {loading} onclick={changePassword}>Change Password</Button>
        </div>
    </form>
</Card>
