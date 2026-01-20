// ui/src/pages/overlays/editor.test.tsx
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { OverlayEditorPage } from './editor';
import { yamsApi } from '../../lib/api';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { MantineProvider } from '@mantine/core';

// Mock the API
vi.mock('../../lib/api', () => ({
  yamsApi: {
    listOverlays: vi.fn(),
    getOverlay: vi.fn(),
    createOverlay: vi.fn(),
    updateOverlay: vi.fn(),
    accountNames: vi.fn(),
    resourceAccounts: vi.fn(),
    listPrincipals: vi.fn(),
    searchPrincipals: vi.fn(),
    getPrincipal: vi.fn(),
  },
}));

const mockOverlaySummaries = [
  {
    id: 'overlay-1',
    name: 'Test Overlay',
    createdAt: new Date().toISOString(),
    numPrincipals: 2,
    numResources: 1,
    numPolicies: 0,
    numAccounts: 0,
    numGroups: 0,
  },
];

const mockOverlayData = {
  id: 'overlay-1',
  name: 'Test Overlay',
  createdAt: new Date().toISOString(),
  principals: [
    {
      Type: 'Role',
      AccountId: '123456789012',
      Name: 'TestRole',
      Arn: 'arn:aws:iam::123456789012:role/TestRole',
    },
  ],
  resources: [],
  policies: [],
  accounts: [],
  groups: [],
};

function renderWithRouter(id: string): ReturnType<typeof render> {
  return render(
    <MemoryRouter initialEntries={[`/overlays/${id}/edit`]}>
      <MantineProvider>
        <Routes>
          <Route path="/overlays/:id/edit" element={<OverlayEditorPage />} />
        </Routes>
      </MantineProvider>
    </MemoryRouter>
  );
}

describe('OverlayEditorPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(yamsApi.listOverlays).mockResolvedValue(mockOverlaySummaries);
    vi.mocked(yamsApi.accountNames).mockResolvedValue({ '123456789012': 'TestAccount' });
    vi.mocked(yamsApi.resourceAccounts).mockResolvedValue({});
  });

  afterEach(() => {
    vi.resetAllMocks();
  });

  describe('New Overlay Creation', () => {
    it('shows new overlay form when id is "new"', async () => {
      renderWithRouter('new');

      await waitFor(() => {
        expect(screen.getByText('Create New Overlay')).toBeInTheDocument();
      });

      // Create button should be disabled without a name
      const createButton = screen.getByRole('button', { name: /create overlay/i });
      expect(createButton).toBeDisabled();
    });

    it('enables create button when name is confirmed', async () => {
      const user = userEvent.setup();
      renderWithRouter('new');

      await waitFor(() => {
        expect(screen.getByPlaceholderText(/type a name and press enter/i)).toBeInTheDocument();
      });

      // Type name and press Enter
      const nameInput = screen.getByPlaceholderText(/type a name and press enter/i);
      await user.type(nameInput, 'My New Overlay{enter}');

      // Create button should now be enabled
      const createButton = screen.getByRole('button', { name: /create overlay/i });
      expect(createButton).not.toBeDisabled();
    });

    it('creates overlay when form is submitted', async () => {
      const user = userEvent.setup();
      vi.mocked(yamsApi.createOverlay).mockResolvedValue({
        id: 'new-overlay-id',
        name: 'My New Overlay',
        createdAt: new Date().toISOString(),
      });

      renderWithRouter('new');

      await waitFor(() => {
        expect(screen.getByPlaceholderText(/type a name and press enter/i)).toBeInTheDocument();
      });

      // Confirm name
      const nameInput = screen.getByPlaceholderText(/type a name and press enter/i);
      await user.type(nameInput, 'My New Overlay{enter}');

      // Click create
      await user.click(screen.getByRole('button', { name: /create overlay/i }));

      await waitFor(() => {
        expect(yamsApi.createOverlay).toHaveBeenCalledWith(
          expect.objectContaining({ name: 'My New Overlay' })
        );
      });
    });
  });

  describe('Existing Overlay Editing', () => {
    it('loads existing overlay data', async () => {
      vi.mocked(yamsApi.getOverlay).mockResolvedValue(mockOverlayData);
      renderWithRouter('overlay-1');

      await waitFor(() => {
        expect(yamsApi.getOverlay).toHaveBeenCalledWith('overlay-1');
      });

      // Should show the overlay name in breadcrumbs
      await waitFor(() => {
        expect(screen.getByText('Test Overlay')).toBeInTheDocument();
      });
    });

    it('displays entity tabs with counts', async () => {
      vi.mocked(yamsApi.getOverlay).mockResolvedValue(mockOverlayData);
      renderWithRouter('overlay-1');

      await waitFor(() => {
        expect(screen.getByRole('tab', { name: /principals \(1\)/i })).toBeInTheDocument();
        expect(screen.getByRole('tab', { name: /resources \(0\)/i })).toBeInTheDocument();
        expect(screen.getByRole('tab', { name: /policies \(0\)/i })).toBeInTheDocument();
        expect(screen.getByRole('tab', { name: /accounts \(0\)/i })).toBeInTheDocument();
      });
    });

    it('shows principal in the list', async () => {
      vi.mocked(yamsApi.getOverlay).mockResolvedValue(mockOverlayData);
      renderWithRouter('overlay-1');

      await waitFor(() => {
        expect(screen.getByText('TestRole')).toBeInTheDocument();
      });
    });

    it('shows save button disabled when no changes', async () => {
      vi.mocked(yamsApi.getOverlay).mockResolvedValue(mockOverlayData);
      renderWithRouter('overlay-1');

      await waitFor(() => {
        const saveButton = screen.getByRole('button', { name: /save changes/i });
        expect(saveButton).toBeDisabled();
      });
    });
  });

  describe('Error Handling', () => {
    it('displays error when overlay fails to load', async () => {
      vi.mocked(yamsApi.getOverlay).mockRejectedValue(new Error('Overlay not found'));
      renderWithRouter('nonexistent');

      await waitFor(() => {
        expect(screen.getByText('Overlay not found')).toBeInTheDocument();
      });
    });
  });

  describe('Entity Selection', () => {
    it('shows detail panel when entity is selected', async () => {
      const user = userEvent.setup();
      vi.mocked(yamsApi.getOverlay).mockResolvedValue(mockOverlayData);
      renderWithRouter('overlay-1');

      await waitFor(() => {
        expect(screen.getByText('TestRole')).toBeInTheDocument();
      });

      // Click on the principal
      await user.click(screen.getByText('TestRole'));

      // Detail panel should show entity info
      await waitFor(() => {
        expect(screen.getByDisplayValue('arn:aws:iam::123456789012:role/TestRole')).toBeInTheDocument();
      });
    });
  });
});
