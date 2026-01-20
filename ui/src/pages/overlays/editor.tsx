import { useCallback, useEffect, useState } from 'react';
import { useParams, Link, useNavigate } from 'react-router-dom';
import {
  ActionIcon,
  Alert,
  Anchor,
  Box,
  Breadcrumbs,
  Button,
  Card,
  Checkbox,
  Combobox,
  Grid,
  Group,
  Input,
  InputBase,
  Modal,
  ScrollArea,
  Stack,
  Tabs,
  Text,
  TextInput,
  useCombobox,
} from '@mantine/core';
import { useDebouncedValue } from '@mantine/hooks';
import { useDisclosure } from '@mantine/hooks';
import {
  IconAlertCircle,
  IconChevronRight,
  IconPlus,
  IconDeviceFloppy,
  IconUser,
  IconDatabase,
  IconShield,
  IconBuilding,
  IconSearch,
  IconX,
} from '@tabler/icons-react';
import { yamsApi } from '../../lib/api';
import type { OverlayData, OverlaySummary, Principal, Resource, Policy, Account } from '../../lib/api';
import { EmptyState, DetailSkeleton } from '../../components';
import { EntityListItem, EntityDetailPanel } from './components';

// Entity type union
type EntityType = 'principal' | 'resource' | 'policy' | 'account';
type EntityUnion = Principal | Resource | Policy | Account;

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

// Format entity counts in a readable way
function formatEntityCounts(overlay: OverlaySummary): string {
  const parts: string[] = [];
  if (overlay.numPrincipals > 0) {
    parts.push(`${overlay.numPrincipals} principal${overlay.numPrincipals !== 1 ? 's' : ''}`);
  }
  if (overlay.numResources > 0) {
    parts.push(`${overlay.numResources} resource${overlay.numResources !== 1 ? 's' : ''}`);
  }
  if (overlay.numPolicies > 0) {
    parts.push(`${overlay.numPolicies} ${overlay.numPolicies !== 1 ? 'policies' : 'policy'}`);
  }
  if (overlay.numAccounts > 0) {
    parts.push(`${overlay.numAccounts} account${overlay.numAccounts !== 1 ? 's' : ''}`);
  }
  return parts.length > 0 ? parts.join(', ') : 'empty';
}

// Extract account ID from ARN
function extractAccountId(arn: string): string | null {
  const parts = arn.split(':');
  if (parts.length >= 5 && parts[4]) {
    return parts[4];
  }
  return null;
}

// Extract service type from ARN
function extractService(arn: string): string | null {
  const parts = arn.split(':');
  if (parts.length >= 3 && parts[2]) {
    return parts[2];
  }
  return null;
}

// Format display name from ARN
function formatArnLabel(arn: string): string {
  const parts = arn.split('/');
  return parts[parts.length - 1] || arn;
}

// Add entity modal for searching and adding entities
interface AddEntityModalProps {
  opened: boolean;
  onClose: () => void;
  entityType: EntityType;
  onAdd: (arns: string[]) => void;
  existingArns: string[];
}

