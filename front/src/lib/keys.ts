/** Where a list selection moves for a key; null when the key does not move it. */
export function moveSelection(key: string, current: number, length: number, page = 10): number | null {
  if (length === 0) return null;
  const last = length - 1;
  const from = current < 0 ? -1 : Math.min(current, last);
  switch (key) {
    case 'ArrowDown':
    case 'j':
      return Math.min(last, from + 1);
    case 'ArrowUp':
    case 'k':
      return from < 0 ? last : Math.max(0, from - 1);
    case 'Home':
      return 0;
    case 'End':
      return last;
    case 'PageDown':
      return Math.min(last, Math.max(0, from) + page);
    case 'PageUp':
      return Math.max(0, from - page);
    default:
      return null;
  }
}

/** Whether a key event comes from a field where typing must not trigger shortcuts. */
export function typing(e: KeyboardEvent): boolean {
  const t = e.target as HTMLElement | null;
  return !!t && (t.isContentEditable || ['INPUT', 'TEXTAREA', 'SELECT'].includes(t.tagName));
}
