// ui/src/pages/simulate/shared/utils.ts
// Shared utility functions for simulation pages

// Extract service type from ARN (3rd segment)
export function extractService(arn: string): string | null {
  const parts = arn.split(':');
  if (parts.length >= 3 && parts[2]) {
    return parts[2];
  }
  return null;
}

// S3 type detection functions

export function isS3Object(arn: string): boolean {
  return arn.startsWith('arn:aws:s3:::') && arn.includes('/');
}

export function isS3Bucket(arn: string): boolean {
  return arn.startsWith('arn:aws:s3:::') && !arn.includes('/');
}

export function getS3BucketFromObject(objectArn: string): string {
  const idx = objectArn.indexOf('/');
  return idx > -1 ? objectArn.substring(0, idx) : objectArn;
}

export function getS3ObjectPath(objectArn: string): string {
  const idx = objectArn.indexOf('/');
  return idx > -1 ? objectArn.substring(idx + 1) : '';
}

// Extract account ID from ARN (5th segment)
export function extractAccountId(arn: string): string | null {
  const parts = arn.split(':');
  if (parts.length >= 5 && parts[4]) {
    return parts[4];
  }
  return null;
}

// Extract display name from principal ARN
export function formatPrincipalLabel(arn: string): string {
  const parts = arn.split('/');
  return parts[parts.length - 1];
}

// Extract display name from resource ARN
export function formatResourceLabel(arn: string): string {
  const parts = arn.split(':');
  return parts.slice(5).join(':') || arn;
}

// Format relative time for dates
export function formatRelativeTime(dateStr: string): string {
  const date = new Date(dateStr);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffSecs = Math.floor(diffMs / 1000);
  const diffMins = Math.floor(diffSecs / 60);
  const diffHours = Math.floor(diffMins / 60);
  const diffDays = Math.floor(diffHours / 24);
  const diffWeeks = Math.floor(diffDays / 7);
  const diffMonths = Math.floor(diffDays / 30);
  const diffYears = Math.floor(diffDays / 365);

  if (diffSecs < 60) return 'just now';
  if (diffMins < 60) return `${diffMins}m ago`;
  if (diffHours < 24) return `${diffHours}h ago`;
  if (diffDays < 7) return `${diffDays}d ago`;
  if (diffWeeks < 5) return `${diffWeeks}w ago`;
  if (diffMonths < 12) return `${diffMonths}mo ago`;
  return `${diffYears}y ago`;
}

// Highlight matching text in search results
export function highlightMatch(text: string, search: string): JSX.Element {
  if (!search || search.length < 2) {
    return <>{text}</>;
  }

  const parts: JSX.Element[] = [];
  const lowerText = text.toLowerCase();
  const lowerSearch = search.toLowerCase();
  let lastIndex = 0;
  let matchIndex = lowerText.indexOf(lowerSearch);
  let keyIndex = 0;

  while (matchIndex !== -1) {
    if (matchIndex > lastIndex) {
      parts.push(<span key={keyIndex++}>{text.slice(lastIndex, matchIndex)}</span>);
    }
    parts.push(
      <span key={keyIndex++} style={{ fontWeight: 700 }}>
        {text.slice(matchIndex, matchIndex + search.length)}
      </span>
    );
    lastIndex = matchIndex + search.length;
    matchIndex = lowerText.indexOf(lowerSearch, lastIndex);
  }

  if (lastIndex < text.length) {
    parts.push(<span key={keyIndex}>{text.slice(lastIndex)}</span>);
  }

  return <>{parts}</>;
}

// Build URL for access check with given params
export function buildAccessCheckUrl(params: {
  principal?: string;
  action?: string;
  resource?: string;
}): string {
  const searchParams = new URLSearchParams();
  if (params.principal) searchParams.set('principal', params.principal);
  if (params.action) searchParams.set('action', params.action);
  if (params.resource) searchParams.set('resource', params.resource);
  return `/simulate/access?${searchParams.toString()}`;
}
