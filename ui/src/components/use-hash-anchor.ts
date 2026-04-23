import { useEffect, useState } from 'react';
import { useLocation } from 'react-router-dom';

// useHashAnchor scrolls to the element matching the current URL hash and
// returns its id so callers can render a transient highlight for ~2s.
// Depends on `ready` so callers can defer until after async content has
// rendered (e.g. a detail fetch completes).
export function useHashAnchor(ready: boolean): string | null {
  const { hash } = useLocation();
  const [highlighted, setHighlighted] = useState<string | null>(null);

  useEffect(() => {
    if (!ready || !hash) {
      setHighlighted(null);
      return;
    }

    const id = decodeURIComponent(hash.replace(/^#/, ''));
    if (!id) return;

    const el = document.getElementById(id);
    if (!el) return;

    el.scrollIntoView({ block: 'start', behavior: 'smooth' });
    setHighlighted(id);

    const timer = window.setTimeout(() => setHighlighted(null), 2000);
    return () => window.clearTimeout(timer);
  }, [hash, ready]);

  return highlighted;
}

// highlightStyle returns the inline style to apply to a policy Box/Card
// when it is the current hash target.
export function highlightStyle(active: boolean): React.CSSProperties {
  return {
    transition: 'background-color 600ms ease-out, box-shadow 600ms ease-out',
    backgroundColor: active ? 'var(--mantine-color-purple-1)' : undefined,
    boxShadow: active ? '0 0 0 2px var(--mantine-color-purple-4)' : undefined,
    borderRadius: active ? 'var(--mantine-radius-sm)' : undefined,
  };
}
