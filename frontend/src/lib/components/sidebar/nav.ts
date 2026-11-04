import { page } from "$app/state";

export const navRowClass =
  "flex items-center gap-1 px-2.5 py-1.5 data-[selected='true']:font-medium text-foreground/90 data-[selected='true']:text-foreground transition-colors hover:bg-hover data-[selected='true']:bg-hover text-sm rounded-r-lg border-transparent data-[selected='true']:border-primary border-l cursor-pointer";

export const isActive = (href: string): boolean => {
  if (href.includes("#")) {
    return page.url.pathname === href.split("#")[0];
  }

  if (href === "/") return page.url.pathname === "/";
  return page.url.pathname.startsWith(href);
};