function AddEntityModal({ opened, onClose, entityType, onAdd, existingArns }: AddEntityModalProps): JSX.Element {
  const [searchQuery, setSearchQuery] = useState('');
  const [results, setResults] = useState<string[]>([]);
  const [loading, setLoading] = useState(false);
  const [selectedArns, setSelectedArns] = useState<Set<string>>(new Set());
  const [accountNames, setAccountNames] = useState<Record<string, string>>({});
  const [resourceAccounts, setResourceAccounts] = useState<Record<string, string>>({});

  // Fetch account info on mount
  useEffect(() => {
    yamsApi.accountNames()
      .then(setAccountNames)
      .catch((err) => console.error('Failed to fetch account names:', err));
    yamsApi.resourceAccounts()
      .then(setResourceAccounts)
      .catch((err) => console.error('Failed to fetch resource accounts:', err));
  }, []);

  useEffect(() => {
    if (!opened) {
      setSearchQuery('');
      setResults([]);
      setSelectedArns(new Set());
      return;
    }

    async function fetchResults(): Promise<void> {
      setLoading(true);
      try {
        let data: string[] = [];
        switch (entityType) {
          case 'principal':
            data = searchQuery
              ? await yamsApi.searchPrincipals(searchQuery)
              : await yamsApi.listPrincipals();
            break;
          case 'resource':
            data = searchQuery
              ? await yamsApi.searchResources(searchQuery)
              : await yamsApi.listResources();
            break;
          case 'policy':
            data = searchQuery
              ? await yamsApi.searchPolicies(searchQuery)
              : await yamsApi.listPolicies();
            break;
          case 'account':
            data = searchQuery
              ? await yamsApi.searchAccounts(searchQuery)
              : await yamsApi.listAccounts();
            break;
        }
        // Filter out already added entities
        setResults(data.filter(arn => !existingArns.includes(arn)).slice(0, 50));
      } catch (err) {
        console.error('Failed to fetch results:', err);
      } finally {
        setLoading(false);
      }
    }

    const timer = setTimeout(fetchResults, 200);
    return () => clearTimeout(timer);
  }, [opened, searchQuery, entityType, existingArns]);

  const toggleArn = (arn: string): void => {
    setSelectedArns((prev) => {
      const next = new Set(prev);
      if (next.has(arn)) {
        next.delete(arn);
      } else {
        next.add(arn);
      }
      return next;
    });
  };

  const handleAdd = (): void => {
    if (selectedArns.size > 0) {
      onAdd(Array.from(selectedArns));
      onClose();
    }
  };

  // Get account info for an ARN
  const getAccountInfo = (arn: string): { accountId: string | null; accountName: string | null; service: string | null } => {
    const accountId = extractAccountId(arn) || resourceAccounts[arn] || null;
    const accountName = accountId ? accountNames[accountId] || null : null;
    const service = extractService(arn);
    return { accountId, accountName, service };
  };

  return (
    <Modal opened={opened} onClose={onClose} title={`Add ${entityType}s`} size="lg">
      <Stack gap="md">
        <TextInput
          placeholder={`Search ${entityType}s...`}
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.currentTarget.value)}
          leftSection={<IconSearch size={16} />}
          data-autofocus
        />
        <ScrollArea h={350}>
          {loading ? (
            <Text c="dimmed" ta="center" py="xl">Loading...</Text>
          ) : results.length === 0 ? (
            <Text c="dimmed" ta="center" py="xl">
              {searchQuery ? 'No results found' : `Type to search ${entityType}s`}
            </Text>
          ) : (
            <Stack gap={4}>
              {results.map((arn) => {
                const isSelected = selectedArns.has(arn);
                const { accountId, accountName, service } = getAccountInfo(arn);
                return (
                  <Group
                    key={arn}
                    gap="sm"
                    p="sm"
                    wrap="nowrap"
                    onClick={() => toggleArn(arn)}
                    style={{
                      cursor: 'pointer',
                      borderRadius: 'var(--mantine-radius-sm)',
                      backgroundColor: isSelected ? 'var(--mantine-color-violet-0)' : undefined,
                      border: isSelected ? '1px solid var(--mantine-color-violet-4)' : '1px solid transparent',
                    }}
                  >
                    <Checkbox
                      checked={isSelected}
                      onChange={() => toggleArn(arn)}
                      onClick={(e) => e.stopPropagation()}
                    />
                    <Box style={{ flex: 1, minWidth: 0 }}>
                      <Text size="sm" fw={500} truncate>
                        {formatArnLabel(arn)}
                      </Text>
                      <Group gap="xs">
                        {accountName && (
                          <Text size="xs" c="dimmed">
                            {accountName} ({accountId})
                          </Text>
                        )}
                        {!accountName && accountId && (
                          <Text size="xs" c="dimmed">{accountId}</Text>
                        )}
                        {service && (
                          <>
                            {(accountName || accountId) && <Text size="xs" c="dimmed">·</Text>}
                            <Text size="xs" c="dimmed">{service}</Text>
                          </>
                        )}
                      </Group>
                      <Text size="xs" c="dimmed" ff="monospace" truncate>
                        {arn}
                      </Text>
                    </Box>
                  </Group>
                );
              })}
            </Stack>
          )}
        </ScrollArea>
        <Group justify="space-between">
          <Text size="sm" c="dimmed">
            {selectedArns.size} selected
          </Text>
          <Group gap="sm">
            <Button variant="default" onClick={onClose}>Cancel</Button>
            <Button onClick={handleAdd} disabled={selectedArns.size === 0}>
              Add {selectedArns.size > 0 ? `(${selectedArns.size})` : ''}
            </Button>
          </Group>
        </Group>
      </Stack>
    </Modal>
  );
}

