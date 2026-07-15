import { describe, expect, it } from 'vitest';
import { buildContext } from './context-editor';

describe('buildContext', () => {
  it('serializes user-provided context', () => {
    expect(buildContext([
      { key: 'aws:RequestTag/team', value: 'security' },
      { key: ' ', value: 'ignored' },
    ])).toEqual({
      'aws:RequestTag/team': 'security',
    });
  });

  it('returns undefined when no user context is set', () => {
    expect(buildContext([])).toBeUndefined();
  });
});
