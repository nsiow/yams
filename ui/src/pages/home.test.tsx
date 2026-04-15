// ui/src/pages/home.test.tsx
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, waitFor } from '../test/utils';
import userEvent from '@testing-library/user-event';
import { HomePage } from './home';
import { yamsApi } from '../lib/api';

// Mock the API
vi.mock('../lib/api', () => ({
  yamsApi: {
    healthcheck: vi.fn(),
    status: vi.fn(),
  },
}));

// Ensure tests clean up properly
afterEach(() => {
  vi.clearAllTimers();
});

const mockStatus = {
  accounts: 5,
  entities: 150,
  groups: 10,
  policies: 50,
  principals: 30,
  resources: 20,
  sources: [
    { source: 'aws-config', updated: new Date().toISOString() },
    { source: 'terraform', updated: new Date(Date.now() - 2 * 60 * 60 * 1000).toISOString() },
  ],
};

describe('HomePage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.resetAllMocks();
  });

  it('shows loading state initially', () => {
    vi.mocked(yamsApi.healthcheck).mockImplementation(() => new Promise(() => {}));
    vi.mocked(yamsApi.status).mockImplementation(() => new Promise(() => {}));

    render(<HomePage />);
    expect(screen.getByText(/connecting to yams server/i)).toBeInTheDocument();
  });

  it('displays dashboard when data loads successfully', async () => {
    vi.mocked(yamsApi.healthcheck).mockResolvedValue('ok');
    vi.mocked(yamsApi.status).mockResolvedValue(mockStatus);

    render(<HomePage />);

    await waitFor(() => {
      expect(screen.getByText('Dashboard')).toBeInTheDocument();
    });

    // Check entity counts are displayed
    expect(screen.getByText('Principals')).toBeInTheDocument();
    expect(screen.getByText('30')).toBeInTheDocument();
    expect(screen.getByText('Resources')).toBeInTheDocument();
    expect(screen.getByText('20')).toBeInTheDocument();
    expect(screen.getByText('Policies')).toBeInTheDocument();
    expect(screen.getByText('50')).toBeInTheDocument();
  });

  it('shows healthy status when healthcheck succeeds', async () => {
    vi.mocked(yamsApi.healthcheck).mockResolvedValue('ok');
    vi.mocked(yamsApi.status).mockResolvedValue(mockStatus);

    render(<HomePage />);

    await waitFor(() => {
      expect(screen.getByText('Healthy')).toBeInTheDocument();
    });
  });

  it('shows unhealthy status when healthcheck fails', async () => {
    vi.mocked(yamsApi.healthcheck).mockRejectedValue(new Error('Connection refused'));
    vi.mocked(yamsApi.status).mockResolvedValue(mockStatus);

    render(<HomePage />);

    await waitFor(() => {
      expect(screen.getByText('Unhealthy')).toBeInTheDocument();
    });
  });

  it('displays error when API fails', async () => {
    vi.mocked(yamsApi.healthcheck).mockRejectedValue(new Error('Network error'));
    vi.mocked(yamsApi.status).mockRejectedValue(new Error('Failed to connect'));

    render(<HomePage />);

    await waitFor(() => {
      expect(screen.getByText('Connection Error')).toBeInTheDocument();
      expect(screen.getByText('Failed to connect')).toBeInTheDocument();
    });
  });

  it('displays data sources with freshness indicators', async () => {
    vi.mocked(yamsApi.healthcheck).mockResolvedValue('ok');
    vi.mocked(yamsApi.status).mockResolvedValue(mockStatus);

    render(<HomePage />);

    await waitFor(() => {
      expect(screen.getByText('Data Sources')).toBeInTheDocument();
    });

    // Fresh source
    expect(screen.getByText('aws-config')).toBeInTheDocument();
    expect(screen.getByText('Fresh')).toBeInTheDocument();

    // Stale source (updated 2 hours ago)
    expect(screen.getByText('terraform')).toBeInTheDocument();
    expect(screen.getByText('Stale')).toBeInTheDocument();
  });

  it('refreshes data when refresh button is clicked', async () => {
    const user = userEvent.setup();
    vi.mocked(yamsApi.healthcheck).mockResolvedValue('ok');
    vi.mocked(yamsApi.status).mockResolvedValue(mockStatus);

    render(<HomePage />);

    await waitFor(() => {
      expect(screen.getByText('Dashboard')).toBeInTheDocument();
    });

    // Initial call
    expect(yamsApi.healthcheck).toHaveBeenCalledTimes(1);
    expect(yamsApi.status).toHaveBeenCalledTimes(1);

    // Click refresh
    await user.click(screen.getByRole('button', { name: /refresh dashboard/i }));

    // Should call APIs again
    await waitFor(() => {
      expect(yamsApi.healthcheck).toHaveBeenCalledTimes(2);
      expect(yamsApi.status).toHaveBeenCalledTimes(2);
    });
  });

  it('displays environment variables when present', async () => {
    const statusWithEnv = {
      ...mockStatus,
      env: { VERSION: '1.0.0', NODE_ENV: 'production' },
    };
    vi.mocked(yamsApi.healthcheck).mockResolvedValue('ok');
    vi.mocked(yamsApi.status).mockResolvedValue(statusWithEnv);

    render(<HomePage />);

    await waitFor(() => {
      expect(screen.getByText('Server Environment')).toBeInTheDocument();
    });

    expect(screen.getByText('VERSION:')).toBeInTheDocument();
    expect(screen.getByText('1.0.0')).toBeInTheDocument();
  });
});
