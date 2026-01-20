import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import {
  ActionIcon,
  Alert,
  Anchor,
  Badge,
  Box,
  Button,
  Card,
  Checkbox,
  Collapse,
  Combobox,
  Grid,
  Group,
  Input,
  InputBase,
  Loader,
  Pagination,
  ScrollArea,
  Stack,
  Table,
  Text,
  TextInput,
  Title,
  Tooltip,
  UnstyledButton,
  useCombobox,
} from '@mantine/core';
import { useDebouncedValue } from '@mantine/hooks';
import {
  IconChevronDown,
  IconChevronRight,
  IconCheck,
  IconOctagonFilled,
  IconPlayerPlay,
  IconPlus,
  IconX,
  IconSearch,
  IconLayersLinked,
} from '@tabler/icons-react';
import { Link, useSearchParams } from 'react-router-dom';
import { yamsApi } from '../../lib/api';
import type { SimulationResponse, OverlaySummary, OverlayData, SimulationOverlay } from '../../lib/api';

// Extract service type from ARN (3rd segment)
function extractService(arn: string): string | null {
  const parts = arn.split(':');
  if (parts.length >= 3 && parts[2]) {
    return parts[2];
  }
  return null;
}

// Tree node for trace visualization
interface TraceNode {
  text: string;
  depth: number;
  children: TraceNode[];
  isExpanded?: boolean;
}

// Parse trace lines into a tree structure based on indentation
function parseTraceToTree(trace: string[]): TraceNode[] {
  const root: TraceNode[] = [];
  const stack: { node: TraceNode; depth: number }[] = [];

  for (const line of trace) {
    const trimmed = line.trimStart();
    const depth = line.length - trimmed.length;

    const node: TraceNode = {
      text: trimmed,
      depth,
      children: [],
      isExpanded: depth < 4,
    };

    while (stack.length > 0 && stack[stack.length - 1].depth >= depth) {
      stack.pop();
    }

    if (stack.length === 0) {
      root.push(node);
    } else {
      stack[stack.length - 1].node.children.push(node);
    }

    stack.push({ node, depth });
  }

  return root;
}

// Highlight keywords in explanation text
function highlightExplanation(text: string): JSX.Element {
  // Decision badges - these get rendered as Badge components
  const decisions = [
    { pattern: /\[implicit deny\]/gi, color: 'red', label: 'IMPLICIT DENY' },
    { pattern: /\[explicit deny\]/gi, color: 'red', label: 'EXPLICIT DENY' },
    { pattern: /\[allow\]/gi, color: 'green', label: 'ALLOW' },
    { pattern: /\[deny\]/gi, color: 'red', label: 'DENY' },
  ];

  // Regular keyword highlights
  const keywords = [
    { pattern: /allow/gi, color: 'green' },
    { pattern: /deny/gi, color: 'red' },
    { pattern: /x-account/gi, color: 'orange' },
    { pattern: /missing/gi, color: 'orange' },
  ];

  const highlights: { start: number; end: number; color: string; badge?: string }[] = [];

  // Find decision badges first (higher priority)
  for (const { pattern, color, label } of decisions) {
    let match;
    const regex = new RegExp(pattern.source, pattern.flags);
    while ((match = regex.exec(text)) !== null) {
      highlights.push({ start: match.index, end: match.index + match[0].length, color, badge: label });
    }
  }

  // Find keyword highlights
  for (const { pattern, color } of keywords) {
    let match;
    const regex = new RegExp(pattern.source, pattern.flags);
    while ((match = regex.exec(text)) !== null) {
      highlights.push({ start: match.index, end: match.index + match[0].length, color });
    }
  }

  highlights.sort((a, b) => a.start - b.start);
  const filtered: typeof highlights = [];
  for (const h of highlights) {
    if (filtered.length === 0 || h.start >= filtered[filtered.length - 1].end) {
      filtered.push(h);
    }
  }

  if (filtered.length === 0) {
    return <Text size="sm">{text}</Text>;
  }

  const parts: JSX.Element[] = [];
  let lastEnd = 0;

  for (let i = 0; i < filtered.length; i++) {
    const h = filtered[i];
    if (h.start > lastEnd) {
      parts.push(<span key={`text-${i}`}>{text.slice(lastEnd, h.start)}</span>);
    }
    if (h.badge) {
      parts.push(
        <Badge key={`hl-${i}`} color={h.color} size="sm" variant="filled" style={{ verticalAlign: 'middle' }}>
          {h.badge}
        </Badge>
      );
    } else {
      parts.push(
        <Text key={`hl-${i}`} component="span" fw={600} c={h.color}>
          {text.slice(h.start, h.end)}
        </Text>
      );
    }
    lastEnd = h.end;
  }

  if (lastEnd < text.length) {
    parts.push(<span key="text-end">{text.slice(lastEnd)}</span>);
  }

  return <Text size="sm">{parts}</Text>;
}

