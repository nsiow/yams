// ui/src/pages/simulate/shared/overlay-selector.tsx
/* eslint-disable react-refresh/only-export-components */
import { useEffect, useMemo, useState } from 'react';
import {
  ActionIcon,
  Badge,
  Button,
  Card,
  Checkbox,
  Group,
  Pagination,
  Stack,
  Table,
  Text,
  TextInput,
  Title,
  Tooltip,
} from '@mantine/core';
import { useDebouncedValue } from '@mantine/hooks';
import { IconCheck, IconLayersLinked, IconPlus, IconSearch, IconX } from '@tabler/icons-react';
import { yamsApi } from '../../../lib/api';
import type { OverlaySummary, OverlayData, SimulationOverlay } from '../../../lib/api';
import { formatRelativeTime } from './utils';

interface OverlaySelectorProps {
  selectedOverlayIds: Set<string>;
  onSelectionChange: (ids: Set<string>) => void;
  loadedOverlays: Map<string, OverlayData>;
  onOverlaysLoaded: (overlays: Map<string, OverlayData>) => void;
}

const OVERLAYS_PER_PAGE = 5;

export function OverlaySelector({
  selectedOverlayIds,
  onSelectionChange,
  loadedOverlays,
  onOverlaysLoaded,
}: OverlaySelectorProps): JSX.Element {
  const [overlays, setOverlays] = useState<OverlaySummary[]>([]);
  const [showOverlaySelector, setShowOverlaySelector] = useState(false);
  const [overlaySearchQuery, setOverlaySearchQuery] = useState('');
  const [overlayPage, setOverlayPage] = useState(1);

  // Fetch overlays on mount
  useEffect(() => {
    yamsApi.listOverlays()
      .then((list) => {
        list.sort((a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime());
        setOverlays(list);
      })
      .catch((err) => console.error('Failed to fetch overlays:', err));
  }, []);

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
      onOverlaysLoaded(loaded);
    };
    loadMissingOverlays();
  }, [selectedOverlayIds, loadedOverlays, onOverlaysLoaded]);

  // Overlay filtering and pagination
  const [debouncedOverlaySearch] = useDebouncedValue(overlaySearchQuery, 200);

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

  const toggleOverlay = (id: string): void => {
    const next = new Set(selectedOverlayIds);
    if (next.has(id)) {
      next.delete(id);
    } else {
      next.add(id);
    }
    onSelectionChange(next);
  };

  return (
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

      {/* Empty state */}
      {!showOverlaySelector && selectedOverlayIds.size === 0 && (
        <Text size="sm" c="dimmed">
          {overlays.length === 0
            ? 'No overlays available. Create overlays to test against hypothetical environments.'
            : 'Add overlays to test against hypothetical environments.'}
        </Text>
      )}
    </Card>
  );
}

// Helper to build combined overlay from selected overlays
export function buildCombinedOverlay(
  selectedOverlayIds: Set<string>,
  loadedOverlays: Map<string, OverlayData>
): SimulationOverlay | undefined {
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

  const hasData =
    combined.accounts!.length > 0 ||
    combined.groups!.length > 0 ||
    combined.policies!.length > 0 ||
    combined.principals!.length > 0 ||
    combined.resources!.length > 0;

  return hasData ? combined : undefined;
}
