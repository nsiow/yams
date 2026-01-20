// ui/src/pages/simulate/which-resources.tsx
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
  Tooltip,
  UnstyledButton,
} from '@mantine/core';
import { useDebouncedValue } from '@mantine/hooks';
import {
  IconCheck,
  IconChevronDown,
  IconChevronRight,
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
  extractService,
  buildAccessCheckUrl,
} from './shared';
import type { ContextVariable } from './shared';

const RESULTS_PER_PAGE = 20;

interface ExpandedRowData {
  loading: boolean;
  result: SimulationResponse | null;
  error: string | null;
}

export function WhichResourcesPage(): JSX.Element {
  const [searchParams, setSearchParams] = useSearchParams();

  // Initialize state from URL params
  const [selectedPrincipal, setSelectedPrincipal] = useState<string | null>(
    searchParams.get('principal')
  );
  const [selectedAction, setSelectedAction] = useState<string | null>(
    searchParams.get('action')
  );

  // Update URL when selections change
  const updateSelection = useCallback(
    (key: 'principal' | 'action', value: string | null): void => {
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
  const searchPrincipals = useCallback((query: string) => yamsApi.searchPrincipals(query), []);
  const searchActions = useCallback((query: string) => yamsApi.searchActions(query), []);

  // Run query when principal is selected
  const runQuery = useCallback(async (): Promise<void> => {
    if (!selectedPrincipal) return;

    setLoading(true);
    setError(null);
    setResults([]);
    setExpandedRows(new Set());
    setRowData(new Map());
    setPage(1);

    try {
      const overlay = buildCombinedOverlay(selectedOverlayIds, loadedOverlays);
      const context = buildContext(contextVarsRef.current);

      const response = await yamsApi.whichResources({
        principal: selectedPrincipal,
        action: selectedAction || undefined,
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
  }, [selectedPrincipal, selectedAction, selectedOverlayIds, loadedOverlays]);

  // Stable reference for context vars in dependency array
  const contextVarsJson = JSON.stringify(contextVars);

  // Auto-run query when selections or options change
  useEffect(() => {
    if (selectedPrincipal) {
      runQuery();
    } else {
      setResults([]);
      setError(null);
    }
  }, [selectedPrincipal, selectedAction, selectedOverlayIds, loadedOverlays, contextVarsJson, runQuery]);

  // Filter and paginate results
  const filteredResults = useMemo(() => {
    if (!debouncedFilter) return results;
    const query = debouncedFilter.toLowerCase();
    return results.filter((r) => r.toLowerCase().includes(query));
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
  const toggleRow = async (resource: string): Promise<void> => {
    const newExpanded = new Set(expandedRows);
    if (newExpanded.has(resource)) {
      newExpanded.delete(resource);
      setExpandedRows(newExpanded);
      return;
    }

    newExpanded.add(resource);
    setExpandedRows(newExpanded);

    // Skip if already loaded or no action selected
    if (rowData.has(resource) || !selectedAction) return;

    // Run simulation for this row
    setRowData((prev) => new Map(prev).set(resource, { loading: true, result: null, error: null }));

    try {
      const overlay = buildCombinedOverlay(selectedOverlayIds, loadedOverlays);
      const context = buildContext(contextVarsRef.current);

      const response = await yamsApi.simulate({
        principal: selectedPrincipal!,
        action: selectedAction,
        resource,
        context,
        overlay,
        explain: true,
      });

      setRowData((prev) => new Map(prev).set(resource, { loading: false, result: response, error: null }));
    } catch (err) {
      setRowData((prev) => new Map(prev).set(resource, {
        loading: false,
        result: null,
        error: err instanceof Error ? err.message : 'Simulation failed',
      }));
    }
  };

  const hasAnySelection = selectedPrincipal || selectedAction || selectedOverlayIds.size > 0;

  const clearAll = (): void => {
    updateSelection('principal', null);
    updateSelection('action', null);
    setContextVars([]);
    setSelectedOverlayIds(new Set());
  };

  return (
    <Box p="md">
      <Stack gap="lg">
        {/* Page header */}
        <Group justify="space-between" align="center">
          <Box>
            <Title order={3} mb={4}>Which Resources</Title>
            <Text size="sm" c="dimmed">
              Find which <Text component="span" fw={500} c="purple.6">resources</Text> a
              <Text component="span" fw={500} c="purple.6"> principal</Text> can
              access{selectedAction && <> with a specific <Text component="span" fw={500} c="purple.6">action</Text></>}.
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
                label="Principal (required)"
                placeholder="Search principals..."
                value={selectedPrincipal}
                onChange={(v) => updateSelection('principal', v)}
                onSearch={searchPrincipals}
                formatLabel={formatPrincipalLabel}
                accountNames={accountNames}
                showAccountName
              />
            </Grid.Col>
            <Grid.Col span={6}>
              <AsyncSearchSelect
                label="Action (optional)"
                placeholder="Search actions..."
                value={selectedAction}
                onChange={(v) => updateSelection('action', v)}
                onSearch={searchActions}
                accessLevels={actionAccessLevels}
                showAccessLevel
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
          showRerunButton={!!selectedPrincipal}
        />

        {/* Results section */}
        {loading && (
          <Card withBorder p="lg">
            <Group justify="center" p="xl">
              <Loader size="md" />
              <Text c="dimmed">Finding resources...</Text>
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
                    ({filteredResults.length} resource{filteredResults.length !== 1 ? 's' : ''})
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
                    <Table.Th style={{ width: 40 }}></Table.Th>
                    <Table.Th>Resource</Table.Th>
                    <Table.Th style={{ width: 100 }}>Service</Table.Th>
                    <Table.Th style={{ width: 130 }}>Account</Table.Th>
                    <Table.Th style={{ width: 70 }}>Simulate</Table.Th>
                  </Table.Tr>
                </Table.Thead>
                <Table.Tbody>
                  {paginatedResults.map((resource) => {
                    const isExpanded = expandedRows.has(resource);
                    const data = rowData.get(resource);
                    const accountId = extractAccountId(resource) || resourceAccounts[resource];
                    const accountName = accountId && accountNames[accountId];
                    const service = extractService(resource);

                    return (
                      <>
                        <Table.Tr key={resource}>
                          <Table.Td>
                            <Tooltip label={isExpanded ? 'Hide result' : 'Show simulation result'} openDelay={300}>
                              <UnstyledButton onClick={() => toggleRow(resource)}>
                                {isExpanded ? (
                                  <IconChevronDown size={16} color="var(--mantine-color-dimmed)" />
                                ) : (
                                  <IconChevronRight size={16} color="var(--mantine-color-dimmed)" />
                                )}
                              </UnstyledButton>
                            </Tooltip>
                          </Table.Td>
                          <Table.Td>
                            <Anchor
                              component={Link}
                              to={`/search/resources?q=${encodeURIComponent(resource)}`}
                              size="sm"
                            >
                              {formatResourceLabel(resource)}
                            </Anchor>
                            <Text size="xs" c="dimmed" truncate style={{ maxWidth: 350 }}>
                              {resource}
                            </Text>
                          </Table.Td>
                          <Table.Td>
                            {service ? (
                              <Badge size="sm" variant="light" color="gray">
                                {service}
                              </Badge>
                            ) : (
                              <Text size="xs" c="dimmed">-</Text>
                            )}
                          </Table.Td>
                          <Table.Td>
                            {accountName ? (
                              <Tooltip label={accountId}>
                                <Text size="sm" truncate style={{ maxWidth: 110 }}>
                                  {accountName}
                                </Text>
                              </Tooltip>
                            ) : (
                              <Text size="sm" c="dimmed" ff="monospace">
                                {accountId || '-'}
                              </Text>
                            )}
                          </Table.Td>
                          <Table.Td>
                            <Tooltip label="Simulate in Access Check">
                              <Anchor
                                component={Link}
                                to={buildAccessCheckUrl({
                                  principal: selectedPrincipal || undefined,
                                  action: selectedAction || undefined,
                                  resource,
                                })}
                              >
                                <IconFlask size={18} />
                              </Anchor>
                            </Tooltip>
                          </Table.Td>
                        </Table.Tr>
                        {isExpanded && (
                          <Table.Tr key={`${resource}-expanded`}>
                            <Table.Td colSpan={5} p={0}>
                              <Collapse in={isExpanded}>
                                <Box p="md" bg="gray.0">
                                  {!selectedAction && (
                                    <Text size="sm" c="dimmed">
                                      Select an action to run a simulation for this resource.
                                    </Text>
                                  )}
                                  {selectedAction && data?.loading && (
                                    <Group gap="xs">
                                      <Loader size="sm" />
                                      <Text size="sm" c="dimmed">Running simulation...</Text>
                                    </Group>
                                  )}
                                  {selectedAction && data?.error && (
                                    <Text size="sm" c="red">{data.error}</Text>
                                  )}
                                  {selectedAction && data?.result && (
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
                                                principal: selectedPrincipal || undefined,
                                                action: selectedAction || undefined,
                                                resource,
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

        {!loading && !error && selectedPrincipal && results.length === 0 && (
          <Card withBorder p="xl">
            <Text ta="center" c="dimmed" size="lg">
              No resources found that this principal can access
              {selectedAction ? ' with this action' : ''}.
            </Text>
          </Card>
        )}

        {!selectedPrincipal && !loading && (
          <Card withBorder p="xl">
            <Text ta="center" c="dimmed" size="lg">
              Search and select a principal to find which resources they can access.
            </Text>
          </Card>
        )}
      </Stack>
    </Box>
  );
}
