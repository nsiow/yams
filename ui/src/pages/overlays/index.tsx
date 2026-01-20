import { useCallback, useEffect, useMemo, useState } from 'react';
import { useNavigate, useParams, useSearchParams, Link } from 'react-router-dom';
import {
  ActionIcon,
  Anchor,
  Badge,
  Box,
  Breadcrumbs,
  Button,
  Card,
  Grid,
  Group,
  Modal,
  Pagination,
  ScrollArea,
  Stack,
  Table,
  Text,
  TextInput,
  Title,
  Tooltip,
} from '@mantine/core';
import { useDisclosure } from '@mantine/hooks';
import { useDebouncedValue } from '@mantine/hooks';
import { CodeHighlight } from '@mantine/code-highlight';
import {
  IconSearch,
  IconLayersLinked,
  IconChevronRight,
  IconPlus,
  IconTrash,
  IconEdit,
} from '@tabler/icons-react';
import { yamsApi } from '../../lib/api';
import type { OverlaySummary, OverlayData } from '../../lib/api';
import {
  CollapsibleCard,
  CopyButton,
  EmptyState,
  ExportButton,
  DetailSkeleton,
} from '../../components';

import '@mantine/code-highlight/styles.css';

function formatNumber(n: number): string {
  return n.toLocaleString();
}

function formatDate(dateStr: string): string {
  const date = new Date(dateStr);
  return date.toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });
}

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
  if (diffMins < 60) return `${diffMins} minute${diffMins === 1 ? '' : 's'} ago`;
  if (diffHours < 24) return `${diffHours} hour${diffHours === 1 ? '' : 's'} ago`;
  if (diffDays < 7) return `${diffDays} day${diffDays === 1 ? '' : 's'} ago`;
  if (diffWeeks < 5) return `${diffWeeks} week${diffWeeks === 1 ? '' : 's'} ago`;
  if (diffMonths < 12) return `${diffMonths} month${diffMonths === 1 ? '' : 's'} ago`;
  return `${diffYears} year${diffYears === 1 ? '' : 's'} ago`;
}

