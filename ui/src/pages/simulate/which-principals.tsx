// ui/src/pages/simulate/which-principals.tsx
import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import {
  Alert,
  Anchor,
  Badge,
  Box,
  Button,
  Card,
  Collapse,
  Grid,
  Group,
  Loader,
  Pagination,
  Stack,
  Table,
  Text,
  TextInput,
  Title,
  UnstyledButton,
} from '@mantine/core';
import { useDebouncedValue } from '@mantine/hooks';
import {
  IconArrowUpRight,
  IconCheck,
  IconChevronDown,
  IconChevronUp,
  IconFlask,
  IconOctagonFilled,
  IconSearch,
  IconX,
} from '@tabler/icons-react';
import { Link, useSearchParams } from 'react-router-dom';
import { yamsApi } from '../../lib/api';
import type { OverlayData, SimulationResponse } from '../../lib/api';
import {
  AsyncSearchSelect,
  OverlaySelector,
  buildCombinedOverlay,
  ContextEditor,
  buildContext,
  formatPrincipalLabel,
  formatResourceLabel,
  extractAccountId,
  buildAccessCheckUrl,
} from './shared';
import type { ContextVariable } from './shared';

const RESULTS_PER_PAGE = 20;

interface ExpandedRowData {
  loading: boolean;
  result: SimulationResponse | null;
  error: string | null;
}

