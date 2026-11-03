import * as star from "$lib/star";
import type { SelectOption } from "@/components/select.svelte";
import { writable } from "svelte/store";

interface Font {
  name: string;
  category: "Sans" | "Serif" | "Display";
  variable: boolean;
}

export const FONTS: Font[] = [
  { name: "Instrument Sans", category: "Sans", variable: true },
  { name: "Manrope", category: "Sans", variable: true },
  { name: "Plus Jakarta Sans", category: "Sans", variable: true },
  { name: "Figtree", category: "Sans", variable: true },
  { name: "Outfit", category: "Sans", variable: true },
  { name: "Work Sans", category: "Sans", variable: true },
  { name: "Lexend", category: "Sans", variable: true },
  { name: "Space Grotesk", category: "Sans", variable: true },
  { name: "Bricolage Grotesque", category: "Sans", variable: true },
  { name: "Fraunces", category: "Serif", variable: true },
  { name: "Newsreader", category: "Serif", variable: true },
  { name: "Source Serif 4", category: "Serif", variable: true },
  { name: "Playfair Display", category: "Serif", variable: true },
  { name: "Instrument Serif", category: "Serif", variable: false },
  { name: "Exo", category: "Display", variable: true },
  { name: "Sora", category: "Display", variable: true },
  { name: "Syne", category: "Display", variable: true },
  { name: "Unbounded", category: "Display", variable: true },
];

export const fontOptions: SelectOption[] = FONTS.map((font) => ({
  value: font.name,
  label: font.name,
  hint: font.category,
}));

export const LAYOUTS: {
  value: star.BrandingLayout;
  label: string;
  description: string;
}[] = [
  {
    value: "centered",
    label: "Centered",
    description: "A single card in the middle of the page.",
  },
  {
    value: "split",
    label: "Split",
    description: "A solid brand panel beside the form.",
  },
  {
    value: "minimal",
    label: "Minimal",
    description: "No card, the form sits directly on the page.",
  },
];

export const DEFAULT_BRANDING: star.Branding = {
  name: "Star",
  accent_color: "#e0a458",
  heading_font: "Exo",
  body_font: "Instrument Sans",
  layout: "centered",
  logo_url: null,
  support_email: null,
  privacy_url: null,
  terms_url: null,
};

export const branding = writable<star.Branding>(DEFAULT_BRANDING);

export const loadFont = (name: string) => {
  const existing = document.head.querySelector(
    `link[data-font="${CSS.escape(name)}"]`,
  );

  if (existing) return;

  const font = FONTS.find((it) => it.name === name);
  const family = name.replaceAll(" ", "+");
  const axes = font?.variable ? ":wght@400..700" : "";

  const link = document.createElement("link");
  link.rel = "stylesheet";
  link.dataset.font = name;
  link.href = `https://fonts.googleapis.com/css2?family=${family}${axes}&display=swap`;

  document.head.append(link);
};

export const fontStack = (name: string) => `"${name}", sans-serif`;

const applyBranding = (value: star.Branding) => {
  const root = document.documentElement.style;

  loadFont(value.heading_font);
  loadFont(value.body_font);

  root.setProperty("--color-primary", value.accent_color);
  root.setProperty("--font-serif", fontStack(value.heading_font));
  root.setProperty("--font-sans", fontStack(value.body_font));
};

export const loadBranding = async () => {
  const result = await star.getBranding();
  if (result) branding.set(result);

  branding.subscribe(applyBranding);
};
