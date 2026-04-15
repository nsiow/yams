import { YamsClient } from './client';

export { YamsClient, YamsApiError } from './client';
export type { YamsClientConfig } from './client';
export * from './types';

// Default client instance pointing to localhost
// Can be reconfigured by creating a new YamsClient with different options
export const yamsApi = new YamsClient();