// Linkify policy ARNs in trace text
function linkifyPolicyArns(text: string): JSX.Element {
  // Match policy ARNs: arn:aws:iam::account:policy/name or arn:aws:organizations::account:policy/type/id
  const arnPattern = /arn:aws:(iam|organizations)::[^:\s]+:policy\/[^\s,)]+/g;
  const matches: { start: number; end: number; arn: string }[] = [];

  let match;
  while ((match = arnPattern.exec(text)) !== null) {
    matches.push({ start: match.index, end: match.index + match[0].length, arn: match[0] });
  }

  if (matches.length === 0) {
    return <>{text}</>;
  }

  const parts: JSX.Element[] = [];
  let lastEnd = 0;

  for (let i = 0; i < matches.length; i++) {
    const m = matches[i];
    if (m.start > lastEnd) {
      parts.push(<span key={`text-${i}`}>{text.slice(lastEnd, m.start)}</span>);
    }
    parts.push(
      <Anchor
        key={`arn-${i}`}
        component={Link}
        to={`/search/policies?q=${encodeURIComponent(m.arn)}`}
        size="xs"
        ff="monospace"
      >
        {m.arn}
      </Anchor>
    );
    lastEnd = m.end;
  }

  if (lastEnd < text.length) {
    parts.push(<span key="text-end">{text.slice(lastEnd)}</span>);
  }

  return <>{parts}</>;
}

// Recursive trace tree node component
function TraceTreeNode({ node, level = 0 }: { node: TraceNode; level?: number }): JSX.Element {
  const [expanded, setExpanded] = useState(node.isExpanded ?? true);
  const hasChildren = node.children.length > 0;

  const isBegin = node.text.startsWith('begin:');
  const isEnd = node.text.startsWith('end:');
  const isMatch = node.text.includes('match:');
  const isDeny = node.text.includes('(deny)') || node.text.includes('[implicit deny]') || node.text.includes('does not match');
  const isAllow = node.text.includes('(allow)');

  let textColor: string | undefined;
  if (isDeny) textColor = 'red';
  else if (isAllow) textColor = 'green';
  else if (isMatch) textColor = 'teal';

  return (
    <Box>
      <Group gap={4} wrap="nowrap">
        {hasChildren ? (
          <UnstyledButton onClick={() => setExpanded(!expanded)} style={{ lineHeight: 1 }}>
            {expanded ? (
              <IconChevronDown size={14} color="var(--mantine-color-dimmed)" />
            ) : (
              <IconChevronRight size={14} color="var(--mantine-color-dimmed)" />
            )}
          </UnstyledButton>
        ) : (
          <Box w={14} />
        )}
        <Text
          size="xs"
          ff="monospace"
          c={textColor}
          fw={isBegin || isEnd ? 500 : undefined}
          style={{ opacity: isEnd ? 0.6 : 1 }}
        >
          {linkifyPolicyArns(node.text)}
        </Text>
      </Group>
      {hasChildren && (
        <Collapse in={expanded}>
          <Box pl="md" style={{ borderLeft: '1px solid var(--mantine-color-gray-3)' }}>
            {node.children.map((child, idx) => (
              <TraceTreeNode key={idx} node={child} level={level + 1} />
            ))}
          </Box>
        </Collapse>
      )}
    </Box>
  );
}

// Format relative time for dates
function formatRelativeTime(dateStr: string): string {
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

// Extract account ID from ARN (5th segment)
function extractAccountId(arn: string): string | null {
  const parts = arn.split(':');
  if (parts.length >= 5 && parts[4]) {
    return parts[4];
  }
  return null;
}

// Highlight matching text in search results
function highlightMatch(text: string, search: string): JSX.Element {
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
    // Add text before match
    if (matchIndex > lastIndex) {
      parts.push(<span key={keyIndex++}>{text.slice(lastIndex, matchIndex)}</span>);
    }
    // Add matched text in bold
    parts.push(
      <Text key={keyIndex++} component="span" fw={700}>
        {text.slice(matchIndex, matchIndex + search.length)}
      </Text>
    );
    lastIndex = matchIndex + search.length;
    matchIndex = lowerText.indexOf(lowerSearch, lastIndex);
  }

  // Add remaining text
  if (lastIndex < text.length) {
    parts.push(<span key={keyIndex}>{text.slice(lastIndex)}</span>);
  }

  return <>{parts}</>;
}

// Async search select component
interface AsyncSearchSelectProps {
  label: string;
  placeholder: string;
  value: string | null;
  onChange: (value: string | null) => void;
  onSearch: (query: string) => Promise<string[]>;
  formatLabel?: (value: string) => string;
  accountNames?: Record<string, string>;
  resourceAccounts?: Record<string, string>;
  accessLevels?: Record<string, string>;
  showAccountName?: boolean;
  showResourceType?: boolean;
  showAccessLevel?: boolean;
  disabled?: boolean;
  disabledMessage?: string;
}

