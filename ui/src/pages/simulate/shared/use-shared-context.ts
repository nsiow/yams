// ui/src/pages/simulate/shared/use-shared-context.ts
import { useEffect, useState } from 'react';
import { yamsApi } from '../../../lib/api';
import type { ContextVariable } from './context-editor';

const SHARED_CONTEXT_KEY = 'yams.enable_shared_request_context';

// Fetches shared context from the server if the user has opted in via localStorage.
export function useSharedContext(): ContextVariable[] {
  const [sharedVars, setSharedVars] = useState<ContextVariable[]>([]);

  useEffect(() => {
    if (localStorage.getItem(SHARED_CONTEXT_KEY) !== 'true') {
      setSharedVars([]);
      return;
    }

    yamsApi.sharedContext()
      .then((ctx) => {
        setSharedVars(
          Object.entries(ctx).map(([key, value]) => ({ key, value }))
        );
      })
      .catch((err) => {
        console.error('Failed to fetch shared context:', err);
        setSharedVars([]);
      });
  }, []);

  return sharedVars;
}
