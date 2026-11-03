<script lang="ts">
    import BackupCodes from "@/components/backup-codes.svelte";
    import Button from "@/components/button.svelte";
    import CodeInput from "@/components/code-input.svelte";
    import Hidden from "@/components/hidden.svelte";
    import Modal from "@/components/modal.svelte";
    import * as star from "@/star";
    import { cn } from "@/utils";
    import { QrCode } from "@nomideusz/svelte-qr";
    import { tick } from "svelte";
    import { cubicOut } from "svelte/easing";
    import { prefersReducedMotion } from "svelte/motion";
    import { toast } from "svelte-sonner";
    import { fly } from "svelte/transition";

    const {
        onsuccess,
    }: {
        onsuccess?: () => void;
    } = $props();

    const steps = ["scan", "verify", "codes"] as const;

    const instructions = [
        "Open your authenticator app",
        "Scan the QR code with it",
        "Enter the code it shows on the next step",
    ];

    let modal = $state<Modal>();
    let codeContainer = $state<HTMLDivElement>();

    let step = $state<(typeof steps)[number]>("scan");
    let direction = $state(1);

    let enrollment = $state<star.TotpEnrollment | undefined>();
    let enrollError = $state("");

    let code = $state("");
    let attempt = $state(0);
    let verifying = $state(false);
    let verifyError = $state("");

    let backupCodes = $state<string[]>([]);

    const duration = $derived(prefersReducedMotion.current ? 0 : 220);

    $effect(() => {
        if (step !== "verify") return;

        attempt;
        tick().then(() => codeContainer?.querySelector("input")?.focus());
    });

    export const open = () => {
        direction = -1;
        step = "scan";
        enrollment = undefined;
        code = "";
        verifyError = "";
        backupCodes = [];

        modal?.open();
        enroll();
    };

    export const close = () => modal?.close();

    const goTo = (target: (typeof steps)[number]) => {
        direction = steps.indexOf(target) > steps.indexOf(step) ? 1 : -1;
        step = target;
    };

    const enroll = async () => {
        enrollError = "";

        const data = await star.enrollTotp();
        if ("error" in data) {
            enrollError = data.message;
            return;
        }

        enrollment = data;
    };

    const verify = async (value: string = code) => {
        if (verifying || value.length < 6) return;

        verifyError = "";
        verifying = true;
        const data = await star.verifyTotp(value);
        verifying = false;

        if ("error" in data) {
            verifyError = data.message;
            attempt++;
            return;
        }

        toast.success("Authenticator app enabled");
        onsuccess?.();

        if (!data.backup_codes?.length) return close();

        backupCodes = data.backup_codes;
        goTo("codes");
    };
</script>

<Modal bind:this={modal} button class="sm:w-xl">
    <div class="flex flex-col gap-6">
        <div class="flex flex-col gap-1 pr-8">
            <h1 class="text-xl font-semibold">Set up authenticator app</h1>
            <h2 class="text-sm text-muted">
                Add a rotating, time-based code as a second step when you sign
                in.
            </h2>
        </div>

        <div class="grid overflow-hidden">
            {#key step}
                <div
                    class="col-start-1 row-start-1 min-h-52"
                    in:fly={{ x: 48 * direction, duration, easing: cubicOut }}
                    out:fly={{ x: -48 * direction, duration, easing: cubicOut }}
                >
                    {#if step === "scan"}
                        <div
                            class="flex flex-col items-center gap-6 sm:flex-row sm:items-start"
                        >
                            <div
                                class="flex size-44 shrink-0 items-center justify-center rounded-xl bg-white p-2"
                            >
                                {#if enrollment}
                                    <QrCode
                                        data={enrollment.uri}
                                        size={160}
                                        padding={0}
                                        label="Authenticator setup QR code"
                                    />
                                {:else}
                                    <div
                                        class="size-full animate-pulse rounded-lg bg-black/10"
                                    ></div>
                                {/if}
                            </div>

                            <div class="flex w-full min-w-0 flex-col gap-4">
                                <ol
                                    class="flex flex-col gap-2.5 text-sm text-foreground/90"
                                >
                                    {#each instructions as instruction, index (instruction)}
                                        <li class="flex items-center gap-3">
                                            <span
                                                class="flex size-5 shrink-0 items-center justify-center rounded-full border border-background-200 font-mono text-xs text-muted select-none"
                                            >
                                                {index + 1}
                                            </span>
                                            {instruction}
                                        </li>
                                    {/each}
                                </ol>

                                <div
                                    class="flex flex-col gap-1.5 border-t border-background-200 pt-4"
                                >
                                    <span class="text-xs text-muted select-none">
                                        Can't scan it? Enter this setup key
                                        instead
                                    </span>
                                    {#if enrollment}
                                        <Hidden content={enrollment.secret} />
                                    {:else if enrollError}
                                        <p
                                            role="alert"
                                            class="text-sm text-red-300"
                                        >
                                            {enrollError}
                                        </p>
                                    {:else}
                                        <div
                                            class="h-5 w-48 animate-pulse rounded bg-background-200"
                                        ></div>
                                    {/if}
                                </div>
                            </div>
                        </div>
                    {:else if step === "verify"}
                        <div class="flex flex-col items-center gap-4 pt-4">
                            <div class="flex flex-col items-center gap-1">
                                <h3 class="font-sans text-sm font-medium">
                                    Enter the 6-digit code
                                </h3>
                                <p class="text-sm text-muted text-center">
                                    Type the code your authenticator app is
                                    currently showing.
                                </p>
                            </div>

                            <div bind:this={codeContainer}>
                                {#key attempt}
                                    <CodeInput
                                        bind:value={code}
                                        oncomplete={verify}
                                    />
                                {/key}
                            </div>

                            {#if verifyError}
                                <p
                                    role="alert"
                                    class="text-sm text-red-300 text-center"
                                >
                                    {verifyError}
                                </p>
                            {/if}
                        </div>
                    {:else}
                        <div class="flex flex-col gap-4">
                            <div class="flex flex-col gap-1">
                                <h3 class="font-sans text-sm font-medium">
                                    Save your backup codes
                                </h3>
                                <p class="text-sm text-muted">
                                    If you lose access to your authenticator
                                    app, you can sign in with one of these
                                    instead.
                                </p>
                            </div>

                            <BackupCodes codes={backupCodes} />
                        </div>
                    {/if}
                </div>
            {/key}
        </div>

        <div class="flex items-center gap-2">
            <div class="mr-auto flex items-center gap-1.5" aria-hidden="true">
                {#each steps as entry (entry)}
                    <span
                        class={cn(
                            "h-1.5 rounded-full transition-all duration-200",
                            entry === step
                                ? "w-5 bg-primary"
                                : "w-1.5 bg-background-200",
                        )}
                    ></span>
                {/each}
            </div>

            {#if step === "scan"}
                <Button variant="secondary" onclick={close}>Cancel</Button>
                <Button
                    disabled={!enrollment}
                    onclick={() => goTo("verify")}
                >
                    Continue
                </Button>
            {:else if step === "verify"}
                <Button variant="secondary" onclick={() => goTo("scan")}>
                    Back
                </Button>
                <Button
                    loading={verifying}
                    disabled={code.length < 6}
                    onclick={() => verify()}
                >
                    Enable
                </Button>
            {:else}
                <Button onclick={close}>Done</Button>
            {/if}
        </div>
    </div>
</Modal>
