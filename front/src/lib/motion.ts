/** Whether the user asked their system for less motion. */
export function reducedMotion(): boolean {
  return typeof matchMedia === 'function' && matchMedia('(prefers-reduced-motion: reduce)').matches;
}

/** A transition length that drops to 0 under reduced motion. */
export function ms(duration: number): number {
  return reducedMotion() ? 0 : duration;
}
