import type {
  StatusResponse,
  Action,
  Principal,
  Resource,
  Policy,
  Account,
  SimulationRequest,
  SimulationResponse,
  WhichPrincipalsRequest,
  WhichPrincipalsResponse,
  WhichResourcesRequest,
  WhichResourcesResponse,
  WhichActionsRequest,
  WhichActionsResponse,
  ApiError,
} from './types';

export class YamsApiError extends Error {
  status: number;

  constructor(message: string, status: number) {
    super(message);
    this.name = 'YamsApiError';
    this.status = status;
  }
}

export interface YamsClientConfig {
  baseUrl: string;
}

function getDefaultBaseUrl(): string {
  // Use empty string for relative URLs
  // In development, Vite's proxy handles /api/* -> localhost:8888
  // In production, the API is served from the same origin
  return '';
}

const DEFAULT_CONFIG: YamsClientConfig = {
  baseUrl: getDefaultBaseUrl(),
};

export class YamsClient {
  private baseUrl: string;

  constructor(config: Partial<YamsClientConfig> = {}) {
    const finalConfig = { ...DEFAULT_CONFIG, ...config };
    this.baseUrl = finalConfig.baseUrl.replace(/\/$/, '');
  }

  // Generic fetch wrapper with error handling
  private async fetch<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<T> {
    const url = `${this.baseUrl}/api/v1${endpoint}`;

    const response = await fetch(url, {
      ...options,
      headers: {
        'Content-Type': 'application/json',
        ...options.headers,
      },
    });

    if (!response.ok) {
      let message = `API error: ${response.statusText}`;
      try {
        const error = (await response.json()) as ApiError;
        message = error.message || message;
      } catch {
        // ignore JSON parse errors
      }
      throw new YamsApiError(message, response.status);
    }

    // Try to parse as JSON, fall back to text for non-JSON responses
    const text = await response.text();
    try {
      return JSON.parse(text) as T;
    } catch {
      return text as unknown as T;
    }
  }

  // Health & Status

  async healthcheck(): Promise<string> {
    return this.fetch<string>('/healthcheck');
  }

  async status(): Promise<StatusResponse> {
    return this.fetch<StatusResponse>('/status');
  }

  // Actions

  async listActions(): Promise<Action[]> {
    return this.fetch<Action[]>('/actions');
  }

  async getAction(key: string): Promise<Action> {
    return this.fetch<Action>(`/actions/${encodeURIComponent(key)}`);
  }

  async searchActions(query: string): Promise<Action[]> {
    return this.fetch<Action[]>(`/actions/search/${encodeURIComponent(query)}`);
  }

  // Principals

  async listPrincipals(): Promise<string[]> {
    return this.fetch<string[]>('/principals');
  }

  async getPrincipal(arn: string): Promise<Principal> {
    return this.fetch<Principal>(`/principals/${encodeURIComponent(arn)}`);
  }

  async searchPrincipals(query: string): Promise<string[]> {
    return this.fetch<string[]>(
      `/principals/search/${encodeURIComponent(query)}`
    );
  }

  // Resources

  async listResources(): Promise<string[]> {
    return this.fetch<string[]>('/resources');
  }

  async getResource(arn: string): Promise<Resource> {
    return this.fetch<Resource>(`/resources/${encodeURIComponent(arn)}`);
  }

  async searchResources(query: string): Promise<string[]> {
    return this.fetch<string[]>(
      `/resources/search/${encodeURIComponent(query)}`
    );
  }

  // Policies

  async listPolicies(): Promise<string[]> {
    return this.fetch<string[]>('/policies');
  }

  async getPolicy(arn: string): Promise<Policy> {
    return this.fetch<Policy>(`/policies/${encodeURIComponent(arn)}`);
  }

  async searchPolicies(query: string): Promise<string[]> {
    return this.fetch<string[]>(
      `/policies/search/${encodeURIComponent(query)}`
    );
  }

  // Accounts

  async listAccounts(): Promise<Account[]> {
    return this.fetch<Account[]>('/accounts');
  }

  async getAccount(key: string): Promise<Account> {
    return this.fetch<Account>(`/accounts/${encodeURIComponent(key)}`);
  }

  async searchAccounts(query: string): Promise<Account[]> {
    return this.fetch<Account[]>(
      `/accounts/search/${encodeURIComponent(query)}`
    );
  }

  async accountNames(): Promise<Record<string, string>> {
    return this.fetch<Record<string, string>>('/accounts/names');
  }

  // Simulation

  async simulate(request: SimulationRequest): Promise<SimulationResponse> {
    return this.fetch<SimulationResponse>('/sim', {
      method: 'POST',
      body: JSON.stringify(request),
    });
  }

  async whichPrincipals(
    request: WhichPrincipalsRequest
  ): Promise<WhichPrincipalsResponse> {
    return this.fetch<WhichPrincipalsResponse>('/sim/whichPrincipals', {
      method: 'POST',
      body: JSON.stringify(request),
    });
  }

  async whichResources(
    request: WhichResourcesRequest
  ): Promise<WhichResourcesResponse> {
    return this.fetch<WhichResourcesResponse>('/sim/whichResources', {
      method: 'POST',
      body: JSON.stringify(request),
    });
  }

  async whichActions(
    request: WhichActionsRequest
  ): Promise<WhichActionsResponse> {
    return this.fetch<WhichActionsResponse>('/sim/whichActions', {
      method: 'POST',
      body: JSON.stringify(request),
    });
  }
}
