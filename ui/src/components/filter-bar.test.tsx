// ui/src/components/filter-bar.test.tsx
import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '../test/utils';
import userEvent from '@testing-library/user-event';
import { FilterBar } from './filter-bar';

describe('FilterBar', () => {
  it('renders children', () => {
    render(
      <FilterBar hasActiveFilters={false} onClearAll={() => {}}>
        <input placeholder="Search" />
      </FilterBar>
    );
    expect(screen.getByPlaceholderText('Search')).toBeInTheDocument();
  });

  it('shows Clear button when hasActiveFilters is true', () => {
    render(
      <FilterBar hasActiveFilters={true} onClearAll={() => {}}>
        <input />
      </FilterBar>
    );
    expect(screen.getByRole('button', { name: /clear/i })).toBeInTheDocument();
  });

  it('hides Clear button when hasActiveFilters is false', () => {
    render(
      <FilterBar hasActiveFilters={false} onClearAll={() => {}}>
        <input />
      </FilterBar>
    );
    expect(screen.queryByRole('button', { name: /clear/i })).not.toBeInTheDocument();
  });

  it('calls onClearAll when Clear button is clicked', async () => {
    const user = userEvent.setup();
    const onClearAll = vi.fn();

    render(
      <FilterBar hasActiveFilters={true} onClearAll={onClearAll}>
        <input />
      </FilterBar>
    );

    await user.click(screen.getByRole('button', { name: /clear/i }));
    expect(onClearAll).toHaveBeenCalledTimes(1);
  });
});
