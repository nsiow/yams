import { useCallback, useEffect, useMemo, useState } from 'react';
import { useNavigate, useParams, useSearchParams, Link } from 'react-router-dom';
import {
  Anchor,
  Badge,
  Box,
  Breadcrumbs,
  Card,
  Grid,
  Group,
  Pagination,
  ScrollArea,
  Select,
  Stack,
  Table,
  Text,
  TextInput,
  Title,
} from '@mantine/core';
import { useDebouncedValue } from '@mantine/hooks';
import { CodeHighlight } from '@mantine/code-highlight';
import { IconSearch, IconDatabase, IconChevronRight } from '@tabler/icons-react';
import { yamsApi } from '../../lib/api';
import type { Resource } from '../../lib/api';
import {
  CollapsibleCard,
  CopyButton,
  CopyEntityButton,
  EmptyState,
  ExportButton,
  FilterBar,
  ListSkeleton,
  DetailSkeleton,
} from '../../components';

import '@mantine/code-highlight/styles.css';

interface ResourceListItem {
  arn: string;
  service: string;
  region: string;
  accountId: string;
  resourceType: string;
  name: string;
}

function formatNumber(n: number): string {
  return n.toLocaleString();
}

function parseResourceArn(arn: string): ResourceListItem {
  // ARN format: arn:aws:service:region:account:resource-type/resource-name
  // or: arn:aws:service:region:account:resource-type:resource-name
  const parts = arn.split(':');
  const service = parts[2] || '';
  const region = parts[3] || '';
  const accountId = parts[4] || '';

  // Resource part can be type/name, type:name, or just name
  const resourceParts = parts.slice(5).join(':');
  let resourceType = '';
  let name = resourceParts;

  if (resourceParts.includes('/')) {
    const idx = resourceParts.indexOf('/');
    resourceType = resourceParts.substring(0, idx);
    name = resourceParts.substring(idx + 1);
  } else if (resourceParts.includes(':')) {
    const idx = resourceParts.indexOf(':');
    resourceType = resourceParts.substring(0, idx);
    name = resourceParts.substring(idx + 1);
  }

  return { arn, service, region, accountId, resourceType, name };
}

// Format service:resourceType for display (lowercase)
function formatResourceType(item: ResourceListItem): string {
  if (item.resourceType) {
    return `${item.service}:${item.resourceType}`.toLowerCase();
  }
  return item.service.toLowerCase();
}

// Convert AWS CloudFormation type (AWS::S3::Bucket) to service:type format
function formatCloudFormationType(cfnType: string): string {
  // AWS::Service::ResourceType -> service:resourcetype
  const parts = cfnType.split('::');
  if (parts.length >= 3) {
    return `${parts[1]}:${parts[2]}`.toLowerCase();
  }
  return cfnType.toLowerCase();
}

