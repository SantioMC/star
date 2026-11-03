import type { Attachment } from "svelte/attachments";
import { positionPopover } from "./popover";

let tooltipCount = 0;

export const tooltip =
  (text: string | null | undefined): Attachment<HTMLElement> =>
  (node) => {
    if (!text) return;

    const element = document.createElement("div");
    element.id = `tooltip-${++tooltipCount}`;
    element.popover = "manual";
    element.role = "tooltip";
    element.textContent = text;
    element.className =
      "pointer-events-none fixed inset-auto m-0 max-w-xs rounded-md border border-background-200 bg-background-100 px-2 py-1 text-xs text-foreground shadow-lg shadow-black/40 animate-in fade-in-0 duration-100";

    document.body.append(element);

    const describedBy = node.getAttribute("aria-describedby");
    node.setAttribute(
      "aria-describedby",
      [describedBy, element.id].filter(Boolean).join(" "),
    );

    let timer: ReturnType<typeof setTimeout> | undefined;

    const show = () => {
      if (element.matches(":popover-open")) return;

      element.showPopover();
      positionPopover(node, element, "center");
    };

    const hide = () => {
      clearTimeout(timer);
      if (element.matches(":popover-open")) element.hidePopover();
    };

    const handlePointerEnter = (event: PointerEvent) => {
      if (event.pointerType === "touch") return;
      timer = setTimeout(show, 400);
    };

    const handleFocus = () => {
      if (node.matches(":focus-visible")) show();
    };

    const handleKeydown = (event: KeyboardEvent) => {
      if (event.key === "Escape") hide();
    };

    node.addEventListener("pointerenter", handlePointerEnter);
    node.addEventListener("pointerleave", hide);
    node.addEventListener("pointerdown", hide);
    node.addEventListener("focus", handleFocus);
    node.addEventListener("blur", hide);
    document.addEventListener("keydown", handleKeydown);
    window.addEventListener("scroll", hide, true);

    return () => {
      hide();

      node.removeEventListener("pointerenter", handlePointerEnter);
      node.removeEventListener("pointerleave", hide);
      node.removeEventListener("pointerdown", hide);
      node.removeEventListener("focus", handleFocus);
      node.removeEventListener("blur", hide);
      document.removeEventListener("keydown", handleKeydown);
      window.removeEventListener("scroll", hide, true);

      if (describedBy) node.setAttribute("aria-describedby", describedBy);
      else node.removeAttribute("aria-describedby");

      element.remove();
    };
  };
