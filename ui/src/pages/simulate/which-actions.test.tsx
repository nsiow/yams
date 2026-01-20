// ui/src/pages/simulate/which-actions.test.tsx
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, waitFor } from '../../test/utils';
import userEvent from '@testing-library/user-event';
import { WhichActionsPage } from './which-actions';
import { yamsApi } from '../../lib/api';

// Mock the API
vi.mock('../../lib/api', () => ({
  yamsApi: {
    accountNames: vi.fn(),
    resourceAccounts: vi.fn(),
    actionAccessLevels: vi.fn(),
    searchPrincipals: vi.fn(),
    searchResources: vi.fn(),
    whichActions: vi.fn(),
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
};

const mockActionAccessLevels = {
  's3:GetObject': 'Read',
  's3:PutObject': 'Write',
  's3:DeleteObject': 'Write',
  's3:ListBucket': 'List',
};

describe('WhichActionsPage', () => {
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
    render(<WhichActionsPage />);

    expect(screen.getByText('Which Actions')).toBeInTheDocument();
    expect(screen.getByText(/Find which/)).toBeInTheDocument();
  });

  it('shows empty state when no selections made', async () => {
    render(<WhichActionsPage />);

    await waitFor(() => {
      expect(screen.getByText(/Search and select a principal and resource/)).toBeInTheDocument();
    });
  });

  it('shows principal and resource search inputs', async () => {
    render(<WhichActionsPage />);

    expect(screen.getByText('Principal (required)')).toBeInTheDocument();
    expect(screen.getByText('Resource (required)')).toBeInTheDocument();
  });

  it('searches for principals when typing in principal input', async () => {
    const user = userEvent.setup();
    vi.mocked(yamsApi.searchPrincipals).mockResolvedValue([
      'arn:aws:iam::123456789012:role/MyRole',
    ]);

    render(<WhichActionsPage />);

    const principalInput = screen.getByPlaceholderText('Search principals...');
    await user.click(principalInput);
    await user.type(principalInput, 'MyRole');

    await waitFor(() => {
      expect(yamsApi.searchPrincipals).toHaveBeenCalledWith('MyRole');
    });
  });

  it('searches for resources when typing in resource input', async () => {
    const user = userEvent.setup();
    vi.mocked(yamsApi.searchResources).mockResolvedValue(['arn:aws:s3:::my-bucket']);

    render(<WhichActionsPage />);

    const resourceInput = screen.getByPlaceholderText('Search resources...');
    await user.click(resourceInput);
    await user.type(resourceInput, 'bucket');

    await waitFor(() => {
      expect(yamsApi.searchResources).toHaveBeenCalledWith('bucket');
    });
  });

  it('runs query when both principal and resource are selected', async () => {
    vi.mocked(yamsApi.whichActions).mockResolvedValue([
      's3:GetObject',
      's3:PutObject',
    ]);

    render(<WhichActionsPage />, {
      routerProps: {
        initialEntries: ['/?principal=arn:aws:iam::123456789012:role/MyRole&resource=arn:aws:s3:::my-bucket'],
      },
    });

    await waitFor(() => {
      expect(yamsApi.whichActions).toHaveBeenCalledWith(
        expect.objectContaining({
          principal: 'arn:aws:iam::123456789012:role/MyRole',
          resource: 'arn:aws:s3:::my-bucket',
        })
      );
    });
  });

  it('displays results when query returns actions', async () => {
    vi.mocked(yamsApi.whichActions).mockResolvedValue([
      's3:GetObject',
      's3:PutObject',
      's3:ListBucket',
    ]);

    render(<WhichActionsPage />, {
      routerProps: {
        initialEntries: ['/?principal=arn:aws:iam::123456789012:role/MyRole&resource=arn:aws:s3:::my-bucket'],
      },
    });

    await waitFor(() => {
      expect(screen.getByText('Results')).toBeInTheDocument();
    });

    expect(screen.getByText(/3 actions/)).toBeInTheDocument();
    expect(screen.getByText('s3:GetObject')).toBeInTheDocument();
    expect(screen.getByText('s3:PutObject')).toBeInTheDocument();
    expect(screen.getByText('s3:ListBucket')).toBeInTheDocument();
  });

  it('shows access level badges for actions', async () => {
    vi.mocked(yamsApi.whichActions).mockResolvedValue([
      's3:GetObject',
      's3:PutObject',
      's3:ListBucket',
    ]);

    render(<WhichActionsPage />, {
      routerProps: {
        initialEntries: ['/?principal=arn:aws:iam::123456789012:role/MyRole&resource=arn:aws:s3:::my-bucket'],
      },
    });

    await waitFor(() => {
      expect(screen.getByText('s3:GetObject')).toBeInTheDocument();
    });

    expect(screen.getByText('Read')).toBeInTheDocument();
    expect(screen.getAllByText('Write')).toHaveLength(1);
    expect(screen.getByText('List')).toBeInTheDocument();
  });

  it('shows no results message when query returns empty', async () => {
    vi.mocked(yamsApi.whichActions).mockResolvedValue([]);

    render(<WhichActionsPage />, {
      routerProps: {
        initialEntries: ['/?principal=arn:aws:iam::123456789012:role/MyRole&resource=arn:aws:s3:::my-bucket'],
      },
    });

    await waitFor(() => {
      expect(screen.getByText(/No actions found/)).toBeInTheDocument();
    });
  });

  it('displays error when query fails', async () => {
    vi.mocked(yamsApi.whichActions).mockRejectedValue(new Error('Query failed'));

    render(<WhichActionsPage />, {
      routerProps: {
        initialEntries: ['/?principal=arn:aws:iam::123456789012:role/MyRole&resource=arn:aws:s3:::my-bucket'],
      },
    });

    await waitFor(() => {
      expect(screen.getByText('Query failed')).toBeInTheDocument();
    });
  });

  it('filters results when using search input', async () => {
    const user = userEvent.setup();
    vi.mocked(yamsApi.whichActions).mockResolvedValue([
      's3:GetObject',
      's3:PutObject',
      's3:DeleteObject',
    ]);

    render(<WhichActionsPage />, {
      routerProps: {
        initialEntries: ['/?principal=arn:aws:iam::123456789012:role/MyRole&resource=arn:aws:s3:::my-bucket'],
      },
    });

    await waitFor(() => {
      expect(screen.getByText('Results')).toBeInTheDocument();
    });

    expect(screen.getByText(/3 actions/)).toBeInTheDocument();

    const filterInput = screen.getByPlaceholderText('Filter results...');
    await user.type(filterInput, 'Get');

    await waitFor(() => {
      expect(screen.getByText(/1 action/)).toBeInTheDocument();
    });

    expect(screen.getByText('s3:GetObject')).toBeInTheDocument();
    expect(screen.queryByText('s3:PutObject')).not.toBeInTheDocument();
  });

  it('shows Clear All button when selections are made', async () => {
    render(<WhichActionsPage />, {
      routerProps: {
        initialEntries: ['/?principal=arn:aws:iam::123456789012:role/MyRole'],
      },
    });

    await waitFor(() => {
      expect(screen.getByText('Clear All')).toBeInTheDocument();
    });
  });

  it('has Go to Simulation links for each result', async () => {
    vi.mocked(yamsApi.whichActions).mockResolvedValue(['s3:GetObject']);

    render(<WhichActionsPage />, {
      routerProps: {
        initialEntries: ['/?principal=arn:aws:iam::123456789012:role/MyRole&resource=arn:aws:s3:::my-bucket'],
      },
    });

    await waitFor(() => {
      expect(screen.getByText('s3:GetObject')).toBeInTheDocument();
    });

    expect(screen.getByText('Go to Simulation')).toBeInTheDocument();
    expect(screen.getByText('Open')).toBeInTheDocument();
  });
});