export function ResourcesPage(): JSX.Element {
  const navigate = useNavigate();
  const { '*': arnFromUrl } = useParams();
  const [searchParams, setSearchParams] = useSearchParams();

  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [resourceArns, setResourceArns] = useState<string[]>([]);
  const [accountNames, setAccountNames] = useState<Record<string, string>>({});
  const [selectedArn, setSelectedArn] = useState<string | null>(arnFromUrl || null);
  const [selectedResource, setSelectedResource] = useState<Resource | null>(null);
  const [loadingDetail, setLoadingDetail] = useState(false);

  // Initialize filters from URL params
  const [searchQuery, setSearchQuery] = useState(searchParams.get('q') || '');
  const [debouncedSearch] = useDebouncedValue(searchQuery, 200);
  const [typeFilter, setTypeFilter] = useState<string | null>(searchParams.get('type'));
  const [regionFilter, setRegionFilter] = useState<string | null>(searchParams.get('region'));
  const [accountFilter, setAccountFilter] = useState<string | null>(searchParams.get('account'));

  // Sync filters to URL
  useEffect(() => {
    const params = new URLSearchParams();
    if (searchQuery) params.set('q', searchQuery);
    if (typeFilter) params.set('type', typeFilter);
    if (regionFilter) params.set('region', regionFilter);
    if (accountFilter) params.set('account', accountFilter);
    setSearchParams(params, { replace: true });
  }, [searchQuery, typeFilter, regionFilter, accountFilter, setSearchParams]);

  const hasActiveFilters = Boolean(searchQuery || typeFilter || regionFilter || accountFilter);

  const clearAllFilters = (): void => {
    setSearchQuery('');
    setTypeFilter(null);
    setRegionFilter(null);
    setAccountFilter(null);
  };

  // Parse all resource ARNs into list items
  const resourceList = useMemo(() => {
    return resourceArns.map(parseResourceArn);
  }, [resourceArns]);

  // Extract unique types for filter dropdown
  const typeOptions = useMemo(() => {
    const types = new Set(resourceList.map(formatResourceType));
    return Array.from(types).sort().map(t => ({ value: t, label: t }));
  }, [resourceList]);

  // Extract unique regions for filter dropdown
  const regionOptions = useMemo(() => {
    const regions = new Set(resourceList.map(r => r.region).filter(Boolean));
    return Array.from(regions).sort().map(r => ({ value: r, label: r }));
  }, [resourceList]);

  // Extract unique accounts for filter dropdown
  const accountOptions = useMemo(() => {
    const ids = new Set(resourceList.map(r => r.accountId).filter(Boolean));
    return Array.from(ids)
      .sort()
      .map(id => {
        const name = accountNames[id];
        const label = name ? `${name} (${id})` : id;
        return { value: id, label };
      });
  }, [resourceList, accountNames]);

  // Helper to format account display
  const formatAccount = (accountId: string): string => {
    if (!accountId) return 'N/A';
    const name = accountNames[accountId];
    return name ? `${name} (${accountId})` : accountId;
  };

  // Filter resources based on search and filters
  const filteredResources = useMemo(() => {
    return resourceList.filter(r => {
      if (typeFilter && formatResourceType(r) !== typeFilter) return false;
      if (regionFilter && r.region !== regionFilter) return false;
      if (accountFilter && r.accountId !== accountFilter) return false;
      if (debouncedSearch) {
        const query = debouncedSearch.toLowerCase();
        return r.name.toLowerCase().includes(query) || r.arn.toLowerCase().includes(query);
      }
      return true;
    });
  }, [resourceList, typeFilter, regionFilter, accountFilter, debouncedSearch]);

  // Fetch all resource ARNs and account names on mount
  useEffect(() => {
    async function fetchData(): Promise<void> {
      try {
        const [arns, names] = await Promise.all([
          yamsApi.listResources(),
          yamsApi.accountNames(),
        ]);
        setResourceArns(arns);
        setAccountNames(names);
        setError(null);
      } catch (err) {
        console.error('Failed to fetch data:', err);
        setError(err instanceof Error ? err.message : 'Failed to fetch data');
      } finally {
        setLoading(false);
      }
    }
    fetchData();
  }, []);

  // Fetch resource detail when selected
  const fetchResourceDetail = useCallback(async (arn: string): Promise<void> => {
    setLoadingDetail(true);
    try {
      const resource = await yamsApi.getResource(arn);
      setSelectedResource(resource);
    } catch (err) {
      console.error('Failed to fetch resource detail:', err);
      setSelectedResource(null);
    } finally {
      setLoadingDetail(false);
    }
  }, []);

  const handleSelectResource = (arn: string): void => {
    setSelectedArn(arn);
    fetchResourceDetail(arn);
    navigate(`/search/resources/${arn}?${searchParams.toString()}`, { replace: true });
  };

  // Load resource from URL on mount or when URL changes
  useEffect(() => {
    if (arnFromUrl && arnFromUrl !== selectedArn) {
      setSelectedArn(arnFromUrl);
      fetchResourceDetail(arnFromUrl);
    }
  }, [arnFromUrl, fetchResourceDetail, selectedArn]);

  // Pagination
  const [page, setPage] = useState(1);
  const itemsPerPage = 20;
  const totalPages = Math.ceil(filteredResources.length / itemsPerPage);

  // Reset to page 1 when filters change
  useEffect(() => {
    setPage(1);
  }, [debouncedSearch, typeFilter, regionFilter, accountFilter]);

  const paginatedResources = useMemo(() => {
    const start = (page - 1) * itemsPerPage;
    return filteredResources.slice(start, start + itemsPerPage);
  }, [filteredResources, page]);

  // Breadcrumb items
  const breadcrumbItems = [
    { title: 'Search', href: '/' },
    { title: 'Resources', href: '/search/resources' },
    ...(selectedResource ? [{ title: selectedResource.Name, href: '' }] : []),
  ].map((item, index, arr) => {
    const isLast = index === arr.length - 1;
    if (isLast) {
      return <Text key={item.title} size="sm" c="dimmed">{item.title}</Text>;
    }
    return (
      <Anchor key={item.title} component={Link} to={item.href} size="sm">
        {item.title}
      </Anchor>
    );
  });

  if (error) {
    return (
      <Box p="md">
        <EmptyState variant="error" message={error} />
      </Box>
    );
  }

  return (
    <Box p="md" h="100%">
      <Grid gutter="md" h="100%">
        {/* Left column */}
        <Grid.Col span={5}>
          <Stack gap="md" h="100%">
            <Breadcrumbs separator={<IconChevronRight size={14} />}>{breadcrumbItems}</Breadcrumbs>
            <Title order={2}>Resources</Title>

            <TextInput
              placeholder="Search resources..."
              leftSection={<IconSearch size={16} />}
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.currentTarget.value)}
            />

            <FilterBar hasActiveFilters={hasActiveFilters} onClearAll={clearAllFilters}>
              <Select
                placeholder="All types"
                size="sm"
                data={typeOptions}
                value={typeFilter}
                onChange={setTypeFilter}
                clearable
                searchable
                style={{ flex: 1 }}
              />
              <Select
                placeholder="All regions"
                size="sm"
                data={regionOptions}
                value={regionFilter}
                onChange={setRegionFilter}
                clearable
                searchable
                style={{ flex: 1 }}
              />
            </FilterBar>
            <Select
              placeholder="All accounts"
              size="sm"
              data={accountOptions}
              value={accountFilter}
              onChange={setAccountFilter}
              clearable
              searchable
            />

            <Text size="sm" c="dimmed">
              {formatNumber(filteredResources.length)} of {formatNumber(resourceList.length)} resources
              {totalPages > 1 && ` (page ${formatNumber(page)} of ${formatNumber(totalPages)})`}
            </Text>

            <Card withBorder p={0} style={{ flex: 1, overflow: 'hidden', display: 'flex', flexDirection: 'column' }}>
              <ScrollArea style={{ flex: 1 }}>
                {loading ? (
                  <ListSkeleton />
                ) : filteredResources.length === 0 ? (
                  <Box p="xl">
                    <EmptyState
                      variant={hasActiveFilters ? 'no-results' : 'no-data'}
                      entityName="resource"
                    />
                  </Box>
                ) : (
                  paginatedResources.map((r) => (
                    <div
                      key={r.arn}
                      onClick={() => handleSelectResource(r.arn)}
                      style={{
                        cursor: 'pointer',
                        padding: '8px 12px',
                        display: 'flex',
                        alignItems: 'center',
                        gap: '12px',
                        backgroundColor: selectedArn === r.arn
                          ? 'var(--mantine-color-primary-light)'
                          : undefined,
                        borderBottom: '1px solid var(--mantine-color-default-border)',
                      }}
                      onMouseEnter={(e) => {
                        if (selectedArn !== r.arn) {
                          e.currentTarget.style.backgroundColor = 'var(--mantine-color-gray-light-hover)';
                        }
                      }}
                      onMouseLeave={(e) => {
                        if (selectedArn !== r.arn) {
                          e.currentTarget.style.backgroundColor = '';
                        }
                      }}
                    >
                      <div style={{ flexShrink: 0 }}>
                        <IconDatabase size={16} color="var(--mantine-color-teal-6)" />
                      </div>
                      <div style={{ minWidth: 0, flex: 1 }}>
                        <Text size="sm" fw={500} truncate>{r.name}</Text>
                        <Text size="xs" c="dimmed" truncate>
                          {formatResourceType(r)} &middot; {r.region || 'global'}
                        </Text>
                      </div>
                    </div>
                  ))
                )}
              </ScrollArea>
              {totalPages > 1 && (
                <Box p="xs" style={{ borderTop: '1px solid var(--mantine-color-default-border)' }}>
                  <Pagination value={page} onChange={setPage} total={totalPages} size="sm" withEdges />
                </Box>
              )}
            </Card>
          </Stack>
        </Grid.Col>

        {/* Right column */}
        <Grid.Col span={7}>
          <Card withBorder h="100%" p="md">
            {!selectedArn ? (
              <EmptyState variant="no-selection" entityName="resource" />
            ) : loadingDetail ? (
              <DetailSkeleton />
            ) : selectedResource ? (
              <ScrollArea h="calc(100vh - 180px)">
                <Stack gap="md">
                  <Group justify="space-between" align="flex-start">
                    <Title order={3}>{selectedResource.Name}</Title>
                    <Group gap="xs">
                      <CopyEntityButton data={selectedResource} />
                      <ExportButton data={selectedResource} filename={`resource-${selectedResource.Name}`} />
                      <Badge color="teal" size="lg">
                        {formatCloudFormationType(selectedResource.Type)}
                      </Badge>
                    </Group>
                  </Group>

                  <CollapsibleCard title="Metadata">
                    <Stack gap="xs">
                      <Group gap="xs">
                        <Text size="sm" fw={600} c="dimmed" w={100}>ARN:</Text>
                        <Text size="sm" ff="monospace" style={{ wordBreak: 'break-all', flex: 1 }}>
                          {selectedResource.Arn}
                        </Text>
                        <CopyButton value={selectedResource.Arn} />
                      </Group>
                      <Group gap="xs">
                        <Text size="sm" fw={600} c="dimmed" w={100}>Account:</Text>
                        <Anchor component={Link} to={`/search/accounts/${selectedResource.AccountId}`} size="sm" ff="monospace">
                          {formatAccount(selectedResource.AccountId)}
                        </Anchor>
                        <CopyButton value={selectedResource.AccountId} />
                      </Group>
                      <Group gap="xs">
                        <Text size="sm" fw={600} c="dimmed" w={100}>Region:</Text>
                        <Text size="sm" ff="monospace">{selectedResource.Region || 'global'}</Text>
                      </Group>
                      <Group gap="xs">
                        <Text size="sm" fw={600} c="dimmed" w={100}>Type:</Text>
                        <Text size="sm" ff="monospace">{formatCloudFormationType(selectedResource.Type)}</Text>
                      </Group>
                    </Stack>
                  </CollapsibleCard>

                  {selectedResource.Tags && selectedResource.Tags.length > 0 && (
                    <CollapsibleCard title="Tags">
                      <Table withRowBorders={false}>
                        <Table.Tbody>
                          {selectedResource.Tags.map((tag) => (
                            <Table.Tr key={tag.Key}>
                              <Table.Td w={200} style={{ verticalAlign: 'top' }}>
                                <Text size="sm" fw={500} c="dimmed">{tag.Key}</Text>
                              </Table.Td>
                              <Table.Td>
                                <Text size="sm" style={{ wordBreak: 'break-word' }}>{tag.Value}</Text>
                              </Table.Td>
                            </Table.Tr>
                          ))}
                        </Table.Tbody>
                      </Table>
                    </CollapsibleCard>
                  )}

                  {selectedResource.Policy && selectedResource.Policy.Statement && selectedResource.Policy.Statement.length > 0 && (
                    <CollapsibleCard title="Resource Policy">
                      <CodeHighlight
                        code={JSON.stringify(selectedResource.Policy, null, 2)}
                        language="json"
                        withCopyButton
                      />
                    </CollapsibleCard>
                  )}
                </Stack>
              </ScrollArea>
            ) : (
              <EmptyState variant="error" message="Failed to load resource details" />
            )}
          </Card>
        </Grid.Col>
      </Grid>
    </Box>
  );
}
