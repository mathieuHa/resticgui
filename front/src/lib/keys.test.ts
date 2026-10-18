import { describe, expect, it } from 'vitest';

import { moveSelection } from './keys';

describe('moveSelection', () => {
  it('moves within the list and stops at its ends', () => {
    expect(moveSelection('ArrowDown', -1, 3)).toBe(0);
    expect(moveSelection('ArrowDown', 2, 3)).toBe(2);
    expect(moveSelection('j', 0, 3)).toBe(1);
    expect(moveSelection('ArrowUp', 0, 3)).toBe(0);
    expect(moveSelection('ArrowUp', -1, 3)).toBe(2);
    expect(moveSelection('k', 2, 3)).toBe(1);
    expect(moveSelection('Home', 2, 3)).toBe(0);
    expect(moveSelection('End', 0, 3)).toBe(2);
    expect(moveSelection('PageDown', 0, 30)).toBe(10);
    expect(moveSelection('PageUp', 5, 30)).toBe(0);
  });
  it('ignores other keys and empty lists', () => {
    expect(moveSelection('x', 0, 3)).toBeNull();
    expect(moveSelection('ArrowDown', 0, 0)).toBeNull();
  });
  it('clamps a selection left past the end of a shorter list', () => {
    expect(moveSelection('ArrowUp', 9, 3)).toBe(1);
  });
});