function AsyncSearchSelect({
  label,
  placeholder,
  value,
  onChange,
  onSearch,
  formatLabel,
  accountNames,
  resourceAccounts,
  accessLevels,
  showAccountName,
  showResourceType,
  showAccessLevel,
  disabled,
  disabledMessage,
}: AsyncSearchSelectProps): JSX.Element {
  const combobox = useCombobox({
    onDropdownClose: () => combobox.resetSelectedOption(),
  });

  const [search, setSearch] = useState('');
  const [debouncedSearch] = useDebouncedValue(search, 300);
  const [loading, setLoading] = useState(false);
  const [options, setOptions] = useState<string[]>([]);

  // Search when debounced value changes
  useEffect(() => {
    if (debouncedSearch.length < 2) {
      setOptions([]);
      return;
    }

    setLoading(true);
    onSearch(debouncedSearch)
      .then((results) => {
        setOptions(results.slice(0, 100)); // Limit to 100 results
      })
      .catch((err) => {
        console.error('Search failed:', err);
        setOptions([]);
      })
      .finally(() => {
        setLoading(false);
      });
  }, [debouncedSearch, onSearch]);

  const displayValue = value ? (formatLabel ? formatLabel(value) : value) : '';

  const handleClear = (e: React.MouseEvent): void => {
    e.stopPropagation();
    onChange(null);
    setSearch('');
  };

  // Render disabled state
  if (disabled) {
    return (
      <Input.Wrapper label={label}>
        <Box
          style={{
            border: '1px solid var(--mantine-color-gray-3)',
            borderRadius: 'var(--mantine-radius-sm)',
            padding: '12px',
            height: '58px',
            display: 'flex',
            alignItems: 'center',
            backgroundColor: 'var(--mantine-color-gray-1)',
            cursor: 'not-allowed',
          }}
        >
          <Text size="sm" c="dimmed">
            {disabledMessage || 'Not available'}
          </Text>
        </Box>
      </Input.Wrapper>
    );
  }

  return (
    <Combobox
      store={combobox}
      onOptionSubmit={(val) => {
        onChange(val);
        setSearch('');
        combobox.closeDropdown();
      }}
    >
      <Combobox.Target>
        <Input.Wrapper label={label}>
          {value && !combobox.dropdownOpened ? (
            (() => {
              // Try ARN extraction first, then fall back to resourceAccounts mapping
              const accountId = showAccountName
                ? (extractAccountId(value) || (resourceAccounts ? resourceAccounts[value] : null))
                : null;
              const accountName = accountId && accountNames ? accountNames[accountId] : null;
              const service = showResourceType ? extractService(value) : null;
              const accessLevel = showAccessLevel && accessLevels ? accessLevels[value] : null;
              return (
                <Tooltip label={value} multiline maw={400} openDelay={500}>
                  <Group
                    gap="xs"
                    justify="space-between"
                    wrap="nowrap"
                    onClick={() => combobox.openDropdown()}
                    style={{
                      border: '1px solid var(--mantine-color-gray-4)',
                      borderRadius: 'var(--mantine-radius-sm)',
                      padding: '8px 12px',
                      height: '58px',
                      cursor: 'pointer',
                    }}
                  >
                    <Box style={{ overflow: 'hidden', flex: 1 }}>
                      <Text size="sm" truncate>
                        {displayValue}
                      </Text>
                      {showResourceType && (accountName || service) && (
                        <Text size="xs" c="dimmed" truncate>
                          {[accountName ? `${accountName} (${accountId})` : accountId, service].filter(Boolean).join(' · ')}
                        </Text>
                      )}
                      {showAccountName && !showResourceType && accountName && (
                        <Text size="xs" c="dimmed" truncate>
                          {accountName} ({accountId})
                        </Text>
                      )}
                      {showAccessLevel && accessLevel && (
                        <Text size="xs" c="dimmed" truncate>
                          {accessLevel}
                        </Text>
                      )}
                      {!showAccountName && !showResourceType && !showAccessLevel && formatLabel && (
                        <Text size="xs" c="dimmed" truncate>
                          {value}
                        </Text>
                      )}
                    </Box>
                    <ActionIcon
                      size="sm"
                      variant="subtle"
                      color="gray"
                      onClick={handleClear}
                      aria-label="Clear selection"
                    >
                      <IconX size={14} />
                    </ActionIcon>
                  </Group>
                </Tooltip>
              );
            })()
          ) : (
            <InputBase
              size="lg"
              rightSection={loading ? <Loader size={18} /> : <IconSearch size={18} color="var(--mantine-color-dimmed)" />}
              value={search}
              onChange={(event) => {
                setSearch(event.currentTarget.value);
                combobox.openDropdown();
                combobox.updateSelectedOptionIndex();
              }}
              onClick={() => combobox.openDropdown()}
              onFocus={() => combobox.openDropdown()}
              onBlur={() => {
                combobox.closeDropdown();
                setSearch('');
              }}
              placeholder={placeholder}
              styles={{ input: { height: '58px' } }}
              rightSectionPointerEvents="none"
            />
          )}
        </Input.Wrapper>
      </Combobox.Target>

      <Combobox.Dropdown>
        <Combobox.Options>
          <ScrollArea.Autosize mah={300} type="scroll">
            {loading ? (
              <Combobox.Empty>Searching...</Combobox.Empty>
            ) : options.length === 0 ? (
              <Combobox.Empty>
                {debouncedSearch.length < 2 ? 'Type at least 2 characters to search' : 'No results found'}
              </Combobox.Empty>
            ) : (
              options.map((option) => {
                // Try ARN extraction first, then fall back to resourceAccounts mapping
                const accountId = showAccountName
                  ? (extractAccountId(option) || (resourceAccounts ? resourceAccounts[option] : null))
                  : null;
                const accountName = accountId && accountNames ? accountNames[accountId] : null;
                const service = showResourceType ? extractService(option) : null;
                const accessLevel = showAccessLevel && accessLevels ? accessLevels[option] : null;
                return (
                  <Combobox.Option value={option} key={option}>
                    <Text size="sm" truncate>
                      {highlightMatch(formatLabel ? formatLabel(option) : option, debouncedSearch)}
                    </Text>
                    {showResourceType && (accountName || service) && (
                      <Text size="xs" c="dimmed" truncate>
                        {[accountName ? `${accountName} (${accountId})` : accountId, service].filter(Boolean).join(' · ')}
                      </Text>
                    )}
                    {showAccountName && !showResourceType && accountName && (
                      <Text size="xs" c="dimmed" truncate>
                        {accountName} ({accountId})
                      </Text>
                    )}
                    {showAccessLevel && accessLevel && (
                      <Text size="xs" c="dimmed" truncate>
                        {accessLevel}
                      </Text>
                    )}
                    {!showAccountName && !showResourceType && !showAccessLevel && formatLabel && (
                      <Text size="xs" c="dimmed" truncate>
                        {option}
                      </Text>
                    )}
                  </Combobox.Option>
                );
              })
            )}
          </ScrollArea.Autosize>
        </Combobox.Options>
      </Combobox.Dropdown>
    </Combobox>
  );
}

