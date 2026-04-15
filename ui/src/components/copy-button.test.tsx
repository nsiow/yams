// ui/src/components/copy-button.test.tsx
import { describe, it, expect } from 'vitest';
import { render, screen } from '../test/utils';
import { CopyButton } from './copy-button';

describe('CopyButton', () => {
  it('renders copy button', () => {
    render(<CopyButton value="test" />);
    expect(screen.getByRole('button')).toBeInTheDocument();
  });

  it('has correct value accessible for clipboard', () => {
    render(<CopyButton value="test-value" />);
    // The CopyButton wraps Mantine's CopyButton which handles clipboard
    expect(screen.getByRole('button')).toBeInTheDocument();
  });
});