// Overlay selector - search existing or create new
interface OverlaySelectorProps {
  overlays: OverlaySummary[];
  currentOverlayId?: string;
  selectedOverlayId: string | null;
  onSelectOverlay: (overlay: OverlaySummary | null) => void;
  confirmedNewName: string;
  onConfirmNewName: (name: string) => void;
  onClearNewName: () => void;
  error?: string;
}

function OverlaySelector({
  overlays,
  currentOverlayId,
  selectedOverlayId,
  onSelectOverlay,
  confirmedNewName,
  onConfirmNewName,
  onClearNewName,
  error,
}: OverlaySelectorProps): JSX.Element {
  const combobox = useCombobox({
    onDropdownClose: () => combobox.resetSelectedOption(),
  });
  const [search, setSearch] = useState('');
  const [debouncedSearch] = useDebouncedValue(search, 150);
  const [newNameInput, setNewNameInput] = useState('');

  const selectedOverlay = selectedOverlayId ? overlays.find((o) => o.id === selectedOverlayId) : null;

  const filteredOverlays = overlays.filter(
    (o) => o.id !== currentOverlayId && o.name.toLowerCase().includes(debouncedSearch.toLowerCase())
  );

  const handleSelect = (overlayId: string): void => {
    const overlay = overlays.find((o) => o.id === overlayId);
    if (overlay) {
      onSelectOverlay(overlay);
      onClearNewName();
      setNewNameInput('');
    }
    combobox.closeDropdown();
  };

  const handleClearExisting = (e: React.MouseEvent): void => {
    e.stopPropagation();
    onSelectOverlay(null);
    setSearch('');
  };

  const handleNewNameKeyDown = (e: React.KeyboardEvent<HTMLInputElement>): void => {
    if (e.key === 'Enter' && newNameInput.trim()) {
      onConfirmNewName(newNameInput.trim());
      onSelectOverlay(null);
      setSearch('');
    }
  };

  const handleClearNewName = (e: React.MouseEvent): void => {
    e.stopPropagation();
    onClearNewName();
    setNewNameInput('');
  };

  return (
    <Group gap="md" align="flex-start" wrap="nowrap">
      {/* Search existing overlays */}
      <Box style={{ flex: 1 }}>
        <Combobox
          store={combobox}
          onOptionSubmit={(val) => handleSelect(val)}
          styles={{
            dropdown: {
              border: '2px solid var(--mantine-color-gray-4)',
              boxShadow: 'var(--mantine-shadow-md)',
            },
            option: {
              '&[data-combobox-selected]': {
                backgroundColor: 'var(--mantine-color-violet-light)',
              },
              '&:hover': {
                backgroundColor: 'var(--mantine-color-violet-light)',
              },
            },
          }}
        >
          <Combobox.Target>
            <Input.Wrapper label="Select Existing Overlay">
              {selectedOverlay && !combobox.dropdownOpened ? (
                <Group
                  gap="xs"
                  justify="space-between"
                  wrap="nowrap"
                  onClick={() => combobox.openDropdown()}
                  style={{
                    border: '1px solid var(--mantine-color-gray-4)',
                    borderRadius: 'var(--mantine-radius-sm)',
                    padding: '8px 12px',
                    minHeight: '36px',
                    cursor: 'pointer',
                  }}
                >
                  <Box style={{ overflow: 'hidden', flex: 1 }}>
                    <Group gap="xs" justify="space-between">
                      <Text size="sm" fw={500} truncate>{selectedOverlay.name}</Text>
                      <Text size="xs" c="dimmed">{formatRelativeTime(selectedOverlay.createdAt)}</Text>
                    </Group>
                    <Text size="xs" c="dimmed" ff="monospace" truncate>
                      {selectedOverlay.id}
                    </Text>
                    <Text size="xs" c="dimmed">
                      {formatEntityCounts(selectedOverlay)}
                    </Text>
                  </Box>
                  <ActionIcon
                    size="sm"
                    variant="subtle"
                    color="gray"
                    onClick={handleClearExisting}
                    aria-label="Clear selection"
                  >
                    <IconX size={14} />
                  </ActionIcon>
                </Group>
              ) : (
                <InputBase
                  value={search}
                  onChange={(e) => {
                    setSearch(e.currentTarget.value);
                    combobox.openDropdown();
                    combobox.updateSelectedOptionIndex();
                  }}
                  onClick={() => combobox.openDropdown()}
                  onFocus={() => combobox.openDropdown()}
                  onBlur={() => {
                    setTimeout(() => combobox.closeDropdown(), 150);
                  }}
                  placeholder="Search existing overlays..."
                  rightSection={<IconSearch size={16} color="var(--mantine-color-dimmed)" />}
                  rightSectionPointerEvents="none"
                  disabled={!!confirmedNewName}
                />
              )}
            </Input.Wrapper>
          </Combobox.Target>

          <Combobox.Dropdown>
            <Combobox.Options>
              <ScrollArea.Autosize mah={200} type="scroll">
                {filteredOverlays.length === 0 ? (
                  <Combobox.Empty>
                    {debouncedSearch ? 'No matching overlays found' : 'Type to search overlays'}
                  </Combobox.Empty>
                ) : (
                  filteredOverlays.slice(0, 10).map((ov) => (
                    <Combobox.Option value={ov.id} key={ov.id}>
                      <Box>
                        <Group gap="xs" justify="space-between">
                          <Text size="sm" fw={500}>{ov.name}</Text>
                          <Text size="xs" c="dimmed">{formatRelativeTime(ov.createdAt)}</Text>
                        </Group>
                        <Text size="xs" c="dimmed" ff="monospace" truncate>
                          {ov.id}
                        </Text>
                        <Text size="xs" c="dimmed">
                          {formatEntityCounts(ov)}
                        </Text>
                      </Box>
                    </Combobox.Option>
                  ))
                )}
              </ScrollArea.Autosize>
            </Combobox.Options>
          </Combobox.Dropdown>
        </Combobox>
      </Box>

      {/* OR divider - vertical */}
      <Stack gap={0} align="center" justify="center" style={{ alignSelf: 'stretch', paddingTop: 24 }}>
        <Box style={{ width: 1, flex: 1, backgroundColor: 'var(--mantine-color-gray-3)' }} />
        <Text size="xs" c="dimmed" fw={500} py="xs">OR</Text>
        <Box style={{ width: 1, flex: 1, backgroundColor: 'var(--mantine-color-gray-3)' }} />
      </Stack>

      {/* Create new overlay */}
      <Box style={{ flex: 1 }}>
        <Input.Wrapper label="Create New Overlay" error={error}>
          {confirmedNewName ? (
            <Group
              gap="xs"
              justify="space-between"
              wrap="nowrap"
              style={{
                border: '1px solid var(--mantine-color-gray-4)',
                borderRadius: 'var(--mantine-radius-sm)',
                padding: '8px 12px',
                minHeight: '36px',
              }}
            >
              <Box style={{ overflow: 'hidden', flex: 1 }}>
                <Text size="sm" fw={500} truncate>{confirmedNewName}</Text>
                <Text size="xs" c="dimmed">New overlay (unsaved)</Text>
              </Box>
              <ActionIcon
                size="sm"
                variant="subtle"
                color="gray"
                onClick={handleClearNewName}
                aria-label="Clear new overlay name"
              >
                <IconX size={14} />
              </ActionIcon>
            </Group>
          ) : (
            <TextInput
              value={newNameInput}
              onChange={(e) => setNewNameInput(e.currentTarget.value)}
              onKeyDown={handleNewNameKeyDown}
              placeholder="Type a name and press Enter"
              disabled={!!selectedOverlayId}
            />
          )}
        </Input.Wrapper>
      </Box>
    </Group>
  );
}

