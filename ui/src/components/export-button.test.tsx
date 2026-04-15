// ui/src/components/export-button.test.tsx
import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '../test/utils';
import userEvent from '@testing-library/user-event';
import { ExportButton } from './export-button';

describe('ExportButton', () => {
  it('renders download button', () => {
    render(<ExportButton data={{ test: 'value' }} filename="test" />);
    expect(screen.getByRole('button')).toBeInTheDocument();
  });

  it('creates downloadable blob on click', async () => {
    const user = userEvent.setup();
    const testData = { key: 'value', nested: { prop: 123 } };

    // Mock the DOM methods needed for download
    const clickMock = vi.fn();
    const originalCreateElement = document.createElement.bind(document);
    vi.spyOn(document, 'createElement').mockImplementation((tag: string) => {
      const element = originalCreateElement(tag);
      if (tag === 'a') {
        element.click = clickMock;
      }
      return element;
    });

    render(<ExportButton data={testData} filename="export-test" />);
    await user.click(screen.getByRole('button'));

    expect(clickMock).toHaveBeenCalled();
    expect(URL.createObjectURL).toHaveBeenCalled();

    vi.restoreAllMocks();
  });
});
