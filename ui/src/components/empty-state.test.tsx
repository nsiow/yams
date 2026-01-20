// ui/src/components/empty-state.test.tsx
import { describe, it, expect } from 'vitest';
import { render, screen } from '../test/utils';
import { EmptyState } from './empty-state';

describe('EmptyState', () => {
  it('renders no-results variant with default message', () => {
    render(<EmptyState variant="no-results" />);
    expect(screen.getByText(/no results match your filters/i)).toBeInTheDocument();
  });

  it('renders no-selection variant with default message', () => {
    render(<EmptyState variant="no-selection" />);
    expect(screen.getByText(/select an item to view details/i)).toBeInTheDocument();
  });

  it('renders no-selection variant with entity name', () => {
    render(<EmptyState variant="no-selection" entityName="principal" />);
    expect(screen.getByText(/select a principal to view details/i)).toBeInTheDocument();
  });

  it('renders error variant with default message', () => {
    render(<EmptyState variant="error" />);
    expect(screen.getByText(/failed to load data/i)).toBeInTheDocument();
  });

  it('renders no-data variant with default message', () => {
    render(<EmptyState variant="no-data" />);
    expect(screen.getByText(/no data available/i)).toBeInTheDocument();
  });

  it('renders custom message when provided', () => {
    render(<EmptyState variant="error" message="Custom error message" />);
    expect(screen.getByText('Custom error message')).toBeInTheDocument();
  });
});
