<script lang="ts">
    import * as star from "$lib/star";
    import {
        branding,
        fontOptions,
        fontStack,
        LAYOUTS,
        loadFont,
    } from "@/branding";
    import Button from "@/components/button.svelte";
    import Card from "@/components/card.svelte";
    import EmailTemplatePreview from "@/components/email-template-preview.svelte";
    import Icon from "@/components/icon.svelte";
    import Input from "@/components/input.svelte";
    import Select from "@/components/select.svelte";
    import SettingsForm from "@/components/settings-form.svelte";
    import { Label } from "@/components/ui/label";
    import { toast } from "svelte-sonner";

    type PendingLogo = { file: File; preview: string } | "remove" | null;

    let name = $state($branding.name);
    let accentColor = $state($branding.accent_color);
    let headingFont = $state($branding.heading_font);
    let bodyFont = $state($branding.body_font);
    let layout = $state($branding.layout);
    let supportEmail = $state($branding.support_email ?? "");
    let privacyUrl = $state($branding.privacy_url ?? "");
    let termsUrl = $state($branding.terms_url ?? "");
    let logo = $state<PendingLogo>(null);
    let logoInput = $state<HTMLInputElement>();

    const validAccent = $derived(/^#[0-9a-fA-F]{6}$/.test(accentColor));
    const previewAccent = $derived(validAccent ? accentColor : $branding.accent_color);

    const logoPreview = $derived.by(() => {
        if (logo === "remove") return null;
        if (logo) return logo.preview;

        return $branding.logo_url;
    });

    const withCurrent = (font: string) =>
        fontOptions.some((option) => option.value === font)
            ? fontOptions
            : [{ value: font, label: font, hint: "Custom" }, ...fontOptions];

    const headingOptions = $derived(withCurrent($branding.heading_font));
    const bodyOptions = $derived(withCurrent($branding.body_font));

    const fields = $derived({
        name: name.trim(),
        accent_color: accentColor.toLowerCase(),
        heading_font: headingFont,
        body_font: bodyFont,
        layout,
        support_email: supportEmail.trim(),
        privacy_url: privacyUrl.trim(),
        terms_url: termsUrl.trim(),
    });

    const fieldsDirty = $derived(
        fields.name !== $branding.name ||
            fields.accent_color !== $branding.accent_color.toLowerCase() ||
            fields.heading_font !== $branding.heading_font ||
            fields.body_font !== $branding.body_font ||
            fields.layout !== $branding.layout ||
            fields.support_email !== ($branding.support_email ?? "") ||
            fields.privacy_url !== ($branding.privacy_url ?? "") ||
            fields.terms_url !== ($branding.terms_url ?? ""),
    );

    const dirty = $derived(fieldsDirty || logo !== null);

    const issue = $derived.by(() => {
        if (!fields.name) return "A display name is required";
        if (fields.name.length > 64) return "The display name may not exceed 64 characters";
        if (!validAccent) return "The accent color must be a hex color like #e0a458";

        if (fields.support_email && !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(fields.support_email)) {
            return "The support email is not a valid email address";
        }

        const links = [fields.privacy_url, fields.terms_url].filter(Boolean);
        if (links.some((link) => !link.startsWith("https://"))) {
            return "Privacy and terms links must start with https://";
        }

        return null;
    });

    $effect(() => {
        loadFont(headingFont);
        loadFont(bodyFont);
    });

    const pickLogo = (event: Event) => {
        const input = event.currentTarget as HTMLInputElement;
        const file = input.files?.[0];
        input.value = "";

        if (!file) return;

        if (file.size > 512 * 1024) {
            toast.error("Logos may not be larger than 512 KB");
            return;
        }

        const reader = new FileReader();
        reader.onload = () => {
            logo = { file, preview: reader.result as string };
        };

        reader.readAsDataURL(file);
    };

    const applyLogo = async () => {
        if (logo === "remove") return star.deleteBrandingLogo();
        if (logo) return star.uploadBrandingLogo(logo.file);

        return null;
    };

    const save = async () => {
        if (fieldsDirty) {
            const result = await star.updateBranding(fields);
            if ("error" in result) {
                toast.error(result.message);
                return;
            }

            branding.set(result);
        }

        const logoResult = await applyLogo();
        if (logoResult && "error" in logoResult) {
            toast.error(logoResult.message);
            return;
        }

        if (logoResult) branding.set(logoResult);

        logo = null;
        toast.success("Successfully updated the branding");
    };

    const discard = () => {
        name = $branding.name;
        accentColor = $branding.accent_color;
        headingFont = $branding.heading_font;
        bodyFont = $branding.body_font;
        layout = $branding.layout;
        supportEmail = $branding.support_email ?? "";
        privacyUrl = $branding.privacy_url ?? "";
        termsUrl = $branding.terms_url ?? "";
        logo = null;
    };
</script>

<SettingsForm
    title="Branding & Appearance"
    description="Customize how sign-in pages, the dashboard and emails look to your users"
    {dirty}
    {issue}
    onsave={save}
    ondiscard={discard}
>
    <Card
        title="Identity"
        icon="id-badge-2"
        description="Shown in the dashboard, on sign-in pages, in the browser tab and in emails."
    >
        <div class="mt-6 grid gap-6 sm:grid-cols-[auto_minmax(0,1fr)] sm:items-start">
            <div class="flex items-center gap-4">
                <div
                    class="flex size-16 shrink-0 items-center justify-center overflow-hidden rounded-xl border border-background-200 bg-background"
                >
                    {#if logoPreview}
                        <img src={logoPreview} alt="Logo preview" class="size-full object-cover" />
                    {:else}
                        <span
                            class="flex size-full items-center justify-center bg-primary text-2xl text-primary-foreground"
                            style:--color-primary={previewAccent}
                        >
                            <Icon icon="star" filled />
                        </span>
                    {/if}
                </div>
                <div class="flex flex-col gap-2">
                    <div class="flex items-center gap-2">
                        <Button
                            type="button"
                            variant="secondary"
                            onclick={() => logoInput?.click()}
                        >
                            <Icon icon="upload" />
                            Upload logo
                        </Button>
                        {#if logoPreview}
                            <Button
                                type="button"
                                variant="ghost"
                                onclick={() => (logo = $branding.logo_url ? "remove" : null)}
                            >
                                Remove
                            </Button>
                        {/if}
                    </div>
                    <p class="text-xs text-muted select-none">
                        PNG, JPEG or WebP up to 512 KB, square works best
                    </p>
                    <input
                        bind:this={logoInput}
                        type="file"
                        accept="image/png,image/jpeg,image/webp"
                        class="hidden"
                        onchange={pickLogo}
                    />
                </div>
            </div>

            <div class="grid gap-2 sm:max-w-sm sm:justify-self-end sm:w-full">
                <Label for="brandName" class="select-none">Display Name</Label>
                <Input bind:value={name} id="brandName" maxlength={64} required />
            </div>
        </div>
    </Card>

    <Card
        title="Appearance"
        icon="palette"
        description="The accent color is used for buttons, links and highlights. Fonts are loaded from Google Fonts."
    >
        <div class="mt-6 grid gap-6 sm:grid-cols-3">
            <div class="grid gap-2 content-start">
                <Label for="accentColor" class="select-none">Accent Color</Label>
                <div class="flex items-center gap-2">
                    <input
                        type="color"
                        aria-label="Pick accent color"
                        value={previewAccent}
                        oninput={(event) => (accentColor = event.currentTarget.value)}
                        class="h-9 w-11 shrink-0 cursor-pointer rounded-md border border-background-200 bg-background p-1 disabled:cursor-not-allowed"
                    />
                    <Input
                        bind:value={accentColor}
                        id="accentColor"
                        maxlength={7}
                        spellcheck={false}
                        class="font-mono"
                        aria-invalid={!validAccent}
                    />
                </div>
            </div>
            <div class="grid gap-2 content-start">
                <Label for="headingFont" class="select-none">Heading Font</Label>
                <Select bind:value={headingFont} id="headingFont" options={headingOptions} />
            </div>
            <div class="grid gap-2 content-start">
                <Label for="bodyFont" class="select-none">Body Font</Label>
                <Select bind:value={bodyFont} id="bodyFont" options={bodyOptions} />
            </div>
        </div>

        <div
            class="mt-6 rounded-lg border border-background-200 bg-background p-6"
            style:font-family={fontStack(bodyFont)}
            style:--color-primary={previewAccent}
        >
            <p
                class="text-xl font-semibold tracking-tight"
                style:font-family={fontStack(headingFont)}
            >
                Login to your account
            </p>
            <p class="mt-1 text-sm text-muted">
                Enter your account details below to login to your account
            </p>
            <div class="mt-5 flex flex-wrap items-center gap-4">
                <span
                    class="rounded-lg bg-primary px-3 py-1.5 text-sm text-primary-foreground select-none"
                >
                    Continue
                </span>
                <span class="text-sm text-primary select-none">
                    Create an account
                </span>
            </div>
        </div>
    </Card>

    <Card
        title="Sign-in Layout"
        icon="layout"
        description="Used for login, registration, email verification and authorization pages."
    >
        <div
            role="radiogroup"
            aria-label="Sign-in layout"
            class="mt-6 grid gap-3 sm:grid-cols-3"
            style:--color-primary={previewAccent}
        >
            {#each LAYOUTS as option (option.value)}
                <label
                    class="flex flex-col gap-3 rounded-lg border border-background-200 p-3 transition-colors has-checked:border-primary/60 has-checked:bg-hover has-enabled:cursor-pointer has-enabled:hover:bg-hover"
                >
                    <input
                        type="radio"
                        name="layout"
                        value={option.value}
                        bind:group={layout}
                        class="sr-only"
                    />
                    <div
                        aria-hidden="true"
                        class="relative flex aspect-[16/10] overflow-hidden rounded-md border border-background-200 bg-background"
                    >
                        {#if option.value === "split"}
                            <div class="h-full w-2/5 bg-primary"></div>
                            <div class="flex flex-1 flex-col justify-center gap-1.5 px-4">
                                <div class="h-1.5 w-1/2 rounded-full bg-foreground/50"></div>
                                <div class="h-1.5 w-full rounded-full bg-background-200"></div>
                                <div class="h-1.5 w-full rounded-full bg-background-200"></div>
                                <div class="mt-1 h-2 w-full rounded-sm bg-primary"></div>
                            </div>
                        {:else if option.value === "minimal"}
                            <div class="flex flex-1 flex-col p-2">
                                <div class="size-2 rounded-sm bg-primary"></div>
                                <div class="mx-auto flex w-2/5 flex-1 flex-col justify-center gap-1.5">
                                    <div class="h-1.5 w-1/2 rounded-full bg-foreground/50"></div>
                                    <div class="h-1.5 w-full rounded-full bg-background-200"></div>
                                    <div class="h-1.5 w-full rounded-full bg-background-200"></div>
                                    <div class="mt-1 h-2 w-full rounded-sm bg-primary"></div>
                                </div>
                                <div class="h-px w-full bg-background-200"></div>
                            </div>
                        {:else}
                            <div class="flex flex-1 flex-col items-center justify-center gap-1.5">
                                <div class="size-2 rounded-sm bg-primary"></div>
                                <div
                                    class="flex w-2/5 flex-col gap-1.5 rounded-sm border border-background-200 bg-background-100 p-2"
                                >
                                    <div class="h-1.5 w-1/2 rounded-full bg-foreground/50"></div>
                                    <div class="h-1.5 w-full rounded-full bg-background-200"></div>
                                    <div class="mt-1 h-2 w-full rounded-sm bg-primary"></div>
                                </div>
                            </div>
                        {/if}
                    </div>
                    <div class="flex items-start justify-between gap-2">
                        <div class="flex flex-col gap-0.5">
                            <span class="text-sm select-none">{option.label}</span>
                            <span class="text-xs text-muted select-none">{option.description}</span>
                        </div>
                        {#if layout === option.value}
                            <Icon icon="circle-check-filled" class="text-primary text-base shrink-0" />
                        {/if}
                    </div>
                </label>
            {/each}
        </div>
    </Card>

    <Card
        title="Support & Legal"
        icon="lifebuoy"
        description="Linked in the footer of sign-in pages and emails. Leave a field empty to hide it."
    >
        <div class="mt-6 grid gap-6 sm:grid-cols-2">
            <div class="grid gap-2 sm:col-span-2 sm:max-w-sm">
                <Label for="supportEmail" class="select-none">Support Email</Label>
                <Input
                    bind:value={supportEmail}
                    id="supportEmail"
                    type="email"
                    placeholder="support@example.com"
                />
            </div>
            <div class="grid gap-2">
                <Label for="privacyUrl" class="select-none">Privacy Policy URL</Label>
                <Input
                    bind:value={privacyUrl}
                    id="privacyUrl"
                    type="url"
                    placeholder="https://example.com/privacy"
                />
            </div>
            <div class="grid gap-2">
                <Label for="termsUrl" class="select-none">Terms of Service URL</Label>
                <Input
                    bind:value={termsUrl}
                    id="termsUrl"
                    type="url"
                    placeholder="https://example.com/terms"
                />
            </div>
        </div>
    </Card>

    {#snippet preview()}
        <EmailTemplatePreview branding={fields} removeLogo={logo === "remove"} />
    {/snippet}
</SettingsForm>
