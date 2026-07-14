// ui/src/pages/simulate/shared/use-shared-context.ts
import { useEffect, useState } from 'react';
import { yamsApi } from '../../../lib/api';
import type { ContextVariable } from './context-editor';

// Fetches shared context from the server for display alongside user-provided context.
export function useSharedContext(): ContextVariable[] {
  const [sharedVars, setSharedVars] = useState<ContextVariable[]>([]);

  useEffect(() => {
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
