// ui/src/components/collapsible-card.test.tsx
import { describe, it, expect } from 'vitest';
import { render, screen } from '../test/utils';
import userEvent from '@testing-library/user-event';
import { CollapsibleCard } from './collapsible-card';

describe('CollapsibleCard', () => {
  it('renders title', () => {
    render(
      <CollapsibleCard title="Test Title">
        <p>Content</p>
      </CollapsibleCard>
    );
    expect(screen.getByText('Test Title')).toBeInTheDocument();
  });

  it('renders children when open by default', () => {
    render(
      <CollapsibleCard title="Test">
        <p>Visible content</p>
      </CollapsibleCard>
    );
    expect(screen.getByText('Visible content')).toBeInTheDocument();
  });

  it('hides children when defaultOpen is false', () => {
    render(
      <CollapsibleCard title="Test" defaultOpen={false}>
        <p>Hidden content</p>
      </CollapsibleCard>
    );
    // Content is rendered but hidden via Collapse
    const content = screen.getByText('Hidden content');
    expect(content.closest('[style*="height: 0"]')).toBeTruthy();
  });

  it('toggles content visibility on click', async () => {
    const user = userEvent.setup();
    render(
      <CollapsibleCard title="Toggle Test">
        <p>Toggleable content</p>
      </CollapsibleCard>
    );

    const button = screen.getByRole('button', { name: /toggle test/i });

    // Initially open
    expect(screen.getByText('Toggleable content')).toBeVisible();

    // Click to close
    await user.click(button);

    // Click to open again
    await user.click(button);
    expect(screen.getByText('Toggleable content')).toBeVisible();
  });
});
