<script lang="ts">
    import { tooltip } from "@/tooltip";
    import Icon from "@/components/icon.svelte";
    import { cn } from "@/utils";
    import { toast } from "svelte-sonner";
    import type { ClassValue } from "svelte/elements";

    const props: {
        content?: string;
        toast?: boolean;
        class?: ClassValue;
    } = $props();

    let clicked = $state(false);

    const click = async () => {
        if (clicked) return;
        clicked = true;

        await navigator.clipboard
            .writeText(props.content ?? "")
            .catch(() => null);

        if (props.toast !== false) {
            toast.success("Copied to clipboard");
        }

        setTimeout(() => {
            clicked = false;
        }, 1500);
    };
</script>

<button
    type="button"
    aria-label="Copy to clipboard"
    {@attach tooltip("Copy to clipboard")}
    onclick={click}
    class={cn(
        "inline-flex shrink-0 items-center justify-center rounded p-0.5 text-muted transition-colors hover:text-foreground",
        clicked && "text-emerald-300 hover:text-emerald-300",
        props.class,
    )}
>
    <Icon icon={clicked ? "check" : "copy"} />
</button>
