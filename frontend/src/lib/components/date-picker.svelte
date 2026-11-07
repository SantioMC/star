<script lang="ts">
    import { cn } from "@/utils";
    import {
        CalendarDate,
        getLocalTimeZone,
        parseDate,
        today,
        type DateValue,
    } from "@internationalized/date";
    import { DatePicker } from "bits-ui";
    import Icon from "./icon.svelte";
    import Select from "./select.svelte";

    let {
        value = $bindable(""),
        label,
        name,
        min,
        max,
        required = false,
        disabled = false,
    }: {
        value?: string;
        label: string;
        name?: string;
        min: CalendarDate;
        max: CalendarDate;
        required?: boolean;
        disabled?: boolean;
    } = $props();

    const locale = navigator.language;
    const selected = $derived(value ? parseDate(value) : undefined);

    let view = $state<DateValue>(today(getLocalTimeZone()));

    const monthName = new Intl.DateTimeFormat(locale, { month: "long" });
    const months = Array.from({ length: 12 }, (_, index) => ({
        value: String(index + 1),
        label: monthName.format(new Date(2000, index, 1)),
    }));

    const years = $derived(
        Array.from({ length: max.year - min.year + 1 }, (_, index) => {
            const year = String(max.year - index);
            return { value: year, label: year };
        }),
    );

    const navigation =
        "flex size-8 shrink-0 items-center justify-center rounded-md text-muted transition-colors hover:bg-background-200 hover:text-foreground disabled:pointer-events-none disabled:opacity-30";

    const row = "flex border-b-0! hover:bg-transparent!";
</script>

<DatePicker.Root
    bind:value={
        () => selected,
        (next) => {
            value = next?.toString() ?? "";
        }
    }
    bind:placeholder={view}
    {locale}
    {required}
    {disabled}
    minValue={min}
    maxValue={max}
    weekdayFormat="short"
    calendarLabel={label}
    fixedWeeks
>
    <div class="grid gap-2">
        <DatePicker.Label class="text-sm leading-none font-medium select-none">
            {label}
        </DatePicker.Label>

        <DatePicker.Input
            {name}
            class={cn(
                "border-background-200 bg-background dark:bg-input/30 flex h-9 w-full min-w-0 items-center rounded-md border py-1 pr-1 pl-3 text-base shadow-xs transition-[color,box-shadow] md:text-sm",
                "focus-within:border-foreground/80 focus-within:ring-foreground/50 focus-within:ring-1",
                "data-disabled:cursor-not-allowed data-disabled:opacity-50",
            )}
        >
            {#snippet children({ segments })}
                {#each segments as { part, value: text }, index (index)}
                    {#if part === "literal"}
                        <DatePicker.Segment {part} class="px-0.5 text-muted">
                            {text}
                        </DatePicker.Segment>
                    {:else}
                        <DatePicker.Segment
                            {part}
                            class="rounded-sm px-0.5 tabular-nums outline-none focus:bg-background-200 aria-[valuetext=Empty]:text-muted"
                        >
                            {text}
                        </DatePicker.Segment>
                    {/if}
                {/each}

                <DatePicker.Trigger class={cn(navigation, "ml-auto size-7")}>
                    <Icon icon="calendar" />
                </DatePicker.Trigger>
            {/snippet}
        </DatePicker.Input>
    </div>

    <DatePicker.Content
        align="end"
        alignOffset={-4}
        sideOffset={8}
        class="z-50 rounded-lg border border-background-200 bg-background-100 p-3 text-sm text-foreground shadow-xl shadow-black/40"
    >
        <DatePicker.Calendar class="w-72">
            {#snippet children({ months: grids, weekdays })}
                <DatePicker.Header class="flex items-center gap-1">
                    <DatePicker.PrevButton class={navigation}>
                        <Icon icon="chevron-left" />
                    </DatePicker.PrevButton>

                    <Select
                        value={String(view.month)}
                        options={months}
                        onchange={(month) => {
                            view = view.set({ month: Number(month) });
                        }}
                        class="h-8 min-w-0 flex-1 px-2 text-sm"
                    />

                    <Select
                        value={String(view.year)}
                        options={years}
                        onchange={(year) => {
                            view = view.set({ year: Number(year) });
                        }}
                        class="h-8 w-20 shrink-0 px-2 text-sm"
                    />

                    <DatePicker.NextButton class={navigation}>
                        <Icon icon="chevron-right" />
                    </DatePicker.NextButton>
                </DatePicker.Header>

                {#each grids as month (month.value.toString())}
                    <DatePicker.Grid class="mt-3 w-full border-collapse select-none">
                        <DatePicker.GridHead>
                            <DatePicker.GridRow class={row}>
                                {#each weekdays as weekday (weekday)}
                                    <DatePicker.HeadCell
                                        class="flex-1 py-0! text-center! text-xs font-normal text-muted"
                                    >
                                        {weekday.slice(0, 2)}
                                    </DatePicker.HeadCell>
                                {/each}
                            </DatePicker.GridRow>
                        </DatePicker.GridHead>

                        <DatePicker.GridBody>
                            {#each month.weeks as week, index (index)}
                                <DatePicker.GridRow class={cn(row, "mt-0.5")}>
                                    {#each week as date (date.toString())}
                                        <DatePicker.Cell
                                            {date}
                                            month={month.value}
                                            class="flex-1 py-0!"
                                        >
                                            <DatePicker.Day
                                                class={cn(
                                                    "mx-auto flex size-9 items-center justify-center rounded-md transition-colors outline-none hover:bg-background-200",
                                                    "focus-visible:ring-foreground/50 focus-visible:ring-1",
                                                    "data-today:font-semibold data-today:text-primary",
                                                    "data-outside-month:text-muted/50",
                                                    "data-selected:bg-primary data-selected:text-primary-foreground! data-selected:hover:bg-primary",
                                                    "data-disabled:pointer-events-none data-disabled:text-muted/30",
                                                )}
                                            />
                                        </DatePicker.Cell>
                                    {/each}
                                </DatePicker.GridRow>
                            {/each}
                        </DatePicker.GridBody>
                    </DatePicker.Grid>
                {/each}
            {/snippet}
        </DatePicker.Calendar>
    </DatePicker.Content>
</DatePicker.Root>
