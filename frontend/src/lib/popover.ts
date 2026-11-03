export type PopoverAlign = "start" | "center" | "end";

const alignedLeft = (rect: DOMRect, width: number, align: PopoverAlign) => {
  if (align === "end") return rect.right - width;
  if (align === "center") return rect.left + rect.width / 2 - width / 2;

  return rect.left;
};

export const positionPopover = (
  trigger: HTMLElement,
  popover: HTMLElement,
  align: PopoverAlign = "start",
  matchWidth: boolean = false,
) => {
  const rect = trigger.getBoundingClientRect();

  if (matchWidth) {
    popover.style.minWidth = `${rect.width}px`;
  }

  const height = popover.offsetHeight;
  const width = popover.offsetWidth;

  const spaceBelow = window.innerHeight - rect.bottom;
  const flip = spaceBelow < height + 12 && rect.top > spaceBelow;
  const top = flip ? rect.top - height - 4 : rect.bottom + 4;

  const preferredLeft = alignedLeft(rect, width, align);
  const maxLeft = window.innerWidth - width - 8;
  const left = Math.max(8, Math.min(preferredLeft, maxLeft));

  popover.style.top = `${top}px`;
  popover.style.left = `${left}px`;
};

export const closeOnViewportChange = (popover: HTMLElement) => {
  const close = (event: Event) => {
    if (event.target instanceof Node && popover.contains(event.target)) return;
    popover.hidePopover();
  };

  window.addEventListener("scroll", close, true);
  window.addEventListener("resize", close);

  return () => {
    window.removeEventListener("scroll", close, true);
    window.removeEventListener("resize", close);
  };
};