export function OverlaysPage(): JSX.Element {
  const navigate = useNavigate();
  const { '*': idFromUrl } = useParams();
  const [searchParams, setSearchParams] = useSearchParams();

  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [overlays, setOverlays] = useState<OverlaySummary[]>([]);
  const [selectedId, setSelectedId] = useState<string | null>(idFromUrl || null);
  const [selectedOverlay, setSelectedOverlay] = useState<OverlayData | null>(null);
  const [loadingDetail, setLoadingDetail] = useState(false);

  // Modal state for creating new overlay
  const [createModalOpened, { open: openCreateModal, close: closeCreateModal }] = useDisclosure(false);
  const [newOverlayName, setNewOverlayName] = useState('');
  const [creating, setCreating] = useState(false);

  // Delete confirmation modal
  const [deleteModalOpened, { open: openDeleteModal, close: closeDeleteModal }] = useDisclosure(false);
  const [deleting, setDeleting] = useState(false);
  const [overlayToDelete, setOverlayToDelete] = useState<OverlaySummary | null>(null);

  // Search
  const [searchQuery, setSearchQuery] = useState(searchParams.get('q') || '');
  const [debouncedSearch] = useDebouncedValue(searchQuery, 200);

  // Sync search to URL
  useEffect(() => {
    const params = new URLSearchParams();
    if (searchQuery) params.set('q', searchQuery);
    setSearchParams(params, { replace: true });
  }, [searchQuery, setSearchParams]);

  const filteredOverlays = useMemo(() => {
    if (!debouncedSearch) return overlays;
    const query = debouncedSearch.toLowerCase();
    return overlays.filter(o => o.name.toLowerCase().includes(query));
  }, [overlays, debouncedSearch]);

  const fetchOverlays = useCallback(async (): Promise<void> => {
    try {
      const list = await yamsApi.listOverlays();
      // Sort by createdAt descending (newest first)
      list.sort((a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime());
      setOverlays(list);
      setError(null);
    } catch (err) {
      console.error('Failed to fetch overlays:', err);
      setError(err instanceof Error ? err.message : 'Failed to fetch overlays');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchOverlays();
  }, [fetchOverlays]);

  const fetchOverlayDetail = useCallback(async (id: string): Promise<void> => {
    setLoadingDetail(true);
    try {
      const overlay = await yamsApi.getOverlay(id);
      setSelectedOverlay(overlay);
    } catch (err) {
      console.error('Failed to fetch overlay detail:', err);
      setSelectedOverlay(null);
    } finally {
      setLoadingDetail(false);
    }
  }, []);

  const handleSelectOverlay = (id: string): void => {
    setSelectedId(id);
    fetchOverlayDetail(id);
    navigate(`/overlays/${id}?${searchParams.toString()}`, { replace: true });
  };

  useEffect(() => {
    if (idFromUrl && idFromUrl !== selectedId) {
      setSelectedId(idFromUrl);
      fetchOverlayDetail(idFromUrl);
    }
  }, [idFromUrl, fetchOverlayDetail, selectedId]);

  const handleCreateOverlay = async (): Promise<void> => {
    if (!newOverlayName.trim()) return;
    setCreating(true);
    try {
      const created = await yamsApi.createOverlay({ name: newOverlayName.trim() });
      await fetchOverlays();
      handleSelectOverlay(created.id);
      closeCreateModal();
      setNewOverlayName('');
    } catch (err) {
      console.error('Failed to create overlay:', err);
    } finally {
      setCreating(false);
    }
  };

  const handleDeleteOverlay = async (): Promise<void> => {
    if (!overlayToDelete) return;
    setDeleting(true);
    try {
      await yamsApi.deleteOverlay(overlayToDelete.id);
      await fetchOverlays();
      // Clear selection if the deleted overlay was selected
      if (selectedId === overlayToDelete.id) {
        setSelectedId(null);
        setSelectedOverlay(null);
        navigate('/overlays', { replace: true });
      }
      closeDeleteModal();
      setOverlayToDelete(null);
    } catch (err) {
      console.error('Failed to delete overlay:', err);
    } finally {
      setDeleting(false);
    }
  };

  // Pagination
  const [page, setPage] = useState(1);
  const itemsPerPage = 15;
  const totalPages = Math.ceil(filteredOverlays.length / itemsPerPage);

  useEffect(() => {
    setPage(1);
  }, [debouncedSearch]);

  const paginatedOverlays = useMemo(() => {
    const start = (page - 1) * itemsPerPage;
    return filteredOverlays.slice(start, start + itemsPerPage);
  }, [filteredOverlays, page]);

  // Breadcrumb items
  const breadcrumbItems = [
    { title: 'Overlays', href: '/overlays' },
    ...(selectedOverlay ? [{ title: selectedOverlay.name, href: '' }] : []),
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
        {/* Left column - Table */}
        <Grid.Col span={6}>
          <Stack gap="md" h="100%">
            <Breadcrumbs separator={<IconChevronRight size={14} />}>{breadcrumbItems}</Breadcrumbs>
            <Group justify="space-between" align="center">
              <Title order={2}>Overlays</Title>
              <Button
                size="sm"
                leftSection={<IconPlus size={16} />}
                onClick={openCreateModal}
              >
                New Overlay
              </Button>
            </Group>

            <TextInput
              placeholder="Search overlays..."
              leftSection={<IconSearch size={16} />}
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.currentTarget.value)}
            />

            <Text size="sm" c="dimmed">
              {formatNumber(filteredOverlays.length)} overlay{filteredOverlays.length !== 1 ? 's' : ''}
              {totalPages > 1 && ` · page ${formatNumber(page)} of ${formatNumber(totalPages)}`}
            </Text>

            <Card withBorder p={0} style={{ flex: 1, overflow: 'hidden', display: 'flex', flexDirection: 'column' }}>
              {loading ? (
                <Box p="xl">
                  <Text c="dimmed" ta="center">Loading...</Text>
                </Box>
              ) : filteredOverlays.length === 0 ? (
                <Box p="xl">
                  <EmptyState
                    variant={searchQuery ? 'no-results' : 'no-data'}
                    entityName="overlay"
                  />
                </Box>
              ) : (
                <>
                  <ScrollArea style={{ flex: 1 }}>
                    <Table striped highlightOnHover>
                      <Table.Thead>
                        <Table.Tr>
                          <Table.Th>Name</Table.Th>
                          <Table.Th style={{ width: 70, textAlign: 'center' }}>Principals</Table.Th>
                          <Table.Th style={{ width: 70, textAlign: 'center' }}>Resources</Table.Th>
                          <Table.Th style={{ width: 70, textAlign: 'center' }}>Policies</Table.Th>
                          <Table.Th style={{ width: 110 }}>Created</Table.Th>
                          <Table.Th style={{ width: 50 }}></Table.Th>
                        </Table.Tr>
                      </Table.Thead>
                      <Table.Tbody>
                        {paginatedOverlays.map((o) => (
                          <Table.Tr
                            key={o.id}
                            style={{
                              cursor: 'pointer',
                              backgroundColor: selectedId === o.id ? 'var(--mantine-color-violet-0)' : undefined,
                            }}
                            onClick={() => handleSelectOverlay(o.id)}
                          >
                            <Table.Td>
                              <Group gap="xs" wrap="nowrap">
                                <IconLayersLinked size={16} color="var(--mantine-color-violet-6)" style={{ flexShrink: 0 }} />
                                <Text size="sm" fw={500} truncate style={{ maxWidth: 180 }}>{o.name}</Text>
                              </Group>
                            </Table.Td>
                            <Table.Td style={{ textAlign: 'center' }}>
                              <Text size="sm">{o.numPrincipals}</Text>
                            </Table.Td>
                            <Table.Td style={{ textAlign: 'center' }}>
                              <Text size="sm">{o.numResources}</Text>
                            </Table.Td>
                            <Table.Td style={{ textAlign: 'center' }}>
                              <Text size="sm">{o.numPolicies}</Text>
                            </Table.Td>
                            <Table.Td>
                              <Tooltip label={formatDate(o.createdAt)} openDelay={300}>
                                <Text size="xs" c="dimmed">{formatRelativeTime(o.createdAt)}</Text>
                              </Tooltip>
                            </Table.Td>
                            <Table.Td>
                              <Tooltip label="Delete">
                                <ActionIcon
                                  variant="subtle"
                                  color="red"
                                  size="sm"
                                  onClick={(e) => {
                                    e.stopPropagation();
                                    setOverlayToDelete(o);
                                    openDeleteModal();
                                  }}
                                >
                                  <IconTrash size={14} />
                                </ActionIcon>
                              </Tooltip>
                            </Table.Td>
                          </Table.Tr>
                        ))}
                      </Table.Tbody>
                    </Table>
                  </ScrollArea>
                  {totalPages > 1 && (
                    <Box p="xs" style={{ borderTop: '1px solid var(--mantine-color-default-border)' }}>
                      <Pagination value={page} onChange={setPage} total={totalPages} size="sm" withEdges />
                    </Box>
                  )}
                </>
              )}
            </Card>
          </Stack>
        </Grid.Col>

        {/* Right column - Details */}
        <Grid.Col span={6}>
          <Card withBorder h="100%" p="md">
            {!selectedId ? (
              <EmptyState variant="no-selection" entityName="overlay" />
            ) : loadingDetail ? (
              <DetailSkeleton />
            ) : selectedOverlay ? (
              <ScrollArea h="calc(100vh - 180px)">
                <Stack gap="md">
                  <Group justify="space-between" align="flex-start">
                    <Title order={3}>{selectedOverlay.name}</Title>
                    <Group gap="xs">
                      <Button
                        variant="light"
                        size="sm"
                        leftSection={<IconEdit size={16} />}
                        onClick={() => navigate(`/overlays/${selectedId}/edit`)}
                      >
                        Edit
                      </Button>
                      <ExportButton data={selectedOverlay} filename={`overlay-${selectedOverlay.name}`} />
                      <ActionIcon
                        variant="light"
                        color="red"
                        size="lg"
                        onClick={openDeleteModal}
                        title="Delete overlay"
                      >
                        <IconTrash size={18} />
                      </ActionIcon>
                    </Group>
                  </Group>

                  <CollapsibleCard title="Metadata">
                    <Stack gap="xs">
                      <Group gap="xs">
                        <Text size="sm" fw={600} c="dimmed" w={100}>ID:</Text>
                        <Text size="sm" ff="monospace" style={{ wordBreak: 'break-all', flex: 1 }}>
                          {selectedOverlay.id}
                        </Text>
                        <CopyButton value={selectedOverlay.id} />
                      </Group>
                      <Group gap="xs">
                        <Text size="sm" fw={600} c="dimmed" w={100}>Created:</Text>
                        <Text size="sm">{formatDate(selectedOverlay.createdAt)}</Text>
                        <Text size="xs" c="dimmed">({formatRelativeTime(selectedOverlay.createdAt)})</Text>
                      </Group>
                    </Stack>
                  </CollapsibleCard>

                  <CollapsibleCard title="Summary">
                    <Group gap="md">
                      <Badge variant="light" size="lg">
                        {selectedOverlay.principals?.length || 0} principals
                      </Badge>
                      <Badge variant="light" size="lg">
                        {selectedOverlay.resources?.length || 0} resources
                      </Badge>
                      <Badge variant="light" size="lg">
                        {selectedOverlay.policies?.length || 0} policies
                      </Badge>
                      <Badge variant="light" size="lg">
                        {selectedOverlay.accounts?.length || 0} accounts
                      </Badge>
                      <Badge variant="light" size="lg">
                        {selectedOverlay.groups?.length || 0} groups
                      </Badge>
                    </Group>
                  </CollapsibleCard>

                  {selectedOverlay.principals && selectedOverlay.principals.length > 0 && (
                    <CollapsibleCard title="Principals">
                      <Stack gap="xs">
                        {selectedOverlay.principals.map((p) => (
                          <Group key={p.Arn} gap="xs">
                            <Anchor component={Link} to={`/search/principals/${p.Arn}`} size="sm" ff="monospace" style={{ flex: 1 }}>
                              {p.Arn}
                            </Anchor>
                            <CopyButton value={p.Arn} />
                          </Group>
                        ))}
                      </Stack>
                    </CollapsibleCard>
                  )}

                  {selectedOverlay.resources && selectedOverlay.resources.length > 0 && (
                    <CollapsibleCard title="Resources">
                      <Stack gap="xs">
                        {selectedOverlay.resources.map((r) => (
                          <Group key={r.Arn} gap="xs">
                            <Anchor component={Link} to={`/search/resources/${r.Arn}`} size="sm" ff="monospace" style={{ flex: 1 }}>
                              {r.Arn}
                            </Anchor>
                            <CopyButton value={r.Arn} />
                          </Group>
                        ))}
                      </Stack>
                    </CollapsibleCard>
                  )}

                  {selectedOverlay.policies && selectedOverlay.policies.length > 0 && (
                    <CollapsibleCard title="Policies">
                      <Stack gap="xs">
                        {selectedOverlay.policies.map((p) => (
                          <Group key={p.Arn} gap="xs">
                            <Anchor component={Link} to={`/search/policies/${p.Arn}`} size="sm" ff="monospace" style={{ flex: 1 }}>
                              {p.Arn}
                            </Anchor>
                            <CopyButton value={p.Arn} />
                          </Group>
                        ))}
                      </Stack>
                    </CollapsibleCard>
                  )}

                  <CollapsibleCard title="Full Data">
                    <CodeHighlight
                      code={JSON.stringify(selectedOverlay, null, 2)}
                      language="json"
                      withCopyButton
                    />
                  </CollapsibleCard>
                </Stack>
              </ScrollArea>
            ) : (
              <EmptyState variant="error" message="Failed to load overlay details" />
            )}
          </Card>
        </Grid.Col>
      </Grid>

      {/* Create Modal */}
      <Modal opened={createModalOpened} onClose={closeCreateModal} title="Create New Overlay">
        <Stack gap="md">
          <TextInput
            label="Overlay Name"
            placeholder="Enter overlay name..."
            value={newOverlayName}
            onChange={(e) => setNewOverlayName(e.currentTarget.value)}
            data-autofocus
          />
          <Group justify="flex-end" gap="sm">
            <Button variant="default" onClick={closeCreateModal}>Cancel</Button>
            <Button
              onClick={handleCreateOverlay}
              loading={creating}
              disabled={!newOverlayName.trim()}
            >
              Create
            </Button>
          </Group>
        </Stack>
      </Modal>

      {/* Delete Modal */}
      <Modal opened={deleteModalOpened} onClose={closeDeleteModal} title="Delete Overlay">
        <Stack gap="md">
          <Text>Are you sure you want to delete "{overlayToDelete?.name}"? This action cannot be undone.</Text>
          <Group justify="flex-end" gap="sm">
            <Button variant="default" onClick={closeDeleteModal}>Cancel</Button>
            <Button color="red" onClick={handleDeleteOverlay} loading={deleting}>
              Delete
            </Button>
          </Group>
        </Stack>
      </Modal>
    </Box>
  );
}