export function WhichPrincipalsPage(): JSX.Element {
  const [searchParams, setSearchParams] = useSearchParams();

  // Initialize state from URL params
  const [selectedAction, setSelectedAction] = useState<string | null>(
    searchParams.get('action')
  );
  const [selectedResource, setSelectedResource] = useState<string | null>(
    searchParams.get('resource')
  );

  // Update URL when selections change
  const updateSelection = useCallback(
    (key: 'action' | 'resource', value: string | null): void => {
      setSearchParams((prev) => {
        const next = new URLSearchParams(prev);
        if (value) {
          next.set(key, value);
        } else {
          next.delete(key);
        }
        return next;
      }, { replace: true });

      if (key === 'action') setSelectedAction(value);
      else if (key === 'resource') setSelectedResource(value);
    },
    [setSearchParams]
  );

  // Account names for display
  const [accountNames, setAccountNames] = useState<Record<string, string>>({});
  const [resourceAccounts, setResourceAccounts] = useState<Record<string, string>>({});
  const [actionAccessLevels, setActionAccessLevels] = useState<Record<string, string>>({});

  // Context variables
  const [contextVars, setContextVars] = useState<ContextVariable[]>([]);

  // Overlay selection
  const [selectedOverlayIds, setSelectedOverlayIds] = useState<Set<string>>(new Set());
  const [loadedOverlays, setLoadedOverlays] = useState<Map<string, OverlayData>>(new Map());

  // Query states
  const [loading, setLoading] = useState(false);
  const [results, setResults] = useState<string[]>([]);
  const [error, setError] = useState<string | null>(null);

  // Results filtering and pagination
  const [filterQuery, setFilterQuery] = useState('');
  const [debouncedFilter] = useDebouncedValue(filterQuery, 200);
  const [page, setPage] = useState(1);

  // Expanded rows for inline simulation
  const [expandedRows, setExpandedRows] = useState<Set<string>>(new Set());
  const [rowData, setRowData] = useState<Map<string, ExpandedRowData>>(new Map());

  // Fetch metadata on mount
  useEffect(() => {
    yamsApi.accountNames()
      .then(setAccountNames)
      .catch((err) => console.error('Failed to fetch account names:', err));
    yamsApi.resourceAccounts()
      .then(setResourceAccounts)
      .catch((err) => console.error('Failed to fetch resource accounts:', err));
    yamsApi.actionAccessLevels()
      .then(setActionAccessLevels)
      .catch((err) => console.error('Failed to fetch action access levels:', err));
  }, []);

  // Use ref for context vars to avoid triggering effect
  const contextVarsRef = useRef(contextVars);
  contextVarsRef.current = contextVars;

  // Search functions
  const searchActions = useCallback((query: string) => yamsApi.searchActions(query), []);
  const searchResources = useCallback((query: string) => yamsApi.searchResources(query), []);

  // Run query when both action and resource are selected
  const runQuery = useCallback(async (): Promise<void> => {
    if (!selectedAction || !selectedResource) return;

    setLoading(true);
    setError(null);
    setResults([]);
    setExpandedRows(new Set());
    setRowData(new Map());
    setPage(1);

    try {
      const overlay = buildCombinedOverlay(selectedOverlayIds, loadedOverlays);
      const context = buildContext(contextVarsRef.current);

      const response = await yamsApi.whichPrincipals({
        action: selectedAction,
        resource: selectedResource,
        context,
        overlay,
      });

      setResults(response);
    } catch (err) {
      console.error('Query failed:', err);
      setError(err instanceof Error ? err.message : 'Query failed');
    } finally {
      setLoading(false);
    }
  }, [selectedAction, selectedResource, selectedOverlayIds, loadedOverlays]);

  // Stable reference for context vars in dependency array
  const contextVarsJson = JSON.stringify(contextVars);

  // Auto-run query when both selections are made
  useEffect(() => {
    if (selectedAction && selectedResource) {
      runQuery();
    } else {
      setResults([]);
      setError(null);
    }
  }, [selectedAction, selectedResource, selectedOverlayIds, loadedOverlays, contextVarsJson, runQuery]);

  // Filter and paginate results
  const filteredResults = useMemo(() => {
    if (!debouncedFilter) return results;
    const query = debouncedFilter.toLowerCase();
    return results.filter((p) => p.toLowerCase().includes(query));
  }, [results, debouncedFilter]);

  const totalPages = Math.ceil(filteredResults.length / RESULTS_PER_PAGE);

  const paginatedResults = useMemo(() => {
    const start = (page - 1) * RESULTS_PER_PAGE;
    return filteredResults.slice(start, start + RESULTS_PER_PAGE);
  }, [filteredResults, page]);

  // Reset page when filter changes
  useEffect(() => {
    setPage(1);
  }, [debouncedFilter]);

  // Toggle row expansion and run simulation
  const toggleRow = async (principal: string): Promise<void> => {
    const newExpanded = new Set(expandedRows);
    if (newExpanded.has(principal)) {
      newExpanded.delete(principal);
      setExpandedRows(newExpanded);
      return;
    }

    newExpanded.add(principal);
    setExpandedRows(newExpanded);

    // Skip if already loaded
    if (rowData.has(principal)) return;

    // Run simulation for this row
    setRowData((prev) => new Map(prev).set(principal, { loading: true, result: null, error: null }));

    try {
      const overlay = buildCombinedOverlay(selectedOverlayIds, loadedOverlays);
      const context = buildContext(contextVarsRef.current);

      const response = await yamsApi.simulate({
        principal,
        action: selectedAction!,
        resource: selectedResource || '*',
        context,
        overlay,
        explain: true,
      });

      setRowData((prev) => new Map(prev).set(principal, { loading: false, result: response, error: null }));
    } catch (err) {
      setRowData((prev) => new Map(prev).set(principal, {
        loading: false,
        result: null,
        error: err instanceof Error ? err.message : 'Simulation failed',
      }));
    }
  };

  const hasAnySelection = selectedAction || selectedResource || selectedOverlayIds.size > 0;
  const allSelected = selectedAction && selectedResource;

  const clearAll = (): void => {
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
            <Title order={3} mb={4}>Which Principals</Title>
            <Text size="sm" c="dimmed">
              Find which <Text component="span" fw={500} c="purple.6">principals</Text> can
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
            <Grid.Col span={6}>
              <AsyncSearchSelect
                label="Action (required)"
                placeholder="Search actions..."
                value={selectedAction}
                onChange={(v) => updateSelection('action', v)}
                onSearch={searchActions}
                accessLevels={actionAccessLevels}
                showAccessLevel
              />
            </Grid.Col>
            <Grid.Col span={6}>
              <AsyncSearchSelect
                label="Resource (required)"
                placeholder="Search resources..."
                value={selectedResource}
                onChange={(v) => updateSelection('resource', v)}
                onSearch={searchResources}
                formatLabel={formatResourceLabel}
                accountNames={accountNames}
                resourceAccounts={resourceAccounts}
                showAccountName
                showResourceType
              />
            </Grid.Col>
          </Grid>
        </Card>

        {/* Overlay selection */}
        <OverlaySelector
          selectedOverlayIds={selectedOverlayIds}
          onSelectionChange={setSelectedOverlayIds}
          loadedOverlays={loadedOverlays}
          onOverlaysLoaded={setLoadedOverlays}
        />

        {/* Request context */}
        <ContextEditor
          contextVars={contextVars}
          onChange={setContextVars}
          onRerun={runQuery}
          showRerunButton={!!allSelected}
        />

        {/* Results section */}
        {loading && (
          <Card withBorder p="lg">
            <Group justify="center" p="xl">
              <Loader size="md" />
              <Text c="dimmed">Finding principals...</Text>
            </Group>
          </Card>
        )}

        {error && (
          <Alert color="red" title="Error" icon={<IconX size={16} />}>
            {error}
          </Alert>
        )}

        {!loading && !error && results.length > 0 && (
          <Card withBorder p="lg">
            <Stack gap="md">
              <Group justify="space-between">
                <Title order={4}>
                  Results
                  <Text component="span" size="sm" c="dimmed" ml="xs">
                    ({filteredResults.length} principal{filteredResults.length !== 1 ? 's' : ''})
                  </Text>
                </Title>
                <TextInput
                  placeholder="Filter results..."
                  leftSection={<IconSearch size={14} />}
                  size="sm"
                  value={filterQuery}
                  onChange={(e) => setFilterQuery(e.currentTarget.value)}
                  style={{ width: 250 }}
                />
              </Group>

              <Table striped highlightOnHover>
                <Table.Thead>
                  <Table.Tr>
                    <Table.Th style={{ width: 50 }}></Table.Th>
                    <Table.Th>Principal</Table.Th>
                    <Table.Th style={{ width: 130 }}>Go to Simulation</Table.Th>
                  </Table.Tr>
                </Table.Thead>
                <Table.Tbody>
                  {paginatedResults.map((principal) => {
                    const isExpanded = expandedRows.has(principal);
                    const data = rowData.get(principal);
                    const accountId = extractAccountId(principal);
                    const accountName = accountId && accountNames[accountId];

                    return (
                      <>
                        <Table.Tr key={principal}>
                          <Table.Td>
                            <UnstyledButton onClick={() => toggleRow(principal)}>
                              <Group gap={2} wrap="nowrap">
                                {isExpanded ? (
                                  <IconChevronUp size={14} color="var(--mantine-color-dimmed)" />
                                ) : (
                                  <IconChevronDown size={14} color="var(--mantine-color-dimmed)" />
                                )}
                                <IconFlask size={14} color="var(--mantine-color-dimmed)" />
                              </Group>
                            </UnstyledButton>
                          </Table.Td>
                          <Table.Td>
                            <Anchor
                              component={Link}
                              to={`/search/principals?q=${encodeURIComponent(principal)}`}
                              size="sm"
                            >
                              {formatPrincipalLabel(principal)}
                            </Anchor>
                            {accountName && (
                              <Text size="xs" c="dimmed">
                                {accountName} ({accountId})
                              </Text>
                            )}
                            <Text size="xs" c="dimmed" truncate style={{ maxWidth: 500 }}>
                              {principal}
                            </Text>
                          </Table.Td>
                          <Table.Td>
                            <Anchor
                              component={Link}
                              to={buildAccessCheckUrl({
                                principal,
                                action: selectedAction || undefined,
                                resource: selectedResource || undefined,
                              })}
                              size="sm"
                            >
                              <Group gap={4} wrap="nowrap">
                                Open
                                <IconArrowUpRight size={14} />
                              </Group>
                            </Anchor>
                          </Table.Td>
                        </Table.Tr>
                        {isExpanded && (
                          <Table.Tr key={`${principal}-expanded`}>
                            <Table.Td colSpan={3} p={0}>
                              <Collapse in={isExpanded}>
                                <Box p="md" bg="gray.0">
                                  {data?.loading && (
                                    <Group gap="xs">
                                      <Loader size="sm" />
                                      <Text size="sm" c="dimmed">Running simulation...</Text>
                                    </Group>
                                  )}
                                  {data?.error && (
                                    <Text size="sm" c="red">{data.error}</Text>
                                  )}
                                  {data?.result && (
                                    <Stack gap="xs">
                                      <Group gap="xs">
                                        {data.result.result === 'ALLOW' ? (
                                          <Badge color="green" leftSection={<IconCheck size={12} />}>
                                            ALLOW
                                          </Badge>
                                        ) : (
                                          <Badge color="red" leftSection={<IconOctagonFilled size={12} />}>
                                            DENY
                                          </Badge>
                                        )}
                                      </Group>
                                      {data.result.explain && data.result.explain.length > 0 && (
                                        <Box>
                                          {data.result.explain.slice(0, 3).map((line, idx) => (
                                            <Text key={idx} size="xs" c="dimmed">{line}</Text>
                                          ))}
                                          {data.result.explain.length > 3 && (
                                            <Anchor
                                              component={Link}
                                              to={buildAccessCheckUrl({
                                                principal,
                                                action: selectedAction || undefined,
                                                resource: selectedResource || undefined,
                                              })}
                                              size="xs"
                                            >
                                              Open in Access Check →
                                            </Anchor>
                                          )}
                                        </Box>
                                      )}
                                    </Stack>
                                  )}
                                </Box>
                              </Collapse>
                            </Table.Td>
                          </Table.Tr>
                        )}
                      </>
                    );
                  })}
                </Table.Tbody>
              </Table>

              {totalPages > 1 && (
                <Group justify="space-between" align="center">
                  <Text size="xs" c="dimmed">
                    Showing {(page - 1) * RESULTS_PER_PAGE + 1}-{Math.min(page * RESULTS_PER_PAGE, filteredResults.length)} of {filteredResults.length}
                  </Text>
                  <Pagination
                    value={page}
                    onChange={setPage}
                    total={totalPages}
                    size="sm"
                  />
                </Group>
              )}
            </Stack>
          </Card>
        )}

        {!loading && !error && allSelected && results.length === 0 && (
          <Card withBorder p="xl">
            <Text ta="center" c="dimmed" size="lg">
              No principals found that can perform this action on this resource.
            </Text>
          </Card>
        )}

        {!allSelected && !loading && (
          <Card withBorder p="xl">
            <Text ta="center" c="dimmed" size="lg">
              Search and select an action and resource to find which principals can perform it.
            </Text>
          </Card>
        )}
      </Stack>
    </Box>
  );
}
