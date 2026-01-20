// ui/src/components/copy-entity-button.test.tsx
import { describe, it, expect } from 'vitest';
import { render, screen } from '../test/utils';
import { CopyEntityButton } from './copy-entity-button';

describe('CopyEntityButton', () => {
  it('renders button', () => {
    render(<CopyEntityButton data={{ test: 'value' }} />);
    expect(screen.getByRole('button')).toBeInTheDocument();
  });

  it('accepts complex data objects', () => {
    const testData = { key: 'value', nested: { prop: 123 } };
    render(<CopyEntityButton data={testData} />);
    // The CopyEntityButton wraps Mantine's CopyButton which handles clipboard
    expect(screen.getByRole('button')).toBeInTheDocument();
  });
});
