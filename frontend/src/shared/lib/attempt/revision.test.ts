import { describe, expect, it } from 'vitest';
import { parseRevision } from './revision';

describe('parseRevision', () => {
  it.each([
    [null, 0],
    [undefined, 0],
    ['', 0],
    ['0', 0],
    ['3', 3],
    ['12', 12],
    ['broken', 0],
    ['-1', 0],
    [3, 3],
  ])('normalizes %s to %s', (raw, expected) => expect(parseRevision(raw)).toBe(expected));
});