// Extract display name from principal ARN
function formatPrincipalLabel(arn: string): string {
  const parts = arn.split('/');
  return parts[parts.length - 1];
}

// Extract display name from resource ARN
function formatResourceLabel(arn: string): string {
  const parts = arn.split(':');
  return parts.slice(5).join(':') || arn;
}

export function AccessCheckPage(): JSX.Element {
  const [searchParams, setSearchParams] = useSearchParams();

  // Initialize state from URL params
  const [selectedPrincipal, setSelectedPrincipal] = useState<string | null>(
    searchParams.get('principal')
  );
  const [selectedAction, setSelectedAction] = useState<string | null>(
    searchParams.get('action')
  );
  const [selectedResource, setSelectedResource] = useState<string | null>(
    searchParams.get('resource')
  );

  // Update URL when selections change
  const updateSelection = useCallback(
    (key: 'principal' | 'action' | 'resource', value: string | null): void => {
      setSearchParams((prev) => {
        const next = new URLSearchParams(prev);
        if (value) {
          next.set(key, value);
        } else {
          next.delete(key);
        }
        return next;
      }, { replace: true });

      if (key === 'principal') setSelectedPrincipal(value);
      else if (key === 'action') setSelectedAction(value);
      else if (key === 'resource') setSelectedResource(value);
    },
    [setSearchParams]
  );

  // Account names and resource-to-account mapping for display
  const [accountNames, setAccountNames] = useState<Record<string, string>>({});
  const [resourceAccounts, setResourceAccounts] = useState<Record<string, string>>({});

  // Resourceless actions - actions that don't require a resource
  const [resourcelessActions, setResourcelessActions] = useState<Set<string>>(new Set());

  // Action access levels for display
  const [actionAccessLevels, setActionAccessLevels] = useState<Record<string, string>>({});

  // Request context variables
  const [contextVars, setContextVars] = useState<Array<{ key: string; value: string }>>([]);

  // Overlay selection
  const [overlays, setOverlays] = useState<OverlaySummary[]>([]);
  const [selectedOverlayIds, setSelectedOverlayIds] = useState<Set<string>>(new Set());
  const [loadedOverlays, setLoadedOverlays] = useState<Map<string, OverlayData>>(new Map());
  const [showOverlaySelector, setShowOverlaySelector] = useState(false);
  const [overlaySearchQuery, setOverlaySearchQuery] = useState('');
  const [overlayPage, setOverlayPage] = useState(1);

  // Simulation states
  const [simulating, setSimulating] = useState(false);
  const [result, setResult] = useState<SimulationResponse | null>(null);
  const [error, setError] = useState<string | null>(null);

  // Fetch account names, resource accounts, resourceless actions, access levels, and overlays on mount
  useEffect(() => {
    yamsApi.accountNames()
      .then(setAccountNames)
      .catch((err) => console.error('Failed to fetch account names:', err));
    yamsApi.resourceAccounts()
      .then(setResourceAccounts)
      .catch((err) => console.error('Failed to fetch resource accounts:', err));
    yamsApi.resourcelessActions()
      .then((actions) => setResourcelessActions(new Set(actions)))
      .catch((err) => console.error('Failed to fetch resourceless actions:', err));
    yamsApi.actionAccessLevels()
      .then(setActionAccessLevels)
      .catch((err) => console.error('Failed to fetch action access levels:', err));
    yamsApi.listOverlays()
      .then((list) => {
        // Sort by createdAt descending (newest first)
        list.sort((a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime());
        setOverlays(list);
      })
      .catch((err) => console.error('Failed to fetch overlays:', err));
  }, []);

  // Check if current action is resourceless
  const isActionResourceless = useMemo(() => {
    if (!selectedAction) return false;
    return resourcelessActions.has(selectedAction);
  }, [selectedAction, resourcelessActions]);

  // Overlay filtering and pagination
  const [debouncedOverlaySearch] = useDebouncedValue(overlaySearchQuery, 200);
  const OVERLAYS_PER_PAGE = 5;

  const filteredOverlays = useMemo(() => {
    if (!debouncedOverlaySearch) return overlays;
    const query = debouncedOverlaySearch.toLowerCase();
    return overlays.filter(
      (o) => o.name.toLowerCase().includes(query) || o.id.toLowerCase().includes(query)
    );
  }, [overlays, debouncedOverlaySearch]);

  const totalOverlayPages = Math.ceil(filteredOverlays.length / OVERLAYS_PER_PAGE);

  const paginatedOverlays = useMemo(() => {
    const start = (overlayPage - 1) * OVERLAYS_PER_PAGE;
    return filteredOverlays.slice(start, start + OVERLAYS_PER_PAGE);
  }, [filteredOverlays, overlayPage]);

  // Reset page when search changes
  useEffect(() => {
    setOverlayPage(1);
  }, [debouncedOverlaySearch]);

  // Clear resource when selecting a resourceless action
  useEffect(() => {
    if (isActionResourceless && selectedResource) {
      updateSelection('resource', null);
    }
  }, [isActionResourceless, selectedResource, updateSelection]);

  // Load overlay data when selection changes
  useEffect(() => {
    const loadMissingOverlays = async (): Promise<void> => {
      const toLoad = Array.from(selectedOverlayIds).filter((id) => !loadedOverlays.has(id));
      if (toLoad.length === 0) return;

      const loaded = new Map(loadedOverlays);
      await Promise.all(
        toLoad.map(async (id) => {
          try {
            const data = await yamsApi.getOverlay(id);
            loaded.set(id, data);
          } catch (err) {
            console.error(`Failed to load overlay ${id}:`, err);
          }
        })
      );
      setLoadedOverlays(loaded);
    };
    loadMissingOverlays();
  }, [selectedOverlayIds, loadedOverlays]);

  // Build combined overlay from selected overlays
  const buildCombinedOverlay = useCallback((): SimulationOverlay | undefined => {
    if (selectedOverlayIds.size === 0) return undefined;

    const combined: SimulationOverlay = {
      accounts: [],
      groups: [],
      policies: [],
      principals: [],
      resources: [],
    };

    for (const id of selectedOverlayIds) {
      const data = loadedOverlays.get(id);
      if (!data) continue;

      if (data.accounts) combined.accounts!.push(...data.accounts);
      if (data.groups) combined.groups!.push(...data.groups);
      if (data.policies) combined.policies!.push(...data.policies);
      if (data.principals) combined.principals!.push(...data.principals);
      if (data.resources) combined.resources!.push(...data.resources);
    }

    // Return undefined if all arrays are empty
    const hasData =
      combined.accounts!.length > 0 ||
      combined.groups!.length > 0 ||
      combined.policies!.length > 0 ||
      combined.principals!.length > 0 ||
      combined.resources!.length > 0;

    return hasData ? combined : undefined;
  }, [selectedOverlayIds, loadedOverlays]);

  // Search functions
  const searchPrincipals = useCallback((query: string) => yamsApi.searchPrincipals(query), []);
  const searchActions = useCallback((query: string) => yamsApi.searchActions(query), []);
  const searchResources = useCallback((query: string) => yamsApi.searchResources(query), []);

  // Use ref to access context vars without triggering re-renders
  const contextVarsRef = useRef(contextVars);
  contextVarsRef.current = contextVars;

  // Build context object from key-value pairs
  const buildContext = (): Record<string, string> | undefined => {
    const validPairs = contextVarsRef.current.filter((cv) => cv.key.trim() && cv.value.trim());
    if (validPairs.length === 0) return undefined;
    return Object.fromEntries(validPairs.map((cv) => [cv.key.trim(), cv.value.trim()]));
  };

  // Run simulation when required inputs are selected
  const runSimulation = useCallback(async (): Promise<void> => {
    // For resourceless actions, we only need principal and action
    const needsResource = !resourcelessActions.has(selectedAction ?? '');
    if (!selectedPrincipal || !selectedAction) return;
    if (needsResource && !selectedResource) return;

    setSimulating(true);
    setError(null);
    setResult(null);

    try {
      const response = await yamsApi.simulate({
        principal: selectedPrincipal,
        action: selectedAction,
        resource: needsResource ? selectedResource! : '*',
        context: buildContext(),
        explain: true,
        trace: true,
        overlay: buildCombinedOverlay(),
      });
      setResult(response);
    } catch (err) {
      console.error('Simulation failed:', err);
      setError(err instanceof Error ? err.message : 'Simulation failed');
    } finally {
      setSimulating(false);
    }
  }, [selectedPrincipal, selectedAction, selectedResource, resourcelessActions, buildCombinedOverlay]);

  // Stable reference for context vars to include in dependency array
  const contextVarsJson = JSON.stringify(contextVars);

  // Auto-run simulation when required selections, overlays, or context vars change
  useEffect(() => {
    const needsResource = !resourcelessActions.has(selectedAction ?? '');
    const hasRequiredInputs = selectedPrincipal && selectedAction && (needsResource ? selectedResource : true);
    if (hasRequiredInputs) {
      runSimulation();
    } else {
      setResult(null);
      setError(null);
    }
  }, [selectedPrincipal, selectedAction, selectedResource, resourcelessActions, runSimulation, selectedOverlayIds, loadedOverlays, contextVarsJson]);

  // Parse trace into tree
  const traceTree = useMemo(() => {
    if (!result?.trace) return [];
    return parseTraceToTree(result.trace);
  }, [result?.trace]);

  // Context variable helpers
  const addContextVar = (): void => {
    setContextVars([...contextVars, { key: '', value: '' }]);
  };

  const removeContextVar = (index: number): void => {
    setContextVars(contextVars.filter((_, i) => i !== index));
  };

  const updateContextVar = (index: number, field: 'key' | 'value', val: string): void => {
    setContextVars(contextVars.map((cv, i) => (i === index ? { ...cv, [field]: val } : cv)));
  };

  const allSelected = selectedPrincipal && selectedAction && (isActionResourceless || selectedResource);
  const hasAnySelection = selectedPrincipal || selectedAction || selectedResource || selectedOverlayIds.size > 0;

  const toggleOverlay = (id: string): void => {
    setSelectedOverlayIds((prev) => {
      const next = new Set(prev);
      if (next.has(id)) {
        next.delete(id);
      } else {
        next.add(id);
      }
      return next;
    });
  };

  const clearAll = (): void => {
    updateSelection('principal', null);
    updateSelection('action', null);
    updateSelection('resource', null);
    setContextVars([]);
    setSelectedOverlayIds(new Set());
  };

  return (
    <Box p="md">
      <Stack gap="lg">
        {/* Page header */}
        <Group justify="space-between" align="center">
          <Box>
            <Title order={3} mb={4}>Access Check</Title>
            <Text size="sm" c="dimmed">
              Test whether a <Text component="span" fw={500} c="purple.6">principal</Text> can
              perform an <Text component="span" fw={500} c="purple.6">action</Text> on
              a <Text component="span" fw={500} c="purple.6">resource</Text>.
            </Text>
          </Box>
          {hasAnySelection && (
            <Button
              variant="subtle"
              color="gray"
              size="xs"
              leftSection={<IconX size={14} />}
              onClick={clearAll}
            >
              Clear All
            </Button>
          )}
        </Group>

        {/* Selection dropdowns */}
        <Card withBorder p="lg">
          <Grid gutter="md">
            <Grid.Col span={4}>
              <AsyncSearchSelect
                label="Principal"
                placeholder="Search principals..."
                value={selectedPrincipal}
                onChange={(v) => updateSelection('principal', v)}
                onSearch={searchPrincipals}
                formatLabel={formatPrincipalLabel}
                accountNames={accountNames}
                showAccountName
              />
            </Grid.Col>
            <Grid.Col span={4}>
              <AsyncSearchSelect
                label="Action"
                placeholder="Search actions..."
                value={selectedAction}
                onChange={(v) => updateSelection('action', v)}
                onSearch={searchActions}
                accessLevels={actionAccessLevels}
                showAccessLevel
              />
            </Grid.Col>
            <Grid.Col span={4}>
              <AsyncSearchSelect
                label="Resource"
                placeholder="Search resources..."
                value={selectedResource}
                onChange={(v) => updateSelection('resource', v)}
                onSearch={searchResources}
                formatLabel={formatResourceLabel}
                accountNames={accountNames}
                resourceAccounts={resourceAccounts}
                showAccountName
                showResourceType
                disabled={isActionResourceless}
                disabledMessage="Not required for this action"
              />
            </Grid.Col>
          </Grid>
        </Card>

        {/* Overlay selection */}
        <Card withBorder p="lg">
          <Group justify="space-between" mb={showOverlaySelector || selectedOverlayIds.size > 0 ? 'md' : undefined}>
            <Group gap="xs">
              <Title order={5}>Overlays</Title>
              {selectedOverlayIds.size > 0 && (
                <Badge size="sm" variant="light" color="violet">
                  {selectedOverlayIds.size} selected
                </Badge>
              )}
            </Group>
            {!showOverlaySelector && overlays.length > 0 && (
              <Button
                variant="subtle"
                size="xs"
                leftSection={<IconPlus size={14} />}
                onClick={() => setShowOverlaySelector(true)}
              >
                Add Overlay
              </Button>
            )}
          </Group>

          {/* Selected overlays summary */}
          {selectedOverlayIds.size > 0 && !showOverlaySelector && (
            <Stack gap="xs" mb="md">
              {Array.from(selectedOverlayIds).map((id) => {
                const overlay = overlays.find((o) => o.id === id);
                if (!overlay) return null;
                return (
                  <Group key={id} gap="xs" justify="space-between" p="xs" style={{ backgroundColor: 'var(--mantine-color-violet-0)', borderRadius: 'var(--mantine-radius-sm)' }}>
                    <Group gap="xs">
                      <IconLayersLinked size={14} color="var(--mantine-color-violet-6)" />
                      <Text size="sm" fw={500}>{overlay.name}</Text>
                      <Text size="xs" c="dimmed">
                        {overlay.numPrincipals} {overlay.numPrincipals === 1 ? 'Principal' : 'Principals'} · {overlay.numResources} {overlay.numResources === 1 ? 'Resource' : 'Resources'} · {overlay.numPolicies} {overlay.numPolicies === 1 ? 'Policy' : 'Policies'}
                      </Text>
                    </Group>
                    <ActionIcon size="sm" variant="subtle" color="gray" onClick={() => toggleOverlay(id)}>
                      <IconX size={14} />
                    </ActionIcon>
                  </Group>
                );
              })}
              <Button
                variant="subtle"
                size="xs"
                leftSection={<IconPlus size={14} />}
                onClick={() => setShowOverlaySelector(true)}
                style={{ alignSelf: 'flex-start' }}
              >
                Add More
              </Button>
            </Stack>
          )}

          {/* Overlay selector table */}
          {showOverlaySelector && (
            <Stack gap="sm">
              <Group justify="space-between">
                <TextInput
                  placeholder="Search overlays..."
                  leftSection={<IconSearch size={14} />}
                  size="sm"
                  value={overlaySearchQuery}
                  onChange={(e) => setOverlaySearchQuery(e.currentTarget.value)}
                  style={{ flex: 1, maxWidth: 300 }}
                />
                <Button
                  variant="subtle"
                  size="xs"
                  color="gray"
                  leftSection={<IconCheck size={14} />}
                  onClick={() => {
                    setShowOverlaySelector(false);
                    setOverlaySearchQuery('');
                    setOverlayPage(1);
                  }}
                >
                  Done
                </Button>
              </Group>

              {filteredOverlays.length === 0 ? (
                <Text size="sm" c="dimmed" ta="center" py="md">
                  {overlaySearchQuery ? 'No overlays match your search' : 'No overlays available'}
                </Text>
              ) : (
                <>
                  <Table striped highlightOnHover>
                    <Table.Thead>
                      <Table.Tr>
                        <Table.Th style={{ width: 40 }}></Table.Th>
                        <Table.Th>Name</Table.Th>
                        <Table.Th style={{ width: 220 }}>ID</Table.Th>
                        <Table.Th style={{ width: 90 }}>Created</Table.Th>
                      </Table.Tr>
                    </Table.Thead>
                    <Table.Tbody>
                      {paginatedOverlays.map((overlay) => {
                        const isSelected = selectedOverlayIds.has(overlay.id);
                        return (
                          <Table.Tr
                            key={overlay.id}
                            style={{ cursor: 'pointer' }}
                            onClick={() => toggleOverlay(overlay.id)}
                          >
                            <Table.Td>
                              <Checkbox
                                checked={isSelected}
                                onChange={() => toggleOverlay(overlay.id)}
                                onClick={(e) => e.stopPropagation()}
                              />
                            </Table.Td>
                            <Table.Td>
                              <Group gap="xs" wrap="nowrap">
                                <IconLayersLinked size={14} color="var(--mantine-color-violet-6)" style={{ flexShrink: 0 }} />
                                <Text size="sm" fw={500} truncate style={{ maxWidth: 200 }}>
                                  {overlay.name}
                                </Text>
                              </Group>
                            </Table.Td>
                            <Table.Td>
                              <Tooltip label={overlay.id} openDelay={300}>
                                <Text size="xs" ff="monospace" c="dimmed" truncate style={{ maxWidth: 200 }}>
                                  {overlay.id}
                                </Text>
                              </Tooltip>
                            </Table.Td>
                            <Table.Td>
                              <Text size="xs" c="dimmed">
                                {formatRelativeTime(overlay.createdAt)}
                              </Text>
                            </Table.Td>
                          </Table.Tr>
                        );
                      })}
                    </Table.Tbody>
                  </Table>

                  {totalOverlayPages > 1 && (
                    <Group justify="space-between" align="center">
                      <Text size="xs" c="dimmed">
                        {filteredOverlays.length} overlay{filteredOverlays.length !== 1 ? 's' : ''}
                      </Text>
                      <Pagination
                        value={overlayPage}
                        onChange={setOverlayPage}
                        total={totalOverlayPages}
                        size="sm"
                      />
                    </Group>
                  )}
                </>
              )}
            </Stack>
          )}

          {/* Empty state when no overlays and selector not shown */}
          {!showOverlaySelector && selectedOverlayIds.size === 0 && (
            <Text size="sm" c="dimmed">
              {overlays.length === 0
                ? 'No overlays available. Create overlays to test against hypothetical environments.'
                : 'Add overlays to test against hypothetical environments.'}
            </Text>
          )}
        </Card>

        {/* Request context variables */}
        <Card withBorder p="lg">
          <Group justify="space-between" mb={contextVars.length > 0 ? 'md' : undefined}>
            <Title order={5}>Request Context</Title>
            <Button
              variant="subtle"
              size="xs"
              leftSection={<IconPlus size={14} />}
              onClick={addContextVar}
            >
              Add Variable
            </Button>
          </Group>
          {contextVars.length > 0 && (
            <Stack gap="xs">
              {contextVars.map((cv, idx) => (
                <Group key={idx} gap="xs" wrap="nowrap">
                  <TextInput
                    placeholder="Key (e.g., aws:SourceIp)"
                    value={cv.key}
                    onChange={(e) => updateContextVar(idx, 'key', e.currentTarget.value)}
                    style={{ flex: 1 }}
                    size="sm"
                  />
                  <TextInput
                    placeholder="Value"
                    value={cv.value}
                    onChange={(e) => updateContextVar(idx, 'value', e.currentTarget.value)}
                    style={{ flex: 1 }}
                    size="sm"
                  />
                  <ActionIcon
                    variant="subtle"
                    color="gray"
                    onClick={() => removeContextVar(idx)}
                    aria-label="Remove variable"
                  >
                    <IconX size={16} />
                  </ActionIcon>
                </Group>
              ))}
              {allSelected && (
                <Button size="xs" variant="light" onClick={runSimulation} mt="xs" leftSection={<IconPlayerPlay size={14} />}>
                  Re-run Simulation
                </Button>
              )}
            </Stack>
          )}
          {contextVars.length === 0 && (
            <Text size="sm" c="dimmed">
              Add context variables to test conditions like aws:SourceIp, aws:RequestTag/*, etc.
            </Text>
          )}
        </Card>

        {/* Results section */}
        {simulating && (
          <Card withBorder p="lg">
            <Group justify="center" p="xl">
              <Loader size="md" />
              <Text c="dimmed">Running simulation...</Text>
            </Group>
          </Card>
        )}

        {error && (
          <Alert color="red" title="Error" icon={<IconX size={16} />}>
            {error}
          </Alert>
        )}

        {result && !simulating && (
          <Stack gap="md">
            {/* Input section */}
            <Card withBorder p="lg">
              <Title order={4} mb="md">Input</Title>
              <Group gap="xl">
                <div>
                  <Text size="sm" c="dimmed">Principal</Text>
                  <Anchor component={Link} to={`/search/principals?q=${encodeURIComponent(result.principal)}`} size="sm" ff="monospace">
                    {result.principal}
                  </Anchor>
                </div>
                <div>
                  <Text size="sm" c="dimmed">Action</Text>
                  <Anchor component={Link} to={`/search/actions?q=${encodeURIComponent(result.action)}`} size="sm" ff="monospace">
                    {result.action}
                  </Anchor>
                </div>
                <div>
                  <Text size="sm" c="dimmed">Resource</Text>
                  {!result.resource || result.resource === '*' ? (
                    <Text size="sm" c="dimmed" ff="monospace" fs="italic">
                      N/A (resourceless action)
                    </Text>
                  ) : (
                    <Anchor component={Link} to={`/search/resources?q=${encodeURIComponent(result.resource)}`} size="sm" ff="monospace">
                      {result.resource}
                    </Anchor>
                  )}
                </div>
              </Group>
            </Card>

            {/* Result section */}
            <Card withBorder p="lg">
              <Title order={4} mb="md">Result</Title>
              {result.result === 'ALLOW' ? (
                <Badge size="xl" color="green" leftSection={<IconCheck size={14} />}>
                  ALLOW
                </Badge>
              ) : (
                <Badge size="xl" color="red" leftSection={<IconOctagonFilled size={14} />}>
                  DENY
                </Badge>
              )}
            </Card>

            {/* Explanation */}
            {result.explain && result.explain.length > 0 && (
              <Card withBorder p="lg">
                <Title order={4} mb="md">Explanation</Title>
                <Stack gap="xs">
                  {result.explain.map((line, idx) => (
                    <Box key={idx}>{highlightExplanation(line)}</Box>
                  ))}
                </Stack>
              </Card>
            )}

            {/* Trace tree */}
            {traceTree.length > 0 && (
              <Card withBorder p="lg">
                <Title order={4} mb="md">Evaluation Trace</Title>
                <Box
                  p="md"
                  bg="gray.0"
                  style={{ borderRadius: 'var(--mantine-radius-sm)', overflow: 'auto', maxHeight: '500px' }}
                >
                  {traceTree.map((node, idx) => (
                    <TraceTreeNode key={idx} node={node} />
                  ))}
                </Box>
              </Card>
            )}
          </Stack>
        )}

        {/* Prompt when not all selected */}
        {!allSelected && !simulating && !result && (
          <Card withBorder p="xl">
            <Text ta="center" c="dimmed" size="lg">
              Search and select a principal, action, and resource to run an access check simulation.
            </Text>
          </Card>
        )}
      </Stack>
    </Box>
  );
}
