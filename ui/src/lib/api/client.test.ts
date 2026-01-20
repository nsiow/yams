// ui/src/lib/api/client.test.ts
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { YamsClient, YamsApiError } from './client';

describe('YamsClient', () => {
  let client: YamsClient;
  let fetchSpy: ReturnType<typeof vi.spyOn>;

  beforeEach(() => {
    client = new YamsClient({ baseUrl: 'http://localhost:8888' });
    fetchSpy = vi.spyOn(global, 'fetch');
  });

  afterEach(() => {
    fetchSpy.mockRestore();
  });

  // Helper to mock a successful JSON response
  function mockJsonResponse<T>(data: T): void {
    fetchSpy.mockResolvedValueOnce({
      ok: true,
      text: () => Promise.resolve(JSON.stringify(data)),
    } as Response);
  }

  // Helper to mock an error response
  function mockErrorResponse(status: number, message: string): void {
    fetchSpy.mockResolvedValueOnce({
      ok: false,
      status,
      statusText: 'Error',
      json: () => Promise.resolve({ message }),
    } as Response);
  }

  describe('healthcheck', () => {
    it('returns health status', async () => {
      mockJsonResponse('ok');
      const result = await client.healthcheck();
      expect(result).toBe('ok');
      expect(fetchSpy).toHaveBeenCalledWith(
        'http://localhost:8888/api/v1/healthcheck',
        expect.objectContaining({ headers: { 'Content-Type': 'application/json' } })
      );
    });
  });

  describe('status', () => {
    it('returns status response', async () => {
      const statusData = {
        accounts: 5,
        entities: 100,
        groups: 10,
        policies: 50,
        principals: 30,
        resources: 20,
        sources: [{ source: 'test', updated: '2024-01-01' }],
      };
      mockJsonResponse(statusData);

      const result = await client.status();
      expect(result).toEqual(statusData);
    });
  });

  describe('actions', () => {
    it('lists all actions', async () => {
      const actions = ['s3:GetObject', 's3:PutObject'];
      mockJsonResponse(actions);

      const result = await client.listActions();
      expect(result).toEqual(actions);
    });

    it('gets single action', async () => {
      const action = { Name: 's3:GetObject', Service: 's3', AccessLevel: 'Read' };
      mockJsonResponse(action);

      const result = await client.getAction('s3:GetObject');
      expect(result).toEqual(action);
      expect(fetchSpy).toHaveBeenCalledWith(
        'http://localhost:8888/api/v1/actions/s3%3AGetObject',
        expect.any(Object)
      );
    });

    it('searches actions', async () => {
      const actions = ['s3:GetObject'];
      mockJsonResponse(actions);

      const result = await client.searchActions('s3:Get*');
      expect(result).toEqual(actions);
    });
  });

  describe('principals', () => {
    it('lists all principals', async () => {
      const principals = ['arn:aws:iam::123:role/Test'];
      mockJsonResponse(principals);

      const result = await client.listPrincipals();
      expect(result).toEqual(principals);
    });

    it('gets single principal', async () => {
      const principal = {
        Type: 'Role',
        AccountId: '123',
        Name: 'TestRole',
        Arn: 'arn:aws:iam::123:role/Test',
      };
      mockJsonResponse(principal);

      const result = await client.getPrincipal('arn:aws:iam::123:role/Test');
      expect(result).toEqual(principal);
    });
  });

  describe('resources', () => {
    it('lists all resources', async () => {
      const resources = ['arn:aws:s3:::bucket'];
      mockJsonResponse(resources);

      const result = await client.listResources();
      expect(result).toEqual(resources);
    });

    it('gets single resource', async () => {
      const resource = {
        Type: 'bucket',
        AccountId: '123',
        Region: 'us-east-1',
        Name: 'test-bucket',
        Arn: 'arn:aws:s3:::bucket',
      };
      mockJsonResponse(resource);

      const result = await client.getResource('arn:aws:s3:::bucket');
      expect(result).toEqual(resource);
    });
  });

  describe('policies', () => {
    it('lists all policies', async () => {
      const policies = ['arn:aws:iam::123:policy/Test'];
      mockJsonResponse(policies);

      const result = await client.listPolicies();
      expect(result).toEqual(policies);
    });

    it('gets single policy', async () => {
      const policy = {
        Type: 'ManagedPolicy',
        AccountId: '123',
        Arn: 'arn:aws:iam::123:policy/Test',
        Name: 'TestPolicy',
        Policy: { Version: '2012-10-17', Statement: [] },
      };
      mockJsonResponse(policy);

      const result = await client.getPolicy('arn:aws:iam::123:policy/Test');
      expect(result).toEqual(policy);
    });
  });

  describe('accounts', () => {
    it('lists all accounts', async () => {
      const accounts = ['123456789012'];
      mockJsonResponse(accounts);

      const result = await client.listAccounts();
      expect(result).toEqual(accounts);
    });

    it('gets single account', async () => {
      const account = {
        Id: '123456789012',
        Name: 'TestAccount',
        OrgId: 'o-abc123',
      };
      mockJsonResponse(account);

      const result = await client.getAccount('123456789012');
      expect(result).toEqual(account);
    });
  });

  describe('simulation', () => {
    it('runs simulation', async () => {
      const response = {
        result: 'ALLOW' as const,
        principal: 'arn:aws:iam::123:role/Test',
        action: 's3:GetObject',
        resource: 'arn:aws:s3:::bucket/*',
      };
      mockJsonResponse(response);

      const result = await client.simulate({
        principal: 'arn:aws:iam::123:role/Test',
        action: 's3:GetObject',
        resource: 'arn:aws:s3:::bucket/*',
      });

      expect(result).toEqual(response);
      expect(fetchSpy).toHaveBeenCalledWith(
        'http://localhost:8888/api/v1/sim',
        expect.objectContaining({ method: 'POST' })
      );
    });

    it('runs whichPrincipals query', async () => {
      const response = { principals: ['arn:aws:iam::123:role/Test'] };
      mockJsonResponse(response);

      const result = await client.whichPrincipals({
        action: 's3:GetObject',
        resource: 'arn:aws:s3:::bucket/*',
      });

      expect(result).toEqual(response);
    });

    it('runs whichResources query', async () => {
      const response = { resources: ['arn:aws:s3:::bucket/*'] };
      mockJsonResponse(response);

      const result = await client.whichResources({
        principal: 'arn:aws:iam::123:role/Test',
        action: 's3:GetObject',
      });

      expect(result).toEqual(response);
    });

    it('runs whichActions query', async () => {
      const response = { actions: ['s3:GetObject', 's3:PutObject'] };
      mockJsonResponse(response);

      const result = await client.whichActions({
        principal: 'arn:aws:iam::123:role/Test',
        resource: 'arn:aws:s3:::bucket/*',
      });

      expect(result).toEqual(response);
    });
  });

  describe('overlays', () => {
    it('lists overlays', async () => {
      const overlays = [
        { id: '1', name: 'test', createdAt: '2024-01-01', numPrincipals: 1, numResources: 0, numPolicies: 0, numAccounts: 0, numGroups: 0 },
      ];
      mockJsonResponse(overlays);

      const result = await client.listOverlays();
      expect(result).toEqual(overlays);
    });

    it('lists overlays with query', async () => {
      mockJsonResponse([]);

      await client.listOverlays('search');
      expect(fetchSpy).toHaveBeenCalledWith(
        'http://localhost:8888/api/v1/overlays?q=search',
        expect.any(Object)
      );
    });

    it('gets single overlay', async () => {
      const overlay = { id: '1', name: 'test', createdAt: '2024-01-01' };
      mockJsonResponse(overlay);

      const result = await client.getOverlay('1');
      expect(result).toEqual(overlay);
    });

    it('creates overlay', async () => {
      const overlay = { id: '1', name: 'new', createdAt: '2024-01-01' };
      mockJsonResponse(overlay);

      const result = await client.createOverlay({ name: 'new' });
      expect(result).toEqual(overlay);
      expect(fetchSpy).toHaveBeenCalledWith(
        'http://localhost:8888/api/v1/overlays',
        expect.objectContaining({ method: 'POST' })
      );
    });

    it('updates overlay', async () => {
      const overlay = { id: '1', name: 'updated', createdAt: '2024-01-01' };
      mockJsonResponse(overlay);

      const result = await client.updateOverlay('1', { name: 'updated' });
      expect(result).toEqual(overlay);
      expect(fetchSpy).toHaveBeenCalledWith(
        'http://localhost:8888/api/v1/overlays/1',
        expect.objectContaining({ method: 'PUT' })
      );
    });

    it('deletes overlay', async () => {
      fetchSpy.mockResolvedValueOnce({
        ok: true,
        text: () => Promise.resolve(''),
      } as Response);

      await client.deleteOverlay('1');
      expect(fetchSpy).toHaveBeenCalledWith(
        'http://localhost:8888/api/v1/overlays/1',
        expect.objectContaining({ method: 'DELETE' })
      );
    });
  });

  describe('error handling', () => {
    it('throws YamsApiError on non-ok response', async () => {
      mockErrorResponse(404, 'Not found');

      await expect(client.getAction('nonexistent')).rejects.toThrow(YamsApiError);
    });

    it('includes status code in error', async () => {
      mockErrorResponse(500, 'Internal error');

      try {
        await client.status();
        expect.fail('Should have thrown');
      } catch (error) {
        expect(error).toBeInstanceOf(YamsApiError);
        expect((error as YamsApiError).status).toBe(500);
        expect((error as YamsApiError).message).toBe('Internal error');
      }
    });

    it('handles non-JSON error responses', async () => {
      fetchSpy.mockResolvedValueOnce({
        ok: false,
        status: 500,
        statusText: 'Internal Server Error',
        json: () => Promise.reject(new Error('Not JSON')),
      } as Response);

      try {
        await client.status();
        expect.fail('Should have thrown');
      } catch (error) {
        expect(error).toBeInstanceOf(YamsApiError);
        expect((error as YamsApiError).message).toBe('API error: Internal Server Error');
      }
    });
  });

  describe('URL encoding', () => {
    it('encodes special characters in action key', async () => {
      mockJsonResponse({ Name: 's3:GetObject' });

      await client.getAction('s3:GetObject');
      expect(fetchSpy).toHaveBeenCalledWith(
        expect.stringContaining('s3%3AGetObject'),
        expect.any(Object)
      );
    });

    it('encodes ARNs properly', async () => {
      mockJsonResponse({ Arn: 'arn:aws:iam::123:role/path/to/role' });

      await client.getPrincipal('arn:aws:iam::123:role/path/to/role');
      expect(fetchSpy).toHaveBeenCalledWith(
        expect.stringContaining(encodeURIComponent('arn:aws:iam::123:role/path/to/role')),
        expect.any(Object)
      );
    });
  });
});
