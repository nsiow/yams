// ui/src/pages/simulate/which-principals.test.tsx
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, waitFor } from '../../test/utils';
import userEvent from '@testing-library/user-event';
import { WhichPrincipalsPage } from './which-principals';
import { yamsApi } from '../../lib/api';

// Mock the API
vi.mock('../../lib/api', () => ({
  yamsApi: {
    accountNames: vi.fn(),
    resourceAccounts: vi.fn(),
    actionAccessLevels: vi.fn(),
    searchActions: vi.fn(),
    searchResources: vi.fn(),
    whichPrincipals: vi.fn(),
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
};

describe('WhichPrincipalsPage', () => {
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
    render(<WhichPrincipalsPage />);

    expect(screen.getByText('Which Principals')).toBeInTheDocument();
    expect(screen.getByText(/Find which/)).toBeInTheDocument();
  });

  it('shows empty state when no selections made', async () => {
    render(<WhichPrincipalsPage />);

    await waitFor(() => {
      expect(screen.getByText(/Search and select an action and resource/)).toBeInTheDocument();
    });
  });

  it('shows action and resource search inputs', async () => {
    render(<WhichPrincipalsPage />);

    expect(screen.getByText('Action (required)')).toBeInTheDocument();
    expect(screen.getByText('Resource (required)')).toBeInTheDocument();
  });

  it('searches for actions when typing in action input', async () => {
    const user = userEvent.setup();
    vi.mocked(yamsApi.searchActions).mockResolvedValue(['s3:GetObject', 's3:PutObject']);

    render(<WhichPrincipalsPage />);

    const actionInput = screen.getByPlaceholderText('Search actions...');
    await user.click(actionInput);
    await user.type(actionInput, 's3:');

    await waitFor(() => {
      expect(yamsApi.searchActions).toHaveBeenCalledWith('s3:');
    });
  });

  it('searches for resources when typing in resource input', async () => {
    const user = userEvent.setup();
    vi.mocked(yamsApi.searchResources).mockResolvedValue(['arn:aws:s3:::my-bucket']);

    render(<WhichPrincipalsPage />);

    const resourceInput = screen.getByPlaceholderText('Search resources...');
    await user.click(resourceInput);
    await user.type(resourceInput, 'bucket');

    await waitFor(() => {
      expect(yamsApi.searchResources).toHaveBeenCalledWith('bucket', undefined);
    });
  });

  it('runs query when both action and resource are selected', async () => {
    vi.mocked(yamsApi.searchActions).mockResolvedValue(['s3:GetObject']);
    vi.mocked(yamsApi.searchResources).mockResolvedValue(['arn:aws:s3:::my-bucket']);
    vi.mocked(yamsApi.whichPrincipals).mockResolvedValue([
      'arn:aws:iam::123456789012:role/MyRole',
      'arn:aws:iam::123456789012:user/MyUser',
    ]);

    render(<WhichPrincipalsPage />, {
      routerProps: {
        initialEntries: ['/?action=s3:GetObject&resource=arn:aws:s3:::my-bucket'],
      },
    });

    await waitFor(() => {
      expect(yamsApi.whichPrincipals).toHaveBeenCalledWith(
        expect.objectContaining({
          action: 's3:GetObject',
          resource: 'arn:aws:s3:::my-bucket',
        })
      );
    });
  });

  it('displays results when query returns principals', async () => {
    vi.mocked(yamsApi.whichPrincipals).mockResolvedValue([
      'arn:aws:iam::123456789012:role/MyRole',
      'arn:aws:iam::123456789012:user/MyUser',
    ]);

    render(<WhichPrincipalsPage />, {
      routerProps: {
        initialEntries: ['/?action=s3:GetObject&resource=arn:aws:s3:::my-bucket'],
      },
    });

    await waitFor(() => {
      expect(screen.getByText('Results')).toBeInTheDocument();
    });

    expect(screen.getByText(/2 principals/)).toBeInTheDocument();
    expect(screen.getByText('MyRole')).toBeInTheDocument();
    expect(screen.getByText('MyUser')).toBeInTheDocument();
  });

  it('shows no results message when query returns empty', async () => {
    vi.mocked(yamsApi.whichPrincipals).mockResolvedValue([]);

    render(<WhichPrincipalsPage />, {
      routerProps: {
        initialEntries: ['/?action=s3:GetObject&resource=arn:aws:s3:::my-bucket'],
      },
    });

    await waitFor(() => {
      expect(screen.getByText(/No principals found/)).toBeInTheDocument();
    });
  });

  it('displays error when query fails', async () => {
    vi.mocked(yamsApi.whichPrincipals).mockRejectedValue(new Error('Query failed'));

    render(<WhichPrincipalsPage />, {
      routerProps: {
        initialEntries: ['/?action=s3:GetObject&resource=arn:aws:s3:::my-bucket'],
      },
    });

    await waitFor(() => {
      expect(screen.getByText('Query failed')).toBeInTheDocument();
    });
  });

  it('filters results when using search input', async () => {
    const user = userEvent.setup();
    vi.mocked(yamsApi.whichPrincipals).mockResolvedValue([
      'arn:aws:iam::123456789012:role/AdminRole',
      'arn:aws:iam::123456789012:role/ReadOnlyRole',
      'arn:aws:iam::123456789012:user/TestUser',
    ]);

    render(<WhichPrincipalsPage />, {
      routerProps: {
        initialEntries: ['/?action=s3:GetObject&resource=arn:aws:s3:::my-bucket'],
      },
    });

    await waitFor(() => {
      expect(screen.getByText('Results')).toBeInTheDocument();
    });

    expect(screen.getByText(/3 principals/)).toBeInTheDocument();

    const filterInput = screen.getByPlaceholderText('Filter results...');
    await user.type(filterInput, 'Admin');

    await waitFor(() => {
      expect(screen.getByText(/1 principal/)).toBeInTheDocument();
    });

    expect(screen.getByText('AdminRole')).toBeInTheDocument();
    expect(screen.queryByText('ReadOnlyRole')).not.toBeInTheDocument();
    expect(screen.queryByText('TestUser')).not.toBeInTheDocument();
  });

  it('shows Clear All button when selections are made', async () => {
    render(<WhichPrincipalsPage />, {
      routerProps: {
        initialEntries: ['/?action=s3:GetObject'],
      },
    });

    await waitFor(() => {
      expect(screen.getByText('Clear All')).toBeInTheDocument();
    });
  });

  it('shows account name for principals when available', async () => {
    vi.mocked(yamsApi.whichPrincipals).mockResolvedValue([
      'arn:aws:iam::123456789012:role/MyRole',
    ]);

    render(<WhichPrincipalsPage />, {
      routerProps: {
        initialEntries: ['/?action=s3:GetObject&resource=arn:aws:s3:::my-bucket'],
      },
    });

    await waitFor(() => {
      expect(screen.getByText('MyRole')).toBeInTheDocument();
    });

    expect(screen.getByText('Production Account (123456789012)')).toBeInTheDocument();
  });

  it('has Go to Simulation links for each result', async () => {
    vi.mocked(yamsApi.whichPrincipals).mockResolvedValue([
      'arn:aws:iam::123456789012:role/MyRole',
    ]);

    render(<WhichPrincipalsPage />, {
      routerProps: {
        initialEntries: ['/?action=s3:GetObject&resource=arn:aws:s3:::my-bucket'],
      },
    });

    await waitFor(() => {
      expect(screen.getByText('MyRole')).toBeInTheDocument();
    });

    expect(screen.getByText('Go to Simulation')).toBeInTheDocument();
    expect(screen.getByText('Open')).toBeInTheDocument();
  });
});