export function OverlayEditorPage(): JSX.Element {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const isNew = id === 'new';

  const [loading, setLoading] = useState(!isNew);
  const [saving, setSaving] = useState(false);
  const [overlay, setOverlay] = useState<OverlayData | null>(
    isNew
      ? { name: '', id: '', createdAt: '', principals: [], resources: [], policies: [], accounts: [], groups: [] }
      : null
  );
  const [error, setError] = useState<string | null>(null);
  const [hasChanges, setHasChanges] = useState(false);
  const [existingOverlays, setExistingOverlays] = useState<OverlaySummary[]>([]);
  const [accountNames, setAccountNames] = useState<Record<string, string>>({});
  const [selectedOverlayId, setSelectedOverlayId] = useState<string | null>(null);
  const [confirmedNewName, setConfirmedNewName] = useState<string>('');

  // Selected entity state
  const [selectedEntityType, setSelectedEntityType] = useState<EntityType | null>(null);
  const [selectedEntityIndex, setSelectedEntityIndex] = useState<number>(-1);
  const [editingEntity, setEditingEntity] = useState<EntityUnion | null>(null);
  const [entityDirty, setEntityDirty] = useState(false);

  // Add modal state
  const [addModalOpened, { open: openAddModal, close: closeAddModal }] = useDisclosure(false);
  const [addEntityType, setAddEntityType] = useState<EntityType>('principal');

  // Fetch existing overlays and account names
  useEffect(() => {
    yamsApi.listOverlays()
      .then((list) => {
        // Sort by createdAt descending (newest first)
        list.sort((a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime());
        setExistingOverlays(list);
      })
      .catch((err) => console.error('Failed to fetch overlays:', err));
    yamsApi.accountNames()
      .then(setAccountNames)
      .catch((err) => console.error('Failed to fetch account names:', err));
  }, []);

  const fetchOverlay = useCallback(async (): Promise<void> => {
    if (!id || isNew) return;
    setLoading(true);
    try {
      const data = await yamsApi.getOverlay(id);
      setOverlay(data);
      setError(null);
    } catch (err) {
      console.error('Failed to fetch overlay:', err);
      setError(err instanceof Error ? err.message : 'Failed to fetch overlay');
    } finally {
      setLoading(false);
    }
  }, [id, isNew]);

  useEffect(() => {
    fetchOverlay();
  }, [fetchOverlay]);

  // Get entity list for a type
  const getEntityList = (type: EntityType): EntityUnion[] => {
    if (!overlay) return [];
    switch (type) {
      case 'principal': return overlay.principals || [];
      case 'resource': return overlay.resources || [];
      case 'policy': return overlay.policies || [];
      case 'account': return overlay.accounts || [];
    }
  };

  // Get entity ID
  const getEntityId = (entity: EntityUnion, type: EntityType): string => {
    if (type === 'account') {
      return (entity as Account).Id;
    }
    return (entity as Principal | Resource | Policy).Arn;
  };

  // Select an entity for editing
  const handleSelectEntity = (type: EntityType, index: number): void => {
    const list = getEntityList(type);
    if (index >= 0 && index < list.length) {
      // Deep clone to avoid direct mutation
      setEditingEntity(JSON.parse(JSON.stringify(list[index])));
      setSelectedEntityType(type);
      setSelectedEntityIndex(index);
      setEntityDirty(false);
    }
  };

  // Update the editing entity
  const handleEntityChange = (updated: EntityUnion): void => {
    setEditingEntity(updated);
    setEntityDirty(true);
  };

  // Save entity changes back to overlay
  const handleSaveEntity = (): void => {
    if (!overlay || !editingEntity || selectedEntityType === null || selectedEntityIndex < 0) return;

    const updatedOverlay = { ...overlay };
    switch (selectedEntityType) {
      case 'principal':
        updatedOverlay.principals = (overlay.principals || []).map((p, i) =>
          i === selectedEntityIndex ? (editingEntity as Principal) : p
        );
        break;
      case 'resource':
        updatedOverlay.resources = (overlay.resources || []).map((r, i) =>
          i === selectedEntityIndex ? (editingEntity as Resource) : r
        );
        break;
      case 'policy':
        updatedOverlay.policies = (overlay.policies || []).map((p, i) =>
          i === selectedEntityIndex ? (editingEntity as Policy) : p
        );
        break;
      case 'account':
        updatedOverlay.accounts = (overlay.accounts || []).map((a, i) =>
          i === selectedEntityIndex ? (editingEntity as Account) : a
        );
        break;
    }

    setOverlay(updatedOverlay);
    setHasChanges(true);
    setEntityDirty(false);
  };

  // Open add modal
  const handleOpenAddModal = (type: EntityType): void => {
    setAddEntityType(type);
    openAddModal();
  };

  // Add entities from search
  const handleAddEntities = async (arns: string[]): Promise<void> => {
    if (!overlay || arns.length === 0) return;

    try {
      switch (addEntityType) {
        case 'principal': {
          const principals = await Promise.all(arns.map((arn) => yamsApi.getPrincipal(arn)));
          setOverlay({
            ...overlay,
            principals: [...(overlay.principals || []), ...principals],
          });
          break;
        }
        case 'resource': {
          const resources = await Promise.all(arns.map((arn) => yamsApi.getResource(arn)));
          setOverlay({
            ...overlay,
            resources: [...(overlay.resources || []), ...resources],
          });
          break;
        }
        case 'policy': {
          const policies = await Promise.all(arns.map((arn) => yamsApi.getPolicy(arn)));
          setOverlay({
            ...overlay,
            policies: [...(overlay.policies || []), ...policies],
          });
          break;
        }
        case 'account': {
          const accounts = await Promise.all(arns.map((arn) => yamsApi.getAccount(arn)));
          setOverlay({
            ...overlay,
            accounts: [...(overlay.accounts || []), ...accounts],
          });
          break;
        }
      }
      setHasChanges(true);
    } catch (err) {
      console.error('Failed to add entities:', err);
    }
  };

  // Remove entity handlers
  const handleRemoveEntity = (type: EntityType, id: string): void => {
    if (!overlay) return;

    // Clear selection if removing the selected entity
    if (selectedEntityType === type) {
      const list = getEntityList(type);
      const removingIndex = list.findIndex((e) => getEntityId(e, type) === id);
      if (removingIndex === selectedEntityIndex) {
        setSelectedEntityType(null);
        setSelectedEntityIndex(-1);
        setEditingEntity(null);
        setEntityDirty(false);
      } else if (removingIndex < selectedEntityIndex) {
        setSelectedEntityIndex(selectedEntityIndex - 1);
      }
    }

    switch (type) {
      case 'principal':
        setOverlay({
          ...overlay,
          principals: overlay.principals?.filter(p => p.Arn !== id),
        });
        break;
      case 'resource':
        setOverlay({
          ...overlay,
          resources: overlay.resources?.filter(r => r.Arn !== id),
        });
        break;
      case 'policy':
        setOverlay({
          ...overlay,
          policies: overlay.policies?.filter(p => p.Arn !== id),
        });
        break;
      case 'account':
        setOverlay({
          ...overlay,
          accounts: overlay.accounts?.filter(a => a.Id !== id),
        });
        break;
    }
    setHasChanges(true);
  };

  // Save overlay
  const handleSave = async (): Promise<void> => {
    if (!overlay) return;

    // Determine if we're creating new or updating existing
    const isCreatingNew = isNew && !selectedOverlayId;
    const overlayIdToUpdate = selectedOverlayId || id;
    const nameToUse = isCreatingNew ? confirmedNewName : overlay.name;

    if (!nameToUse.trim()) {
      setError('Overlay name is required');
      return;
    }
    setSaving(true);
    setError(null);
    try {
      if (isCreatingNew) {
        const created = await yamsApi.createOverlay({
          name: confirmedNewName,
          principals: overlay.principals,
          resources: overlay.resources,
          policies: overlay.policies,
          accounts: overlay.accounts,
          groups: overlay.groups,
        });
        navigate(`/overlays/${created.id}/edit`, { replace: true });
      } else {
        await yamsApi.updateOverlay(overlayIdToUpdate!, {
          name: overlay.name,
          principals: overlay.principals,
          resources: overlay.resources,
          policies: overlay.policies,
          accounts: overlay.accounts,
          groups: overlay.groups,
        });
      }
      setHasChanges(false);
    } catch (err) {
      console.error('Failed to save overlay:', err);
      setError(err instanceof Error ? err.message : 'Failed to save overlay');
    } finally {
      setSaving(false);
    }
  };

  // Determine breadcrumb title based on current state
  const getBreadcrumbTitle = (): string => {
    if (selectedOverlayId && overlay) {
      return overlay.name;
    }
    if (confirmedNewName) {
      return confirmedNewName;
    }
    if (!isNew && overlay) {
      return overlay.name;
    }
    return 'New Overlay';
  };

  const breadcrumbItems = (isNew && !selectedOverlayId
    ? [
        { title: 'Overlays', href: '/overlays' },
        { title: getBreadcrumbTitle(), href: '' },
      ]
    : [
        { title: 'Overlays', href: '/overlays' },
        { title: getBreadcrumbTitle() || 'Loading...', href: selectedOverlayId ? `/overlays/${selectedOverlayId}` : `/overlays/${id}` },
        { title: 'Edit', href: '' },
      ]
  ).map((item, index, arr) => {
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

  if (error && !overlay) {
    return (
      <Box p="md">
        <EmptyState variant="error" message={error} />
      </Box>
    );
  }

  if (loading) {
    return (
      <Box p="md">
        <DetailSkeleton />
      </Box>
    );
  }

  if (!overlay) {
    return (
      <Box p="md">
        <EmptyState variant="error" message="Overlay not found" />
      </Box>
    );
  }

  const existingArns = [
    ...(overlay.principals?.map(p => p.Arn) || []),
    ...(overlay.resources?.map(r => r.Arn) || []),
    ...(overlay.policies?.map(p => p.Arn) || []),
    ...(overlay.accounts?.map(a => a.Id) || []),
    ...(overlay.groups?.map(g => g.Arn) || []),
  ];

  // Render entity list for a tab
  const renderEntityList = (type: EntityType): JSX.Element => {
    const list = getEntityList(type);

    return (
      <Stack gap="md" p="md" h="100%">
        <Button
          variant="light"
          leftSection={<IconPlus size={16} />}
          onClick={() => handleOpenAddModal(type)}
        >
          Add {type.charAt(0).toUpperCase() + type.slice(1)}
        </Button>
        <ScrollArea style={{ flex: 1 }}>
          {list.length === 0 ? (
            <Text c="dimmed" ta="center" py="xl">
              No {type}s in overlay
            </Text>
          ) : (
            <Stack gap="xs">
              {list.map((entity, idx) => (
                <EntityListItem
                  key={getEntityId(entity, type)}
                  entity={entity}
                  type={type}
                  isSelected={selectedEntityType === type && selectedEntityIndex === idx}
                  onClick={() => handleSelectEntity(type, idx)}
                  onDelete={() => handleRemoveEntity(type, getEntityId(entity, type))}
                  accountNames={accountNames}
                />
              ))}
            </Stack>
          )}
        </ScrollArea>
      </Stack>
    );
  };

  return (
    <Box p="md" h="100%">
      <Stack gap="md" h="100%">
        {/* Header row */}
        <Box>
          <Breadcrumbs separator={<IconChevronRight size={14} />} mb="md">{breadcrumbItems}</Breadcrumbs>

          {/* Error alert */}
          {error && (
            <Alert
              icon={<IconAlertCircle size={16} />}
              color="red"
              mb="md"
              withCloseButton
              onClose={() => setError(null)}
            >
              {error}
            </Alert>
          )}

          {/* Overlay selector row */}
          <Card withBorder p="md">
            <Stack gap="md">
              <OverlaySelector
                overlays={existingOverlays}
                currentOverlayId={isNew ? undefined : id}
                selectedOverlayId={selectedOverlayId}
                onSelectOverlay={async (selected) => {
                  if (selected) {
                    // Load the selected overlay's data without navigating
                    // Don't set loading state to avoid full page flicker
                    setSelectedOverlayId(selected.id);
                    setConfirmedNewName('');
                    // Clear entity selection
                    setSelectedEntityType(null);
                    setSelectedEntityIndex(-1);
                    setEditingEntity(null);
                    setEntityDirty(false);
                    try {
                      const data = await yamsApi.getOverlay(selected.id);
                      setOverlay(data);
                      setHasChanges(false);
                      setError(null);
                    } catch (err) {
                      console.error('Failed to load overlay:', err);
                      setError(err instanceof Error ? err.message : 'Failed to load overlay');
                    }
                  } else {
                    setSelectedOverlayId(null);
                  }
                }}
                confirmedNewName={confirmedNewName}
                onConfirmNewName={(name) => {
                  setConfirmedNewName(name);
                  setOverlay({ ...overlay, name });
                  setHasChanges(true);
                }}
                onClearNewName={() => {
                  setConfirmedNewName('');
                  setOverlay({ ...overlay, name: '' });
                  setHasChanges(false);
                }}
                error={error && !confirmedNewName && !selectedOverlayId ? 'Name is required' : undefined}
              />
              <Button
                leftSection={<IconDeviceFloppy size={16} />}
                onClick={handleSave}
                loading={saving}
                disabled={
                  isNew && !selectedOverlayId
                    ? !confirmedNewName
                    : !hasChanges
                }
                fullWidth
              >
                {isNew && !selectedOverlayId ? 'Create Overlay' : 'Save Changes'}
              </Button>
            </Stack>
          </Card>
        </Box>

        {/* Master/detail row */}
        <Grid gutter="md" style={{ flex: 1, minHeight: 0 }}>
          {/* Left column - Entity list */}
          <Grid.Col span={6} h="100%">
            <Card withBorder h="100%" style={{ overflow: 'hidden' }}>
              <Tabs defaultValue="principals" style={{ height: '100%', display: 'flex', flexDirection: 'column' }}>
                <Tabs.List>
                  <Tabs.Tab value="principals" leftSection={<IconUser size={14} />}>
                    Principals ({overlay.principals?.length || 0})
                  </Tabs.Tab>
                  <Tabs.Tab value="resources" leftSection={<IconDatabase size={14} />}>
                    Resources ({overlay.resources?.length || 0})
                  </Tabs.Tab>
                  <Tabs.Tab value="policies" leftSection={<IconShield size={14} />}>
                    Policies ({overlay.policies?.length || 0})
                  </Tabs.Tab>
                  <Tabs.Tab value="accounts" leftSection={<IconBuilding size={14} />}>
                    Accounts ({overlay.accounts?.length || 0})
                  </Tabs.Tab>
                </Tabs.List>

                <Tabs.Panel value="principals" style={{ flex: 1, overflow: 'hidden' }}>
                  {renderEntityList('principal')}
                </Tabs.Panel>

                <Tabs.Panel value="resources" style={{ flex: 1, overflow: 'hidden' }}>
                  {renderEntityList('resource')}
                </Tabs.Panel>

                <Tabs.Panel value="policies" style={{ flex: 1, overflow: 'hidden' }}>
                  {renderEntityList('policy')}
                </Tabs.Panel>

                <Tabs.Panel value="accounts" style={{ flex: 1, overflow: 'hidden' }}>
                  {renderEntityList('account')}
                </Tabs.Panel>
              </Tabs>
            </Card>
          </Grid.Col>

          {/* Right column - Entity detail/edit panel */}
          <Grid.Col span={6} h="100%">
            <Card withBorder h="100%" p="md">
              <EntityDetailPanel
                entity={editingEntity}
                type={selectedEntityType}
                onChange={handleEntityChange}
                onSave={handleSaveEntity}
                isDirty={entityDirty}
                overlayPolicies={overlay?.policies?.map(p => p.Arn) || []}
              />
            </Card>
          </Grid.Col>
        </Grid>
      </Stack>

      <AddEntityModal
        opened={addModalOpened}
        onClose={closeAddModal}
        entityType={addEntityType}
        onAdd={handleAddEntities}
        existingArns={existingArns}
      />
    </Box>
  );
}
