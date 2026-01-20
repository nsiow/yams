// ui/src/pages/simulate/which-resources.test.tsx
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, waitFor } from '../../test/utils';
import userEvent from '@testing-library/user-event';
import { WhichResourcesPage } from './which-resources';
import { yamsApi } from '../../lib/api';

// Mock the API
vi.mock('../../lib/api', () => ({
  yamsApi: {
    accountNames: vi.fn(),
    resourceAccounts: vi.fn(),
    actionAccessLevels: vi.fn(),
    searchPrincipals: vi.fn(),
    searchActions: vi.fn(),
    whichResources: vi.fn(),
    simulate: vi.fn(),
    listOverlays: vi.fn(),
    getOverlay: vi.fn(),
  },
}));

const mockAccountNames = {
  '123456789012': 'Production Account',
  '987654321098': 'Development Account',
};

const mockResourceAccounts = {
  'arn:aws:s3:::my-bucket': '123456789012',
  'arn:aws:s3:::other-bucket': '987654321098',
};

const mockActionAccessLevels = {
  's3:GetObject': 'Read',
  's3:PutObject': 'Write',
};

describe('WhichResourcesPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(yamsApi.accountNames).mockResolvedValue(mockAccountNames);
    vi.mocked(yamsApi.resourceAccounts).mockResolvedValue(mockResourceAccounts);
    vi.mocked(yamsApi.actionAccessLevels).mockResolvedValue(mockActionAccessLevels);
    vi.mocked(yamsApi.listOverlays).mockResolvedValue([]);
  });

  afterEach(() => {
    vi.resetAllMocks();
  });

  it('renders the page with title and description', async () => {
    render(<WhichResourcesPage />);

    expect(screen.getByText('Which Resources')).toBeInTheDocument();
    expect(screen.getByText(/Find which/)).toBeInTheDocument();
  });

  it('shows empty state when no selections made', async () => {
    render(<WhichResourcesPage />);

    await waitFor(() => {
      expect(screen.getByText(/Search and select a principal/)).toBeInTheDocument();
    });
  });

  it('shows principal and action search inputs', async () => {
    render(<WhichResourcesPage />);

    expect(screen.getByText('Principal (required)')).toBeInTheDocument();
    expect(screen.getByText('Action (optional)')).toBeInTheDocument();
  });

  it('searches for principals when typing in principal input', async () => {
    const user = userEvent.setup();
    vi.mocked(yamsApi.searchPrincipals).mockResolvedValue([
      'arn:aws:iam::123456789012:role/MyRole',
    ]);

    render(<WhichResourcesPage />);

    const principalInput = screen.getByPlaceholderText('Search principals...');
    await user.click(principalInput);
    await user.type(principalInput, 'MyRole');

    await waitFor(() => {
      expect(yamsApi.searchPrincipals).toHaveBeenCalledWith('MyRole');
    });
  });

  it('searches for actions when typing in action input', async () => {
    const user = userEvent.setup();
    vi.mocked(yamsApi.searchActions).mockResolvedValue(['s3:GetObject', 's3:PutObject']);

    render(<WhichResourcesPage />);

    const actionInput = screen.getByPlaceholderText('Search actions...');
    await user.click(actionInput);
    await user.type(actionInput, 's3:');

    await waitFor(() => {
      expect(yamsApi.searchActions).toHaveBeenCalledWith('s3:');
    });
  });

  it('runs query when principal is selected (action optional)', async () => {
    vi.mocked(yamsApi.whichResources).mockResolvedValue([
      'arn:aws:s3:::my-bucket',
      'arn:aws:s3:::other-bucket',
    ]);

    render(<WhichResourcesPage />, {
      routerProps: {
        initialEntries: ['/?principal=arn:aws:iam::123456789012:role/MyRole'],
      },
    });

    await waitFor(() => {
      expect(yamsApi.whichResources).toHaveBeenCalledWith(
        expect.objectContaining({
          principal: 'arn:aws:iam::123456789012:role/MyRole',
        })
      );
    });
  });

  it('runs query with action when both are selected', async () => {
    vi.mocked(yamsApi.whichResources).mockResolvedValue([
      'arn:aws:s3:::my-bucket',
    ]);

    render(<WhichResourcesPage />, {
      routerProps: {
        initialEntries: ['/?principal=arn:aws:iam::123456789012:role/MyRole&action=s3:GetObject'],
      },
    });

    await waitFor(() => {
      expect(yamsApi.whichResources).toHaveBeenCalledWith(
        expect.objectContaining({
          principal: 'arn:aws:iam::123456789012:role/MyRole',
          action: 's3:GetObject',
        })
      );
    });
  });

  it('displays results when query returns resources', async () => {
    vi.mocked(yamsApi.whichResources).mockResolvedValue([
      'arn:aws:s3:::my-bucket',
      'arn:aws:s3:::other-bucket',
    ]);

    render(<WhichResourcesPage />, {
      routerProps: {
        initialEntries: ['/?principal=arn:aws:iam::123456789012:role/MyRole'],
      },
    });

    await waitFor(() => {
      expect(screen.getByText('Results')).toBeInTheDocument();
    });

    expect(screen.getByText(/2 resources/)).toBeInTheDocument();
    expect(screen.getByText('my-bucket')).toBeInTheDocument();
    expect(screen.getByText('other-bucket')).toBeInTheDocument();
  });

  it('shows service and account columns', async () => {
    vi.mocked(yamsApi.whichResources).mockResolvedValue([
      'arn:aws:s3:::my-bucket',
    ]);

    render(<WhichResourcesPage />, {
      routerProps: {
        initialEntries: ['/?principal=arn:aws:iam::123456789012:role/MyRole'],
      },
    });

    await waitFor(() => {
      expect(screen.getByText('my-bucket')).toBeInTheDocument();
    });

    expect(screen.getByText('Service')).toBeInTheDocument();
    expect(screen.getByText('Account')).toBeInTheDocument();
    expect(screen.getByText('s3')).toBeInTheDocument();
    expect(screen.getByText('Production Account')).toBeInTheDocument();
  });

  it('shows no results message when query returns empty', async () => {
    vi.mocked(yamsApi.whichResources).mockResolvedValue([]);

    render(<WhichResourcesPage />, {
      routerProps: {
        initialEntries: ['/?principal=arn:aws:iam::123456789012:role/MyRole'],
      },
    });

    await waitFor(() => {
      expect(screen.getByText(/No resources found/)).toBeInTheDocument();
    });
  });

  it('displays error when query fails', async () => {
    vi.mocked(yamsApi.whichResources).mockRejectedValue(new Error('Query failed'));

    render(<WhichResourcesPage />, {
      routerProps: {
        initialEntries: ['/?principal=arn:aws:iam::123456789012:role/MyRole'],
      },
    });

    await waitFor(() => {
      expect(screen.getByText('Query failed')).toBeInTheDocument();
    });
  });

  it('filters results when using search input', async () => {
    const user = userEvent.setup();
    vi.mocked(yamsApi.whichResources).mockResolvedValue([
      'arn:aws:s3:::my-bucket',
      'arn:aws:s3:::other-bucket',
      'arn:aws:s3:::test-bucket',
    ]);

    render(<WhichResourcesPage />, {
      routerProps: {
        initialEntries: ['/?principal=arn:aws:iam::123456789012:role/MyRole'],
      },
    });

    await waitFor(() => {
      expect(screen.getByText('Results')).toBeInTheDocument();
    });

    expect(screen.getByText(/3 resources/)).toBeInTheDocument();

    const filterInput = screen.getByPlaceholderText('Filter results...');
    await user.type(filterInput, 'my-bucket');

    await waitFor(() => {
      expect(screen.getByText(/1 resource/)).toBeInTheDocument();
    });

    expect(screen.getByText('my-bucket')).toBeInTheDocument();
    expect(screen.queryByText('other-bucket')).not.toBeInTheDocument();
  });

  it('shows Clear All button when selections are made', async () => {
    vi.mocked(yamsApi.whichResources).mockResolvedValue([]);

    render(<WhichResourcesPage />, {
      routerProps: {
        initialEntries: ['/?principal=arn:aws:iam::123456789012:role/MyRole'],
      },
    });

    await waitFor(() => {
      expect(screen.getByText('Clear All')).toBeInTheDocument();
    });
  });

  it('has Go to Simulation links for each result', async () => {
    vi.mocked(yamsApi.whichResources).mockResolvedValue(['arn:aws:s3:::my-bucket']);

    render(<WhichResourcesPage />, {
      routerProps: {
        initialEntries: ['/?principal=arn:aws:iam::123456789012:role/MyRole'],
      },
    });

    await waitFor(() => {
      expect(screen.getByText('my-bucket')).toBeInTheDocument();
    });

    expect(screen.getByText('Go to Simulation')).toBeInTheDocument();
    expect(screen.getByText('Open')).toBeInTheDocument();
  });
});
